package go_hostctl

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestHostsFileCtl_Parse(t *testing.T) {

	hosts := new(HostsFileCtl)
	if err := hosts.Parse( "testdata/etc/hosts/mixed_hosts"); err != nil {
		t.Fatal(err)
	}
}

func TestHostsFileCtl_Write(t *testing.T) {

	f, err := ioutil.TempFile(os.TempDir(), "hostfilectl")
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

	hosts := new(HostsFileCtl)
	if err := hosts.Parse( "testdata/etc/hosts/mixed_hosts"); err != nil {
		t.Fatal(err)
	}

	if err := hosts.Write(f); err != nil {
		t.Fatal(err)
	}
}