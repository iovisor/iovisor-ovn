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
	//"time"
	"fmt"

	"github.com/netgroup-polito/iovisor-ovn/common"
	"github.com/netgroup-polito/iovisor-ovn/config"
	"github.com/netgroup-polito/iovisor-ovn/hoverctl"
	"github.com/netgroup-polito/iovisor-ovn/iomodules/l2switch"
	l "github.com/op/go-logging"
)

var log = l.MustGetLogger("switch-test")

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

	log.Noticef("Creating and deploying switch module...")
	l2 := l2switch.Create(dataplane)
	if err := l2.Deploy(); err != nil {
		log.Errorf("%s", err)
		return
	}

	log.Noticef("Attaching external interfaces...")
	if err := l2.AttachExternalInterface("veth1_"); err != nil {
		log.Errorf("%s", err)
		return
	}

	if err := l2.AttachExternalInterface("veth2_"); err != nil {
		log.Errorf("%s", err)
		return
	}

	fmt.Println("Press enter to remove the module")
	bufio.NewReader(os.Stdin).ReadBytes('\n')

	log.Noticef("Detaching external interfaces...")
	if err := l2.DetachExternalInterface("veth1"); err != nil {
		log.Errorf("%s", err)
		return
	}

	if err := l2.DetachExternalInterface("veth2"); err != nil {
		log.Errorf("%s", err)
		return
	}

	log.Noticef("Destroying the switch module...")
	if err := l2.Destroy(); err != nil {
		log.Errorf("%s", err)
		return
	}

	log.Noticef("Done...")
}
