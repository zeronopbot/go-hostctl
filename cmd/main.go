package main

import (
	"flag"
	. "github.com/zeronopbot/go-hostctl"
	"log"
)

const (
	HostControlModeCreate = "create"
	HostControlModeRead   = "read"
	HostControlModeUpdate = "update"
	HostControlModeDelete = "delete"
)

func main() {

	fpath := flag.String("f", "/etc/hosts", "Hosts file path")
	name := flag.String("n", "", "Host name")
	alias := flag.String("a", "", "Host alias")
	ipAddr := flag.String("i", "", "IP address")
	comment := flag.String("c", "", "Comment for new host")
	operation := flag.String("m", HostControlModeRead, "Host control command")
	flag.Parse()

	entry, err := NewHostEntry(*ipAddr, *name, *alias, *comment)
	if err != nil {
		log.Fatalf("failed to build new host entry: %s", err)
	}

	log.Println(entry, fpath, operation)
}
