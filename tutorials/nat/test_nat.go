// Copyright 2016 Politecnico di Torino
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
package main

/*
  This tutorial setup an environment with two namespaces
	ns1 -> veth1_ 10.10.1.1/24
	ns2 -> veth2_ 10.10.1.2/24

	We put a NAT in the middle. The NAT hides the ns1 ip address, and replace it
	with the PUBLIC_IP address 10.10.1.100

	NAT is not L2 aware, so we have to forse static arp entries in order to
	make it works.
	(ns1) arp -s 10.10.1.2   mac_veth2_
	(ns2) arp -s 10.10.1.100 mac_veth1_

	in order to test the connection we can use nc o iperf

	TCP:
	sudo ip netns exec ns2 iperf -s
	sudo ip netns exec ns1 iperf -c 10.10.1.2

	UDP:
	sudo ip netns exec ns2 iperf -u -s
	sudo ip netns exec ns1 iperf -u -c 10.10.1.2
*/

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"

	"github.com/netgroup-polito/iovisor-ovn/common"
	"github.com/netgroup-polito/iovisor-ovn/config"
	"github.com/netgroup-polito/iovisor-ovn/hoverctl"
	"github.com/netgroup-polito/iovisor-ovn/iomodules/nat"
	l "github.com/op/go-logging"
)

var log = l.MustGetLogger("nat-test")

func init() {
	flag.BoolVar(&config.Debug, "debug", false, "enable DEBUG level in logger")
	flag.BoolVar(&config.Info, "info", true, "enable INFO  level in logger")

	flag.StringVar(&config.Hover, "hover", config.Hover, "hover url")
}

func main() {

	//Parse Cmdline args
	flag.Parse()

	//Init Logger
	common.LogInit()

	log.Noticef("Starting...")

	dataplane := hoverctl.NewDataplane()

	// Connect to hover and initialize HoverDataplane
	if err := dataplane.Init(config.Hover); err != nil {
		log.Errorf("unable to conect to Hover %s\n%s\n", config.Hover, err)
		os.Exit(1)
	}

	log.Noticef("Creating and deploying nat module...")
	n := nat.Create(dataplane)
	if err := n.Deploy(); err != nil {
		log.Errorf("%s", err)
		return
	}

	publicIp := "10.10.1.100"
	log.Noticef("Set NAT public ip: " + publicIp)
	n.SetPublicIp(publicIp)

	publicMac := getIfaceMacInNs("ns1", "veth1_")
	log.Noticef("Set NAT public ip and mac : " + publicIp + " " + publicMac)
	n.SetPublicPortAddresses(publicIp, publicMac)

	log.Noticef("Attaching external interfaces...")

	if err := n.AttachExternalInterface("veth1"); err != nil {
		log.Errorf("%s", err)
		return
	}

	if err := n.AttachExternalInterface("veth2"); err != nil {
		log.Errorf("%s", err)
		return
	}

	fmt.Println("Press enter to remove the module")
	bufio.NewReader(os.Stdin).ReadBytes('\n')

	log.Noticef("Detaching external interfaces...")
	if err := n.DetachExternalInterface("veth1"); err != nil {
		log.Errorf("%s", err)
		return
	}

	if err := n.DetachExternalInterface("veth2"); err != nil {
		log.Errorf("%s", err)
		return
	}

	log.Noticef("Destroying the router module...")
	if err := n.Destroy(); err != nil {
		log.Errorf("%s", err)
		return
	}
}

func getInterfaceMac(iface string) string {
	ifc, err := net.InterfaceByName(iface)
	if err != nil {
		return ""
	}

	return ifc.HardwareAddr.String()
}

func getIfaceMacInNs(namespace string, iface string) string {

	var (
		cmdOut []byte
		err    error
	)

	cmdName := "sudo"
	cmdArgs := []string{"ip", "netns", "exec", namespace, "ifconfig", iface /*, "|", "grep", "veth" "|", "awk", "'{print $5}'" */}
	if cmdOut, err = exec.Command(cmdName, cmdArgs...).Output(); err == nil {
		cmdStr := string(cmdOut)
		res := strings.Fields(cmdStr)
		return res[4]
	} else {
		return ""
	}
}
