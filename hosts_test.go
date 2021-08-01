package go_hostctl

import (
	"io/ioutil"
	"log"
	"os"
	"testing"
)

var (
	hostEntry1 *HostEntry
	hostEntry2 *HostEntry
	hostEntry3 *HostEntry
	hostEntry4 *HostEntry
)

func init() {
	var err error
	hostEntry1, err = NewHostEntry("1.1.1.1","host_one", "host entry one", "h1")
	if err != nil {
		log.Fatal(err)
	}

	hostEntry2, err = NewHostEntry("2.2.2.2","host_two", "host entry two", "h2")
	if err != nil {
		log.Fatal(err)
	}

	hostEntry3, err = NewHostEntry("3.3.3.3","host_three", "host entry three", "h3")
	if err != nil {
		log.Fatal(err)
	}

	hostEntry4, err = NewHostEntry("4.4.4.4","host_four", "host entry four", "h4")
	if err != nil {
		log.Fatal(err)
	}
}

func TestHostsFileCtl_Parse(t *testing.T) {
	hctl, err := NewHostFileCtl("testdata/etc/hosts/mixed_hosts")
	if err != nil {
		t.Fatal(err)
	}

	hctl.Write(os.Stdout)
}

func TestHostsFileCtl_Delete(t *testing.T) {
	f, err := ioutil.TempFile(os.TempDir(), "TestHostsFileCtl_Delete")
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

	for _, entry := range []*HostEntry{hostEntry1, hostEntry2, hostEntry3, hostEntry4} {
		if err := hctl.Add(*entry, -1); err != nil {
			t.Fatal(f)
		}
	}

	hctl.Write(os.Stdout)

	if err := hctl.Delete(1); err != nil {
		t.Fatal(err)
	}

	if err := hctl.Delete(-1); err != nil {
		t.Fatal(err)
	}

	if err := hctl.Delete(1); err != nil {
		t.Fatal(err)
	}

	entries, err := hctl.GetIP("1.1.1.1")
	if err != nil {
		t.Fatal(err)
	}

	for _, entry := range entries {
		if entry.Position != 0 {
			t.Fatalf("expecting entry in position 0 got %d", entry.Position)
		}
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

	entry, err := NewHostEntry("", "", "Host comment", "")
	if err != nil {
		t.Fatal(err)
	}

	if err := hctl.Add(*entry, 0); err != nil {
		t.Fatal(err)
	}

	if err := hctl.Add(*hostEntry1, -1); err != nil {
		t.Fatal(err)
	}

	if err := hctl.Add(*entry, 0); err != nil {
		t.Fatal(err)
	}

	if err := hctl.Add(*hostEntry2, 1); err != nil {
		t.Fatal(err)
	}

	entries, err := hctl.GetIP("1.1.1.1")
	if err != nil {
		t.Fatal(err)
	}

	for _, entry := range entries {
		if entry.Position != 3 {
			t.Fatalf("expecting entry in position 3 got %d", entry.Position)
		}
	}

	entries, err = hctl.GetIP("2.2.2.2")
	if err != nil {
		t.Fatal(err)
	}

	for _, entry := range entries {
		if entry.Position != 1 {
			t.Fatalf("expecting entry in position 1 got %d", entry.Position)
		}
	}

	hctl.Write(os.Stdout)
}

func TestHostsFileCtl_GetIP(t *testing.T) {

	hctl, err := NewHostFileCtl("testdata/etc/hosts/mixed_hosts")
	if err != nil {
		t.Fatal(err)
	}

	entries, err := hctl.GetIP("127.0.1.1")
	if err != nil {
		t.Fatal(err)
	}

	if len(entries) != 1 {
		t.Fatalf("expecting one entry, got: %d", len(entries))
	}

	if entries[0].Position != 29 {
		t.Fatal(err)
	}

	hctl.Write(os.Stdout)
}

func TestHostsFileCtl_GetAlias(t *testing.T) {

	hctl, err := NewHostFileCtl("testdata/etc/hosts/mixed_hosts")
	if err != nil {
		t.Fatal(err)
	}

	entries, err := hctl.GetAlias("some_macos")
	if err != nil {
		t.Fatal(err)
	}

	if len(entries) != 1 {
		t.Fatalf("expecting one entry, got: %d", len(entries))
	}

	if entries[0].Position != 29 {
		t.Fatal(err)
	}

	hctl.Write(os.Stdout)
}

func TestHostsFileCtl_GetHostname(t *testing.T) {

	hctl, err := NewHostFileCtl("testdata/etc/hosts/mixed_hosts")
	if err != nil {
		t.Fatal(err)
	}

	entries, err := hctl.GetHostname("some_macos.local")
	if err != nil {
		t.Fatal(err)
	}

	if len(entries) != 1 {
		t.Fatalf("expecting one entry, got: %d", len(entries))
	}

	if entries[0].Position != 29 {
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
