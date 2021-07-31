package go_hostctl

import (
	"io/ioutil"
	"net"
	"os"
	"testing"
)

func TestHostsFileCtl_Parse(t *testing.T) {
	hctl, err := NewHostFileCtl("testdata/etc/hosts/mixed_hosts")
	if err != nil {
		t.Fatal(err)
	}

	hctl.Write(os.Stdout)
}

func TestHostsFileCtl_Add(t *testing.T) {

	f, err := ioutil.TempFile(os.TempDir(), "TestHostsFileCtl_Add")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		defer os.Remove(f.Name())

		out, err := ioutil.ReadFile(f.Name())
		if err != nil {
			t.Fatal(err)
		}

		t.Logf("\r\n%s", out)
	}()
	defer f.Close()

	hctl, err := NewHostFileCtl(f.Name())
	if err != nil {
		t.Fatal(err)
	}

	entry, err := NewHostEntry("", "", "# This is a comment", "")
	if err != nil {
		t.Fatal(err)
	}

	if err := hctl.Add(*entry, -1); err != nil {
		t.Fatal(err)
	}

	entry.IPAddress = net.ParseIP("8.8.8.8")
	entry.Hostname = "google-host"
	entry.Aliases = []string{"g1", "g2", "g3"}
	entry.Comment = "# Some google host"
	if err := hctl.Add(*entry, 1); err != nil {
		t.Fatal(err)
	}

	hctl.Write(os.Stdout)
}

func TestHostsFileCtl_Write(t *testing.T) {

	f, err := ioutil.TempFile(os.TempDir(), "hostfilectl-test")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		defer os.Remove(f.Name())

		out, err := ioutil.ReadFile(f.Name())
		if err != nil {
			t.Fatal(err)
		}

		t.Logf("\r\n%s", out)
	}()
	defer f.Close()

	hctl, err := NewHostFileCtl("testdata/etc/hosts/mixed_hosts")
	if err != nil {
		t.Fatal(err)
	}

	if _, err := hctl.Write(f); err != nil {
		t.Fatal(err)
	}
}
