package main

import (
	"flag"
	. "github.com/zeronopbot/go-hostctl"
	"log"
	"os"
)

func main() {

	fpath := flag.String("f", "/tmp/hosts", "Hosts file path")
	name := flag.String("n", "", "Host name")
	alias := flag.String("a", "", "Host alias")
	ipAddr := flag.String("i", "", "IP address")
	comment := flag.String("c", "", "Comment for new host")
	flag.Parse()

	ctl, err := NewHostFileCtl(*fpath)
	if err != nil {
		log.Fatal(err)
	}

	entry, err := NewHostEntry(*ipAddr, *name, *comment, *alias)
	if err != nil {
		log.Fatalf("failed to create new host entry: %s", err)
	}

	// Add entry to end of file
	if err := ctl.Add(*entry, -1); err != nil {
		log.Fatal(err)
	}

	entry.Hostname = "another_host"
	if err := ctl.Add(*entry, 0); err != nil {
		log.Fatal(err)
	}

	hdr := HostEntry{
		Comment: "# Some google stuff",
	}
	if err := ctl.Add(hdr, 1); err != nil {
		log.Fatal(err)
	}

	entries, err := ctl.GetAlias("some_alias")
	if err != nil {
		log.Fatal(err)
	}

	for n, e := range entries {
		log.Printf("%d - %+v", n+1, e)
	}

	ctl.Write(os.Stdout)
}
