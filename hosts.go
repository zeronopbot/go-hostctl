package go_hostctl

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
)

type HostEntry struct {
	Comment string
	Hostname string
	IPAddress string
}

type HostsFileCtl struct {
	HostsFile string
	StrictMode bool
	Entries map[string]HostEntry
}

func (hfc *HostsFileCtl) Parse(hostFilePath string) error {

	f, err := os.Open(hostFilePath)
	if err != nil {
		return err
	}
	defer f.Close()

	rdr := bufio.NewReader(f)

	var lineNumber int
	line, prefix, err := rdr.ReadLine()
	for {

		if prefix {
			return fmt.Errorf("line is too long: %d", lineNumber)
		}

		if err != nil {
			if err == io.EOF {
				return nil
			}
		}

		log.Printf(	"%d - %s", lineNumber, line)


	}

}