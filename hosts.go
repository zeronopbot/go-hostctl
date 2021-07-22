package go_hostctl

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"os"
	"strings"
)

func Normalize(item *string) string {
	if item == nil || len(*item) == 0 {
		return ""
	}

	*item = strings.Trim(strings.Trim(strings.TrimSpace(*item), "\t"), "\n")
	return *item
}

func IsComment(item string) bool {
	return strings.HasPrefix(item, "#")
}

type HostEntry struct {
	rawLine []byte
	Comment string
	IPAddress net.IP
	DomainName string
	Alias string
}

func (he *HostEntry) Validate() error {

	// Just a comment line
	if he.IPAddress == nil && len(he.DomainName) <= 0 && len(he.Alias) <= 0 && len(he.Comment) > 0 {
		return nil
	}

	// Non comment line should have valid IP
	if he.IPAddress == nil {
		return fmt.Errorf("no valid ip address parsed")
	}

	// Non comment line should have at least a domain name
	if len(he.DomainName) <= 0 {
		return fmt.Errorf("entry must have a domain name if ip address is set")
	}

	return nil
}

func (he HostEntry) Write(writer io.Writer, prefix string) error {

	if he.IPAddress == nil {

		line := fmt.Sprintf("%s%s\r\n", prefix, he.Comment)
		if _, err := writer.Write([]byte(line)); err != nil {
			return err
		}
		return nil
	}

	line := he.IPAddress.String()
	for _, str := range []string{he.DomainName, he.Alias, he.Comment} {

		if len(str) <= 0 {
			continue
		}

		line += fmt.Sprintf("\t%s", str)
	}

	line = fmt.Sprintf("%s\r\n", line)
	if _, err := writer.Write([]byte(line)); err != nil {
		return err
	}
	return nil
}

func ParseHostEntry(line []byte) (*HostEntry, error) {

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
		rawLine:    line,
	}

	for n, token := range tokens {

		tok := token

		// Comment is last or only thing in entry
		if strings.HasPrefix(token, "#") {
			hostEntry.Comment = tok
			return hostEntry, nil
		}

		switch n {

		// ip address
		case 0:
			hostEntry.IPAddress = net.ParseIP(tok)
			if hostEntry.IPAddress == nil {
				return nil, fmt.Errorf("invalid ip address: %s", tok)
			}

		// fqdn
		case 1:
			hostEntry.DomainName = tok

		// alias
		case 2:
			hostEntry.Alias = tok

		// Can only be a comment and shouldn't reach here
		case 3:
			return nil, fmt.Errorf("expecting last entry to be a comment starting with a '#' character - got '%s'", tokens[n:])
		}
	}

	return hostEntry, hostEntry.Validate()
}

func NewHostEntry(ipaddr, host, alias, comment string) (*HostEntry, error) {

	var hostEntry HostEntry

	// IP address
	ip := net.ParseIP(ipaddr)
	if ip == nil {
		return nil, fmt.Errorf("no ip address specified")
	}

	if ip.To4() == nil || ip.To16() == nil {
		return nil, fmt.Errorf("must be valid ipv4 or ipv6 address")
	}

	// Host
	if len(Normalize(&host)) <= 0 {
		return nil, fmt.Errorf("no hostname specified")
	}

	if IsComment(host)




}

type HostsFileCtl struct {
	HostsFile string
	Entries []HostEntry
}

func NewHostFileCtl(fpath string) (HostsFileCtl, error) {
	var hostctl HostsFileCtl
	return hostctl, hostctl.Parse(fpath)
}

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

func (hfc *HostsFileCtl) Parse(hostFilePath string) error {

	hfc.HostsFile = hostFilePath
	if hfc.Entries == nil {
		hfc.Entries = make([]HostEntry, 0)
	}

	f, err := os.Open(hostFilePath)
	if err != nil {
		return err
	}
	defer f.Close()

	rdr := bufio.NewReader(f)
	var lineNumber int
	for {

		line, prefix, err := rdr.ReadLine()

		if prefix {
			return fmt.Errorf("line is too long: %d", lineNumber)
		}

		if err != nil {
			if err == io.EOF {
				return nil
			}
		}

		// Skip newlines
		if len(line) <= 0 {
			lineNumber++
			continue
		}

		entry, err := ParseHostEntry(line)
		if err != nil {
			return fmt.Errorf("invalid host entry on line %d - %s", lineNumber, err)
		}

		lineNumber++
		hfc.Entries = append(hfc.Entries, *entry)
	}
}

func (hfc *HostsFileCtl) Write(writer io.Writer) error {

	if hfc.Entries == nil || len(hfc.Entries) <= 0 {
		return nil
	}

	for n, entry := range hfc.Entries {

		var prefix string

		// Only append prefix if the previous line was a comment only
		if n != 0 {
			prevEntry := hfc.Entries[n-1]
			if len(entry.Comment) > 0 && entry.IPAddress == nil && prevEntry.IPAddress != nil {
				prefix = fmt.Sprintf("\r\n")
			}
		}

		if err := entry.Write(writer, prefix); err != nil {
			return err
		}
	}
	return nil
}