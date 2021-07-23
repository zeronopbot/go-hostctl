package go_hostctl

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestHostsFileCtl_Parse(t *testing.T) {
	hctl, err := NewHostFileCtl("testdata/etc/hosts/mixed_hosts")
	if err != nil {
		t.Fatal(err)
	}

	t.Log(len(hctl.Entries), hctl.HostsFile)

	for n, entry := range hctl.Entries {
		t.Log(n, entry.isComment, string(entry.rawLine))
		if entry.isComment {
			continue
		}

		t.Logf("\t - %s", entry.IPAddress)
		t.Logf("\t - %s", entry.Hostname)
		t.Logf("\t - %s", entry.Aliases)
		t.Logf("\t - %s", entry.Comment)
	}
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

	if err := hctl.Write(f); err != nil {
		t.Fatal(err)
	}
}
