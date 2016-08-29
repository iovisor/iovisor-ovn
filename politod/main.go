// Copyright 2016 PLUMgrid
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

//Bertrone Matteo - Polytechnic of Turin - 20-07-2016 first modify
//08-08-2016 - Testing simple module dummt switch
//18-08-2016 - Helper to talk with Hover & Logger
//21-08/2016 - Monitor  OVN databases

package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	l "github.com/op/go-logging"

	"github.com/netgroup-polito/iovisor-ovn/common"
	"github.com/netgroup-polito/iovisor-ovn/helper"
	"github.com/netgroup-polito/iovisor-ovn/monitor"
	"github.com/netgroup-polito/iovisor-ovn/testenv"
)

var listenSocket string
var hoverUrl string
var helpFlag bool
var Log = l.MustGetLogger("politoctrl")

func init() {
	const (
		hoverDefault = "http://localhost:5002"
		hoverHelp    = "Local hover URL"
		//listenSocketDefault = "127.0.0.1:5002"
		//listenSocketHelp    = "address:port to listen for updates"
	)
	flag.StringVar(&hoverUrl, "hover", hoverDefault, hoverHelp)
	//	flag.StringVar(&listenSocket, "listen", listenSocketDefault, listenSocketHelp)
	flag.BoolVar(&helpFlag, "h", false, "print this help")

	flag.Usage = func() {
		//TODO manage multiple hover clients
		fmt.Printf("Usage: %s -hover http://localhost:5002\n", filepath.Base(os.Args[0]))
		fmt.Printf(" -hover   URL       %s (default=%s)\n", hoverHelp, hoverDefault)
		//	fmt.Printf(" -listen  ADDR:PORT %s (default=%s)\n", listenSocketHelp, listenSocketDefault)
	}
}

//Start Polito Controller Daemon
func main() {

	//Init Logger
	common.LogInit()

	//Parse Cmdline args
	flag.Parse()
	if helpFlag {
		flag.Usage()
		os.Exit(0)
	}

	//TODO start without one simgle hover, but can change hover(s) connections
	if len(hoverUrl) == 0 {
		fmt.Println("missing argument -hover")
		flag.Usage()
		os.Exit(1)
	}

	//TODO manage multiple hosts (arrays/maps oh HoverDataplane)
	dataplane := helper.NewDataplane()

	//Connect to hover and initialize HoverDataplane
	if err := dataplane.Init(hoverUrl); err != nil {
		Log.Errorf("unable to conect to Hover %s\n%s\n", hoverUrl, err)
		os.Exit(1)
	}

	//Start monitoring ovn/s databases
	go monitor.MonitorOvsDb()
	go monitor.MonitorOvnNb()
	go monitor.MonitorOvnSb()

	//simple test enviroment (see testenv/env.go)
	go testenv.TestEnv(dataplane)

	//wait forever. if main is killed, Go kills all other goroutines
	quit := make(chan bool)
	<-quit
}
