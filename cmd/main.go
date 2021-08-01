package main

import (
	"flag"
	. "github.com/zeronopbot/go-hostctl"
	"log"
	"os"
)

func main() {

	fpath := flag.String("f", "/tmp/hosts", "Hosts file path")
	mode := flag.String("m", "", "Mode")
	entry := flag.String("e", "", "Entry")
	flag.Parse()

	hctl, err := NewHostFileCtl(*fpath)
	if err != nil {
		log.Fatal(err)
	}

	e, err := ParseHostEntryLine([]byte(*entry))
	if err != nil {
		log.Fatal(err)
	}

	switch *mode {
	case "get":

	}

}
