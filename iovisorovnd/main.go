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
	"os"
	"time"

	"github.com/iovisor/iovisor-ovn/cli"
	"github.com/iovisor/iovisor-ovn/config"
	l "github.com/op/go-logging"

	"github.com/iovisor/iovisor-ovn/common"
	"github.com/iovisor/iovisor-ovn/mainlogic"
	"github.com/iovisor/iovisor-ovn/servicetopology"
)

var log = l.MustGetLogger("iovisor-ovn-daemon")

func init() {

	flag.StringVar(&config.File, "file", "", "path to configuration file")

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

	//Init Logger
	common.LogInit()

	if config.File != "" {
		// file has been passed, deploy this without connecting to OVN
		err := servicetopology.DeployTopology(config.File)
		if err != nil {
			log.Errorf("Error deploying topology, please verify your topology file");
		}
		log.Infof("Press enter to close");
		bufio.NewReader(os.Stdin).ReadBytes('\n')
		// TODO: defer call to servicetopologyUndeploy
	} else {
		// topologyFile argument was not passed, then connect to OVN
		config.PrintConfig()
		//Monitors started here!
		mainlogic.MainLogic()

		// Cli start
		time.Sleep(1 * time.Second)
		go cli.Cli(mainlogic.GetHoverClient())

		quit := make(chan bool)
		<-quit
	}
}
