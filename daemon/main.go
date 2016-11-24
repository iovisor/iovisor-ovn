package main

import (
	"flag"

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

	quit := make(chan bool)
	<-quit
}
