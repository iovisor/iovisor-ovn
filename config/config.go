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
package config

import (
	"fmt"
	"net"
	"strconv"
	"time"

	l "github.com/op/go-logging"
)

var Nb = "127.0.0.1:6641"
var NbSock = "/home/matteo/ovs/tutorial/sandbox/ovnnb_db.sock"
var Sb = "127.0.0.1:6642"
var SbSock = "/home/matteo/ovs/tutorial/sandbox/ovnsb_db.sock"
var Ovs = "127.0.0.1:6640"
var OvsSock = "/home/matteo/ovs/tutorial/sandbox/db.sock"

var Sandbox = false
var Hover = "http://localhost:5002"
var TestEnv = false

//Constant
const SwitchPortsNumber = 32

var SleepTime = 2000 * time.Millisecond
var FlushTime = 30000 * time.Millisecond
var FlushEnabled = false
var SwitchSecurityPolicy = true

var Debug = false
var Info = true

//for debug purposes, print Notification on single changes
var PrintOvnNbChanges = false
var PrintOvnSbChanges = false
var PrintOvsChanges = false

//for debug purposes, print the whole db on changes notification
var PrintOvnNbFilteredTables = false
var PrintOvnSbFilteredTables = false
var PrintOvsFilteredTables = false

//for debug purposes, print ALL the database
var PrintOvnNb = false
var PrintOvnSb = false
var PrintOvs = false

// file with the iomodules configuration to the used, if empty the deamon starts
// in "OpenStack/ovn mode"
var File = ""

var log = l.MustGetLogger("iovisor-ovn-daemon")

func PrintConfigCli() {
	fmt.Printf("***************CONFIGURATION*******************\n")
	// fmt.Printf("%30s:%d\n", "SleepTime", SleepTime)
	// fmt.Printf("%30s:%d\n", "FlushTime", FlushTime)
	// fmt.Printf("%30s:%d\n\n", "FlushEnabled", FlushEnabled)
	fmt.Printf("%30s: %t\n", "PrintOvnNbChanges", PrintOvnNbChanges)
	fmt.Printf("%30s: %t\n", "PrintOvnSbChanges", PrintOvnSbChanges)
	fmt.Printf("%30s: %t\n\n", "PrintOvsChanges", PrintOvsChanges)
	fmt.Printf("%30s: %t\n", "PrintOvnNbFilteredTables", PrintOvnNbFilteredTables)
	fmt.Printf("%30s: %t\n", "PrintOvnSbFilteredTables", PrintOvnSbFilteredTables)
	fmt.Printf("%30s: %t\n\n", "PrintOvsFilteredTables", PrintOvsFilteredTables)
	fmt.Printf("%30s: %t\n", "PrintOvnNb", PrintOvnNb)
	fmt.Printf("%30s: %t\n", "PrintOvnSb", PrintOvnSb)
	fmt.Printf("%30s: %t\n", "PrintOvs", PrintOvs)

	fmt.Printf("************************************************\n")
}

func PrintConfig() {
	fmt.Printf("-----IOVisor-OVN Daemon---------------------------------------\n")
	fmt.Printf("starting configuration\n")

	if !Sandbox {
		fmt.Printf("attaching to Openstack\n")
		fmt.Printf("%8s:%s\n", "Nb", Nb)
		fmt.Printf("%8s:%s\n", "Sb", Sb)
		fmt.Printf("%8s:%s\n", "Ovs", Ovs)
	} else {
		fmt.Printf("attaching to Sandbox\n")
		fmt.Printf("%8s:%s\n", "NbSock", NbSock)
		fmt.Printf("%8s:%s\n", "SbSock", SbSock)
		fmt.Printf("%8s:%s\n", "OvsSock", OvsSock)
	}
	fmt.Printf("%8s:%s\n", "Hover", Hover)
	// fmt.Printf("%8s:%t\n", "TestEnv", TestEnv)
	fmt.Printf("--------------------------------------------------------------\n\n")
}

func FromStringToIpPort(s string) (string, int) {
	host, portStr, err := net.SplitHostPort(s)
	if err != nil {
		log.Errorf("Error in parsing address %s : %s\n", s, err)
		return "", -1
	}
	port, errp := strconv.Atoi(portStr)
	if errp != nil {
		log.Errorf("Error in converting port %s : %s\n", portStr, err)
		return "", -1
	}
	return host, port
}
