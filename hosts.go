package go_hostctl

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"regexp"
	"strings"
	"sync"
)

const (
	RegexPatternName       = "^[a-zA-Z0-9\\.\\-_]*$"
	CarriageReturnLineFeed = "\r\n"
)

var (
	nameMatcher = regexp.MustCompile(RegexPatternName)
)

func tokenize(line []byte) ([]string, error) {

	rawLine := line
	tmpBuf := make([]byte, len(rawLine))
	copy(tmpBuf, rawLine)

	// Trim all whitespace left and right
	tmpBuf = bytes.TrimLeft(bytes.TrimLeft(tmpBuf, " "), "\t")
	tmpBuf = bytes.TrimRight(bytes.TrimLeft(tmpBuf, " "), "\t")

	tokens := make([]string, 0)
	for {

		i, words, err := bufio.ScanWords(tmpBuf, false)
		if err != nil {
			return nil, err
		}

		if i == 0 {
			tokens = append(tokens, fmt.Sprintf("%s", tmpBuf))
			break
		}

		if bytes.HasPrefix(words, []byte{'#'}) {
			tokens = append(tokens, fmt.Sprintf("%s", tmpBuf))
			break
		}

		tokens = append(tokens, fmt.Sprintf("%s", words))
		tmpBuf = bytes.TrimLeft(bytes.TrimLeft(tmpBuf[i:], " "), "\t")
		tmpBuf = bytes.TrimLeft(bytes.TrimLeft(tmpBuf, " "), "\t")
	}

	return tokens, nil
}

func Normalize(item *string) string {
	if item == nil || len(*item) == 0 {
		return ""
	}

	*item = strings.Trim(strings.Trim(strings.TrimSpace(*item), "\t"), "\n")
	return *item
}

func IsComment(item string) bool {
	return strings.HasPrefix(Normalize(&item), "#")
}

func IsValidName(name string) bool {
	if len(name) <= 0 {
		return false
	}
	return len(nameMatcher.FindStringSubmatch(Normalize(&name))) > 0
}

func IsValidIP(ip net.IP) bool {

	// No ip
	if ip == nil {
		return false
	}

	// Ipv4/6
	if ip.To4() != nil || ip.To16() != nil {
		return true
	}

	return false
}

func IsValidAliases(aliases []string) bool {
	if aliases == nil || len(aliases) == 0 {
		return false
	}

	for _, alias := range aliases {
		if !IsValidName(alias) {
			return false
		}
	}

	return true
}

type HostEntry struct {
	rawLine   []byte
	isComment bool
	Position  int
	Comment   string
	IPAddress net.IP
	Hostname  string
	Aliases   []string
}

func (he *HostEntry) Validate() error {

	he.isComment = false
	if he.Aliases == nil {
		he.Aliases = make([]string, 0)
	}

	if he.isComment {
		he.rawLine = []byte(he.Comment)
		return nil
	}

	if he.IPAddress != nil && IsComment(he.IPAddress.String()) {
		return fmt.Errorf("ip address cannot be a comment: %s", he.IPAddress.String())
	}

	if IsComment(he.Hostname) {
		return fmt.Errorf("hostname cannot be a comment: %s", he.Hostname)
	}

	for n, alias := range he.Aliases {
		if IsComment(alias) {
			return fmt.Errorf("alias %d cannot be a comment: %s", n+1, alias)
		}
 	}

	// Just a comment line
	if !IsValidIP(he.IPAddress) && !IsValidName(he.Hostname) && !IsValidAliases(he.Aliases) && IsComment(Normalize(&he.Comment)) {
		he.rawLine = []byte(he.Comment)
		he.isComment = true
		return nil
	}

	// Non comment line should have valid IP
	if !IsValidIP(he.IPAddress) {
		return fmt.Errorf("no valid ip address parsed: %v", he)
	}

	// Non comment line should have at least a domain name
	if !IsValidName(Normalize(&he.Hostname)) {
		return fmt.Errorf("entry must have a hostname if ip address is set")
	}

	// Cleanup and set aliases
	aliases := make([]string, len(he.Aliases))
	for n, alias := range he.Aliases {
		aliases[n] = Normalize(&alias)
	}

	he.Aliases = aliases

	// Setup the raw line based on what is provided and valid
	if IsComment(he.Comment) && len(aliases) > 0 {
		he.rawLine = []byte(fmt.Sprintf("%s\t%s\t%s\t%s", he.IPAddress.String(), he.Hostname, strings.Join(aliases, " "), he.Comment))
	} else if IsComment(he.Comment) {
		he.rawLine = []byte(fmt.Sprintf("%s\t%s\t%s", he.IPAddress.String(), he.Hostname, he.Comment))
	} else if len(he.Aliases) > 0 {
		he.rawLine = []byte(fmt.Sprintf("%s\t%s\t%s", he.IPAddress.String(), he.Hostname, strings.Join(aliases, " ")))
	} else {
		he.rawLine = []byte(fmt.Sprintf("%s\t%s", he.IPAddress.String(), he.Hostname))
	}

	return nil
}

func (he *HostEntry) Write(writer io.Writer) (int, error) {
	if err := he.Validate(); err != nil {
		return 0, err
	}

	return writer.Write(he.rawLine)
}

func ParseHostEntryLine(line []byte) (*HostEntry, error) {

	if line == nil || len(line) <= 0 {
		return nil, fmt.Errorf("invalid line, empty or nil")
	}

	tokens, err := tokenize(line)
	if err != nil {
		return nil, err
	}

	if len(tokens) == 0 {
		return nil, fmt.Errorf("no tokens parsed for line: '%s'", line)
	}

	hostEntry := &HostEntry{
		rawLine: line,
		Aliases: make([]string, 0),
	}

	for n, token := range tokens {

		tok := token

		// Everything after this is part of the comment
		if strings.HasPrefix(token, "#") {
			hostEntry.Comment = strings.Join(tokens[n:], " ")
			hostEntry.isComment = n == 0 // only a comment line IF its the first token
			return hostEntry, nil
		}

		switch n {

		// IP Address
		case 0:
			hostEntry.IPAddress = net.ParseIP(tok)
			if hostEntry.IPAddress == nil {
				return nil, fmt.Errorf("invalid ip address: %s", tok)
			}

		// Hostname
		case 1:
			if !IsValidName(tok) {
				return nil, fmt.Errorf("not a valid hostname: %s", tok)
			}
			hostEntry.Hostname = tok

		// Aliases
		default:
			if !IsValidName(tok) {
				return nil, fmt.Errorf("not a valid alias: %s", tok)
			}
			hostEntry.Aliases = append(hostEntry.Aliases, tok)
		}
	}

	return hostEntry, hostEntry.Validate()
}

func NewHostEntry(ipaddr, hostname, comment string, aliases ...string) (*HostEntry, error) {

	if len(comment) != 0 && !strings.HasPrefix(Normalize(&comment), "#") {
		comment = fmt.Sprintf("# %s", comment)
	}

	entry := &HostEntry{
		Comment:   comment,
		IPAddress: net.ParseIP(ipaddr),
		Hostname:  hostname,
		Aliases:   aliases,
	}

	return entry, entry.Validate()
}

type HostFileCtl interface {
	Delete(position int) error
	Add(entry HostEntry, position int) error
	GetIP(ip string) ([]HostEntry, error)
	GetAlias(alias string) ([]HostEntry, error)
	GetHostname(alias string) ([]HostEntry, error)
	Write(writer io.Writer) (int, error)
	Flush() (int, error)
}

type hostsFileCtl struct {
	rwLck     *sync.RWMutex
	hostsFile string
	entries   []HostEntry
}

func NewHostFileCtl(hostFilePath string) (HostFileCtl, error) {

	// Get existing file mode
	mode := os.FileMode(0644)
	if stat, err := os.Stat(hostFilePath); err == nil {
		mode = stat.Mode()
	}

	f, err := os.OpenFile(hostFilePath, os.O_CREATE | os.O_RDWR | os.O_SYNC, mode)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	rdr := bufio.NewReader(f)

	htctl := hostsFileCtl{
		rwLck:     new(sync.RWMutex),
		hostsFile: hostFilePath,
		entries:   make([]HostEntry, 0),
	}

	var lineNumber int
readLoop:
	for {

		line, prefix, err := rdr.ReadLine()
		if prefix {
			return nil, fmt.Errorf("line is too long: %d", lineNumber)
		}

		if err != nil {
			if err == io.EOF {
				break readLoop
			}
			return nil, err
		}

		// Skip newlines
		if len(line) <= 0 {
			lineNumber++
			continue
		}

		entry, err := ParseHostEntryLine(line)
		if err != nil {
			return nil, fmt.Errorf("invalid host entry on line %d - %s", lineNumber, err)
		}

		lineNumber++
		entry.Position = len(htctl.entries)
		htctl.entries = append(htctl.entries, *entry)
	}

	return &htctl, nil
}

func (hfc *hostsFileCtl) updatePosition() {
	for n, _ := range hfc.entries {
		hfc.entries[n].Position = n
	}
}

func (hfc *hostsFileCtl) Delete(position int) error {

	if position < -1 {
		return fmt.Errorf("invalid position: %d", position)
	}

	hfc.rwLck.Lock()
	defer hfc.rwLck.Unlock()

	if position >= len(hfc.entries) {
		return fmt.Errorf("postion out of range: %d", position)
	}

	if len(hfc.entries) == 0 {
		return nil
	}

	defer hfc.updatePosition()

	switch position {
	case 0:
		if len(hfc.entries) > 1 {
			hfc.entries = hfc.entries[1:]
			break
		}
		hfc.entries = make([]HostEntry, 0)

	case -1:
		if len(hfc.entries) > 1 {
			hfc.entries = hfc.entries[:len(hfc.entries)-1]
			break
		}
		hfc.entries = make([]HostEntry, 0)

	default:
		hfc.entries = append(hfc.entries[:position], hfc.entries[position+1:] ...)
	}

	return nil

}

func (hfc *hostsFileCtl) Add(entry HostEntry, position int) error {

	if err := entry.Validate(); err != nil {
		return err
	}

	if position < -1 {
		return fmt.Errorf("invalid position: %d", position)
	}

	hfc.rwLck.Lock()
	defer hfc.rwLck.Unlock()

	if position == len(hfc.entries) {
		position = -1
	}

	if position > len(hfc.entries) {
		return fmt.Errorf("postion out of range: %d", position)
	}

	defer hfc.updatePosition()

	switch position {
	case 0:
		hfc.entries = append([]HostEntry{entry}, hfc.entries ...)
	case -1:
		hfc.entries = append(hfc.entries, entry)
	default:
		hfc.entries = append(hfc.entries[:position], append([]HostEntry{entry}, hfc.entries[position:] ...) ...)
	}

	return nil
}

func (hfc *hostsFileCtl) GetIP(ip string) ([]HostEntry, error) {
	if len(hfc.entries) == 0 {
		return nil, fmt.Errorf("no entries in file")
	}

	ipaddr := net.ParseIP(ip)
	if ipaddr == nil {
		return nil, fmt.Errorf("invalid ip address specified: %s", ip)
	}

	hfc.rwLck.RLock()
	defer hfc.rwLck.RUnlock()

	entries := make([]HostEntry, 0)
	for _, entry := range hfc.entries {
		tmpEntry := entry
		if strings.Compare(ip, entry.IPAddress.String()) == 0 {
			entries = append(entries, tmpEntry)
			break
		}
	}

	return entries, nil
}

func (hfc *hostsFileCtl) GetAlias(alias string) ([]HostEntry, error) {

	if len(hfc.entries) == 0 {
		return nil, fmt.Errorf("no entries in file")
	}

	hfc.rwLck.RLock()
	defer hfc.rwLck.RUnlock()

	entries := make([]HostEntry, 0)
	for _, entry := range hfc.entries {

		if entry.Aliases == nil || len(entry.Aliases) == 0 {
			continue
		}

		tmpEntry := entry
		for _, a := range entry.Aliases {
			if strings.Compare(alias, a) == 0 {
				entries = append(entries, tmpEntry)
				break
			}
		}
	}

	return entries, nil
}

func (hfc *hostsFileCtl) GetHostname(hostname string) ([]HostEntry, error) {

	if len(hfc.entries) == 0 {
		return nil, fmt.Errorf("no entries in file")
	}

	hfc.rwLck.RLock()
	defer hfc.rwLck.RUnlock()

	entries := make([]HostEntry, 0)
	for _, entry := range hfc.entries {
		tmpEntry := entry
		if strings.Compare(hostname, entry.Hostname) == 0 {
			entries = append(entries, tmpEntry)
			break
		}
	}

	return entries, nil
}

// Write will write all the entries to the write specified
func (hfc *hostsFileCtl) Write(writer io.Writer) (int, error) {

	if hfc.entries == nil || len(hfc.entries) <= 0 {
		return 0, nil
	}

	count := 0
	for n, entry := range hfc.entries {

		c, err := writer.Write([]byte(CarriageReturnLineFeed))
		if err != nil {
			return 0, err
		}
		count += c

		if n != 0 && !hfc.entries[n-1].isComment && entry.isComment {
			c, err := writer.Write([]byte(CarriageReturnLineFeed))
			if err != nil {
				return 0, nil
			}
			count += c
		}

		c, err = entry.Write(writer)
		if err != nil {
			return 0, err
		}
		count += c
	}

	// New Line
	if count > 0 {
		c, err := writer.Write([]byte(CarriageReturnLineFeed))
		if err != nil {
			return 0, nil
		}
		count += c
	}

	return count, nil
}

// Flush entries to the existing file
// Reverts on any failure back to the original file contents
func (hfc hostsFileCtl) Flush() (int, error) {

	s, err := os.Stat(hfc.hostsFile)
	if err != nil {
		return 0, err
	}

	// Read in the existing contents to revert in cause of failure
	contents, err := ioutil.ReadFile(hfc.hostsFile)
	if err != nil {
		return 0, fmt.Errorf("failed to read existing file to make backup: %s", err)
	}

	defer func(err error) {
		if err != nil {
			ioutil.WriteFile(hfc.hostsFile, contents, s.Mode())
		}
	}(err)

	var f *os.File
	f, err = os.OpenFile(hfc.hostsFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, s.Mode())
	if err != nil {
		return 0, err
	}
	defer f.Close()

	return hfc.Write(f)
}
