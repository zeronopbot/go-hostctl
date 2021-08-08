package main

import (
	"flag"
	"fmt"
	. "github.com/zeronopbot/go-hostctl"
	"io/ioutil"
	"log"
	"os"
)

func main() {

	fpath := flag.String("f", "testdata/etc/hosts/mixed_hosts", "Hosts file path")
	flag.Parse()

	// Open existing host file to copy its entries (for testing)
	hctlFile1, err := NewHostFileCtl(*fpath)
	if err != nil {
		log.Fatal(err)
	}

	// Remove it if exists
	os.Remove("/tmp/hosts")

	// File we want to "actually" use
	hctl, err := NewHostFileCtl("/tmp/hosts")
	if err != nil {
		log.Fatal(err)
	}

	// Show any existing entries
	for _, entry := range hctlFile1.Entries() {
		if err := hctl.Add(entry, 0); err != nil {
			log.Fatal(err)
		}
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

	// Show the host file before we sync (empty)
	out, err := ioutil.ReadFile("/tmp/hosts")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("\n---- Before Sync() ----\n%s\n-----------------------\n", string(out))

	// Write remaining contents to an io.Writer (still not in file, just in memory)
	if _, err := hctl.Write(os.Stdout); err != nil {
		log.Fatal(err)
	}

	// Sync the memory contents to disk
	if _, err := hctl.Sync(); err != nil {
		log.Fatal(err)
	}

	// Show the host file after we sync (empty)
	out, err = ioutil.ReadFile("/tmp/hosts")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("\n---- After Sync() ----\n%s\n-----------------------\n", string(out))
}
