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

var SleepTime = 3500 * time.Millisecond
var SwitchSecurityPolicy = true

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

var log = l.MustGetLogger("politoctrl")

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
	fmt.Printf("%8s:%t\n", "TestEnv", TestEnv)
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
