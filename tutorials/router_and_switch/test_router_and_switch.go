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

import (
	"flag"
	"os"
	"bufio"
	"net"
	"fmt"

	"github.com/netgroup-polito/iovisor-ovn/common"
	"github.com/netgroup-polito/iovisor-ovn/config"
	"github.com/netgroup-polito/iovisor-ovn/hoverctl"
	"github.com/netgroup-polito/iovisor-ovn/iomodules"
	"github.com/netgroup-polito/iovisor-ovn/iomodules/router"
	"github.com/netgroup-polito/iovisor-ovn/iomodules/l2switch"
	l "github.com/op/go-logging"
)

var log = l.MustGetLogger("router-test")

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


// ********** SW1 *******************
	sw1 := l2switch.Create(dataplane)
	if err := sw1.Deploy(); err != nil {
		log.Errorf("%s", err)
		return
	}

// ********** SW2 *******************
	sw2 := l2switch.Create(dataplane)
	if err := sw2.Deploy(); err != nil {
		log.Errorf("%s", err)
		return
	}

// ********** Router *******************
	r := router.Create(dataplane)
	if err := r.Deploy(); err != nil {
		log.Errorf("%s", err)
		return
	}

// Connect R to SW1

	if err := iomodules.AttachIoModules(dataplane, sw1, "toRouter", r, "toSW1"); err != nil {
		log.Errorf("%s", err)
		return
	}

	if err := r.ConfigureInterface("toSW1", "10.0.1.1", "255.255.255.0", "3C:E9:F9:DA:A3:80"); err != nil {
		log.Errorf("%s", err)
		return
	}

	if err := sw1.AddForwardingTableEntry("3C:E9:F9:DA:A3:80", "toRouter"); err != nil {
		log.Errorf("%s", err)
		return
	}

// Connect R to SW2
	if err := iomodules.AttachIoModules(dataplane, sw2, "toRouter", r, "toSW2"); err != nil {
		log.Errorf("%s", err)
		return
	}

	if err := r.ConfigureInterface("toSW2", "10.0.2.1", "255.255.255.0", "3C:E9:F9:DA:A3:81"); err != nil {
		log.Errorf("%s", err)
		return
	}

	if err := sw2.AddForwardingTableEntry("3C:E9:F9:DA:A3:81", "toRouter"); err != nil {
		log.Errorf("%s", err)
		return
	}

// attach virtual interfaces to the switches

	if err := sw1.AttachExternalInterface("veth1"); err != nil {
		log.Errorf("%s", err)
		return
	}

	if err := sw1.AddForwardingTableEntry(getInterfaceMac("veth1"), "veth1"); err != nil {
		log.Errorf("%s", err)
		return
	}

	if err := sw2.AttachExternalInterface("veth2"); err != nil {
		log.Errorf("%s", err)
		return
	}

	if err := sw2.AddForwardingTableEntry(getInterfaceMac("veth2"), "veth2"); err != nil {
		log.Errorf("%s", err)
		return
	}


	fmt.Println("Press enter to remove the module")
	bufio.NewReader(os.Stdin).ReadBytes('\n')

}

func getInterfaceMac(iface string) string {
	ifc, err := net.InterfaceByName(iface)
	if err != nil {
		return ""
	}

	return ifc.HardwareAddr.String()
}
