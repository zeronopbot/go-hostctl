package main

import (
	"flag"
	"fmt"
	. "github.com/zeronopbot/go-hostctl"
	"log"
	"os"
)

func main() {

	fpath := flag.String("f", "testdata/etc/hosts/mixed_hosts", "Hosts file path")
	flag.Parse()

	// Create new host control file (or open existing)
	hctl, err := NewHostFileCtl(*fpath)
	if err != nil {
		log.Fatal(err)
	}

	// Show any existing entries
	for n, entry := range hctl.Entries() {
		fmt.Printf("Entry: %d - %s\n", n, entry.String())
	}

	// Delete all localhost ips in loop
	// Position of all entries are updated after each host operations (e.g. add, delete)
	for {
		entries, err := hctl.GetIP("127.0.0.1")
		if err != nil {
			log.Fatal(err)
		}

		if len(entries) <= 0 {
			break
		}

		if err := hctl.Delete(entries[0].Position); err != nil {
			log.Fatal(err)
		}
	}

	// Delete all host_entry_4 host names
	for {

		entries, err := hctl.GetHostname("host_entry_4")
		if err != nil {
			log.Fatal(err)
		}

		if len(entries) <= 0 {
			break
		}

		if err := hctl.Delete(entries[0].Position); err != nil {
			log.Fatal(err)
		}
	}

	// Delete some_macos alias
	entries, err := hctl.GetAlias("some_macos")
	if err != nil {
		log.Fatal(err)
	}

	if len(entries) > 0 {
		if err := hctl.Delete(entries[0].Position); err != nil {
			log.Fatal(err)
		}
	}

	// Write remaining contents to an io.Writer
	hctl.Write(os.Stdout)
}
