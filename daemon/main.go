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
	"time"

	"github.com/netgroup-polito/iovisor-ovn/cli"
	"github.com/netgroup-polito/iovisor-ovn/config"
	l "github.com/op/go-logging"

	"github.com/netgroup-polito/iovisor-ovn/common"
	"github.com/netgroup-polito/iovisor-ovn/mainlogic"
)

var Log = l.MustGetLogger("iovisor-ovn-daemon")

func init() {
	flag.StringVar(&config.Nb, "nb", config.Nb, "nb db address:port")
	flag.StringVar(&config.Sb, "sb", config.Sb, "sb db address:port")
	flag.StringVar(&config.Ovs, "ovs", config.Ovs, "ovs db address:port")

	flag.StringVar(&config.NbSock, "nbsock", config.NbSock, "nb db local .sock file")
	flag.StringVar(&config.SbSock, "sbsock", config.SbSock, "sb db local .sock file")
	flag.StringVar(&config.OvsSock, "ovssock", config.OvsSock, "ovs db local .sock file")

	flag.BoolVar(&config.Sandbox, "sandbox", false, "connect to sandbox with local .sock files")
	flag.BoolVar(&config.TestEnv, "testenv", false, "enable testenv")

	flag.BoolVar(&config.Debug, "debug", false, "enable DEBUG level in logger")
	flag.BoolVar(&config.Info, "info", true, "enable INFO  level in logger")

	flag.StringVar(&config.Hover, "hover", config.Hover, "hover url")
}

//Start iovisor-ovn Daemon
func main() {

	//Parse Cmdline args
	flag.Parse()
	config.PrintConfig()

	//Init Logger
	common.LogInit()

	//simple test enviroment (see testenv/env.go)
	//if config.TestEnv {
	//	go tests.TestEnv(dataplane)
	//}

	//Monitors started here!
	mainlogic.MainLogic()

	//Cli start
	time.Sleep(1 * time.Second)
	go cli.Cli(mainlogic.Dataplane)

	quit := make(chan bool)
	<-quit
}
