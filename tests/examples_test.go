// Copyright 2017 Politecnico di Torino
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package tests

import (
	"os/exec"
	_ "fmt"
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/iovisor/iovisor-ovn/servicetopology"

	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
)

func prepareEnvironment(t *testing.T, path string) {
	err := exec.Command(path).Run()
	if err != nil {
		t.Fatal("error setting environment", err)
	}
}

func startHover(t *testing.T) func() {
	path := os.Getenv("GOPATH") + "/bin/hoverd"
	cmd := exec.Command(path, "-listen", "127.0.0.1:5002")

	err := cmd.Start()
	if err != nil {
		t.Fatal("Error starting hover: ", err)
	}

	time.Sleep(5 * time.Second)

	return func() {
		if err := cmd.Process.Kill(); err != nil {
			t.Fatal("failed to kill hover: ", err)
		}
	}
}

func deployTopology(t *testing.T, path string) {
	err := servicetopology.DeployTopology(path)
	if err != nil {
		t.Fatal("Error deploying switch")
	}
}

func TestSwitchExample(t *testing.T) {
	base := os.Getenv("GOPATH") + "/src/github.com/iovisor/iovisor-ovn/examples/"

	// prepare environment
	prepareEnvironment(t, base + "switch/setup.sh")

	// start hoverd
	kill_hover := startHover(t)
	defer kill_hover()

	// deploy example
	deployTopology(t, base + "switch/switch.yaml")

	// test connectivity
	var out []byte
	var err error
	out, err = exec.Command("ip", "netns", "exec", "ns1", "ping", "-c", "1", "10.0.0.2").Output()
	if err != nil {
		t.Error(string(out), err)
	}

	out, err = exec.Command("ip", "netns", "exec", "ns2", "ping", "-c", "1", "10.0.0.1").Output()
	if err != nil {
		t.Error(string(out), err)
	}
}

func TestRouterExample(t *testing.T) {
	base := os.Getenv("GOPATH") + "/src/github.com/iovisor/iovisor-ovn/examples/"

	// prepare environment
	prepareEnvironment(t, base + "router/setup.sh")

	// start hoverd
	kill_hover := startHover(t)
	defer kill_hover()

	// deploy example
	deployTopology(t, base + "router/router.yaml")

	// test connectivity
	var out []byte
	var err error
	out, err = exec.Command("ip", "netns", "exec", "ns1",
		"ping", "-c", "1", "10.0.2.100").Output()
	if err != nil {
		t.Error(string(out), err)
	}

	out, err = exec.Command("ip", "netns", "exec", "ns1",
		"ping", "-c", "1", "10.0.3.100").Output()
	if err != nil {
		t.Error(string(out), err)
	}

	out, err = exec.Command("ip", "netns", "exec", "ns3",
		"ping", "-c", "1", "10.0.1.100").Output()
	if err != nil {
		t.Error(string(out), err)
	}
}

func getIpAddressInNs(t *testing.T, ns string, iface string) string {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	var nsmain netns.NsHandle
	var err error

	nsmain, err = netns.Get()
	if err != nil {
		t.Fatal("Error getting ns", err)
	}

	path := "/var/run/netns/" + ns
	var nsH netns.NsHandle
	nsH, err = netns.GetFromPath(path)
	if err != nil {
		t.Fatal("Error getting ns", err)
	}

	err = netns.Set(nsH);
	if err != nil {
		t.Fatal("Error setting ns", err)
	}

	link, err := netlink.LinkByName(iface)
	if err != nil {
		t.Fatal("Error getting iface", err)
	}
	var addrs []netlink.Addr
	addrs, err = netlink.AddrList(link, netlink.FAMILY_V4)
	if err != nil {
		t.Fatal("Error getting addresses", err)
	}
	var ip string
	if len(addrs) > 0 {
		ip = addrs[0].IP.String()
	}

	err = netns.Set(nsmain);
	if err != nil {
		t.Fatal("Error setting ns", err)
	}

	return ip
}

func TestDhcpExample(t *testing.T) {
	base := os.Getenv("GOPATH") + "/src/github.com/iovisor/iovisor-ovn/examples/"

	// prepare environment
	prepareEnvironment(t, base + "dhcp/setup.sh")

	// start hoverd
	kill_hover := startHover(t)
	defer kill_hover()

	// deploy example
	deployTopology(t, base + "dhcp/dhcp.yaml")

	// perform dhcp requests
	var out []byte
	var err error
	out, err = exec.Command("ip", "netns", "exec", "ns1",
		"dhclient").Output()
	if err != nil {
		t.Error(string(out), err)
	}

	out, err = exec.Command("ip", "netns", "exec", "ns2",
		"dhclient").Output()
	if err != nil {
		t.Error(string(out), err)
	}

	out, err = exec.Command("ip", "netns", "exec", "ns3",
		"dhclient",).Output()
	if err != nil {
		t.Error(string(out), err)
	}

	// get the ips assigend to those ifaces
	ip1 := getIpAddressInNs(t, "ns1", "veth1_")
	if ip1 == "" {
		t.Error("No ip assigned to veth1_")
	}

	ip2 := getIpAddressInNs(t, "ns2", "veth2_")
	if ip2 == "" {
		t.Error("No ip assigned to veth1_")
	}

	ip3 := getIpAddressInNs(t, "ns3", "veth3_")
	if ip3 == "" {
		t.Error("No ip assigned to veth1_")
	}

	// ping between the namespaces
	out, err = exec.Command("ip", "netns", "exec", "ns1",
		"ping", "-c", "1", ip2).Output()
	if err != nil {
		t.Error(string(out), err)
	}

	out, err = exec.Command("ip", "netns", "exec", "ns1",
		"ping", "-c", "1", ip3).Output()
	if err != nil {
		t.Error(string(out), err)
	}

	out, err = exec.Command("ip", "netns", "exec", "ns3",
		"ping", "-c", "1", ip1).Output()
	if err != nil {
		t.Error(string(out), err)
	}
}
