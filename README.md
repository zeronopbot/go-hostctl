# go-hostctl
Cross platform hosts file control library

# Purpose
The intent of this project is to provide a simple, cross-platform, easy to use library to manage and neatly format the 
contents of the hosts file. It makes no attempt or assumption to ensure entries don't overlap (as allowed by the system)
but will ensure the entries are all valid, organised and formatted correctly regardless of the source content.

# Usage
This library can be used to manage, update, delete, search, copy or create host file entries in the host file as needed. 
You'll need escalated permissions to modify the file while reading it and modding entries without a Sync() only requires
read permissions.  

The intended workflow is as follows;
1) Create a new HostFileCtl interface by parsing any given file/reader
2) Modify the entries using the HostFileCtl interface class methods
    a) This will only modify them in-memory, not on the backing disk
3) Sync() the HostFileCtl interface to write the new entry state back to the file

## Example
Below is an example (from cmd/main.go) that shows how you can use this library to parse entries from at hosts file, 
search by IP, hostname, alias and even organise the contents. It will ensure all entries are properly formatted in their
respective positions in the file while also ensuring the file neatly presents the content for maximum readability when
written back to the backing file. The library also exposes Write() method that enables the in-memory contents to be
written to any other file stream as desired.

```go
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
```