package go_hostctl

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
)

type HostEntry struct {
	rawLine []byte
	Comment string
	IPAddress string
	Hostname string
	Alias string
}

type HostsFileCtl struct {
	HostsFile string
}

func parseLine(line []byte) ([]string, error) {

	log.Printf("--- %s ---", line)
	rawLine := line
	tmpBuf := make([]byte, len(rawLine))
	copy(tmpBuf, rawLine)

	// Trim all whitespace left and right
	tmpBuf = bytes.TrimLeft(bytes.TrimLeft(tmpBuf, " "), "\t")
	tmpBuf = bytes.TrimRight(bytes.TrimLeft(tmpBuf, " "), "\t")

	tokens := make([]string, 0)
	var entry int
	for {

		log.Printf("Checking line: %s - %t", tmpBuf, false)
		i, words, err := bufio.ScanWords(tmpBuf, false)
		log.Printf("HERE %d: %d - '%s' - %v", entry, i, words, err)

		if i == 0 {
			tokens = append(tokens, fmt.Sprintf("%s", tmpBuf))
			break
		}

		if bytes.HasPrefix(words, []byte{'#'}) {
			tokens = append(tokens, fmt.Sprintf("%s", tmpBuf))
			break
		}

		entry++
		tokens = append(tokens, fmt.Sprintf("%s", words))
		tmpBuf = bytes.TrimLeft(bytes.TrimLeft(tmpBuf[i:], " "), "\t")

		if entry > 10 {
			return nil, fmt.Errorf("tokens exceeded")
		}
	}

	return tokens, nil
}

func (hfc *HostsFileCtl) Parse(hostFilePath string) error {

	f, err := os.Open(hostFilePath)
	if err != nil {
		return err
	}
	defer f.Close()

	hfc.HostsFile = hostFilePath
	rdr := bufio.NewReader(f)

	var lineNumber int
	for {

		line, prefix, err := rdr.ReadLine()
		log.Printf(	"%d - %s", lineNumber, line)

		if prefix {
			return fmt.Errorf("line is too long: %d", lineNumber)
		}

		if err != nil {
			if err == io.EOF {
				return nil
			}
		}

		if len(line) <= 0 {
			lineNumber++
			continue
		}

		tokens, err := parseLine(line)
		if err != nil {
			return err
		}

		log.Printf("%d - %s", len(tokens), tokens)
		lineNumber++

		if lineNumber > 20 {
			return nil
		}
	}
}