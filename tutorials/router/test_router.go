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
	"bufio"
	"flag"
	"fmt"
	"net"
	"os"

	"github.com/netgroup-polito/iovisor-ovn/common"
	"github.com/netgroup-polito/iovisor-ovn/config"
	"github.com/netgroup-polito/iovisor-ovn/hoverctl"
	"github.com/netgroup-polito/iovisor-ovn/iomodules/router"
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

	log.Noticef("Creating and deploying router module...")
	r := router.Create(dataplane)
	if err := r.Deploy(); err != nil {
		log.Errorf("%s", err)
		return
	}

	log.Noticef("Attaching external interfaces...")
	if err := r.AttachExternalInterface("veth1"); err != nil {
		log.Errorf("%s", err)
		return
	}

	if err := r.ConfigureInterface("veth1", "10.0.1.1", "255.255.255.0",
		getInterfaceMac("veth1")); err != nil {
		log.Errorf("%s", err)
		return
	}

	if err := r.AttachExternalInterface("veth2"); err != nil {
		log.Errorf("%s", err)
		return
	}

	if err := r.ConfigureInterface("veth2", "10.0.2.1", "255.255.255.0",
		getInterfaceMac("veth2")); err != nil {
		log.Errorf("%s", err)
		return
	}

	if err := r.AttachExternalInterface("veth3"); err != nil {
		log.Errorf("%s", err)
		return
	}

	if err := r.ConfigureInterface("veth3", "10.0.3.1", "255.255.255.0",
		getInterfaceMac("veth3")); err != nil {
		log.Errorf("%s", err)
		return
	}

	fmt.Println("Press enter to remove the module")
	bufio.NewReader(os.Stdin).ReadBytes('\n')

	log.Noticef("Detaching external interfaces...")
	if err := r.DetachExternalInterface("veth1"); err != nil {
		log.Errorf("%s", err)
		return
	}

	if err := r.DetachExternalInterface("veth2"); err != nil {
		log.Errorf("%s", err)
		return
	}

	if err := r.DetachExternalInterface("veth3"); err != nil {
		log.Errorf("%s", err)
		return
	}

	log.Noticef("Destroying the router module...")
	if err := r.Destroy(); err != nil {
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
