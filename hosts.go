package go_hostctl

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"os"
	"regexp"
	"strings"
)

const (
	RegexPatternName       = "^[a-zA-Z0-9\\.-_]*$"
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
	Comment   string
	IPAddress net.IP
	Hostname  string
	Aliases   []string
}

func (he *HostEntry) Validate() error {

	if he.Aliases == nil {
		he.Aliases = make([]string, 0)
	}

	// Just a comment line
	if !IsValidIP(he.IPAddress) && !IsValidName(he.Hostname) && !IsValidAliases(he.Aliases) && IsComment(Normalize(&he.Comment)) {
		he.rawLine = []byte(he.Comment)
		he.isComment = true
		return nil
	}

	// Non comment line should have valid IP
	if !IsValidIP(he.IPAddress) {
		return fmt.Errorf("no valid ip address parsed: %v", he.IPAddress)
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
		he.rawLine = []byte(fmt.Sprintf("%s\t%s\t%s\t%s\r\n", he.IPAddress.String(), he.Hostname, strings.Join(aliases, " "), he.Comment))
	} else if IsComment(he.Comment) {
		he.rawLine = []byte(fmt.Sprintf("%s\t%s\t%s\r\n", he.IPAddress.String(), he.Hostname, he.Comment))
	} else if len(he.Aliases) > 0 {
		he.rawLine = []byte(fmt.Sprintf("%s\t%s\t%s\r\n", he.IPAddress.String(), he.Hostname, strings.Join(aliases, " ")))
	} else {
		he.rawLine = []byte(fmt.Sprintf("%s\t%s\r\n", he.IPAddress.String(), he.Hostname))
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
			hostEntry.isComment = true
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

	entry := &HostEntry{
		Comment:   comment,
		IPAddress: net.ParseIP(ipaddr),
		Hostname:  hostname,
		Aliases:   aliases,
	}

	return entry, entry.Validate()
}

type HostsFileCtl struct {
	HostsFile string
	Entries   []HostEntry
}

func NewHostFileCtl(hostFilePath string) (*HostsFileCtl, error) {

	f, err := os.Open(hostFilePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	rdr := bufio.NewReader(f)

	htctl := HostsFileCtl{
		HostsFile: hostFilePath,
		Entries:   make([]HostEntry, 0),
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
		htctl.Entries = append(htctl.Entries, *entry)
	}

	return &htctl, nil
}

func (hfc *HostsFileCtl) Write(writer io.Writer) error {

	if hfc.Entries == nil || len(hfc.Entries) <= 0 {
		return nil
	}

	for n, entry := range hfc.Entries {
		if n != 0 && hfc.Entries[n-1].isComment {
			writer.Write([]byte(CarriageReturnLineFeed))
		}

		if _, err := entry.Write(writer); err != nil {
			return err
		}
	}

	return nil
}
