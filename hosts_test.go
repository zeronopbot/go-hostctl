package go_hostctl

import "testing"

func TestHostsFileCtl_ParseTabs(t *testing.T) {

	tabHosts := new(HostsFileCtl)
	if err := tabHosts.Parse( "testdata/etc/hosts/tab_hosts"); err != nil {
		t.Fatal(err)
	}


}