package main

import (
	"flag"
	"os"
	"time"

	"github.com/netgroup-polito/iovisor-ovn/config"
	"github.com/netgroup-polito/iovisor-ovn/testenv"
	l "github.com/op/go-logging"

	"github.com/netgroup-polito/iovisor-ovn/cli"
	"github.com/netgroup-polito/iovisor-ovn/common"
	"github.com/netgroup-polito/iovisor-ovn/global"
	"github.com/netgroup-polito/iovisor-ovn/hoverctl"
	"github.com/netgroup-polito/iovisor-ovn/mainlogic"
)

var Log = l.MustGetLogger("politoctrl")

func init() {
	flag.StringVar(&config.Nb, "nb", config.Nb, "nb db address:port")
	flag.StringVar(&config.Sb, "sb", config.Sb, "sb db address:port")
	flag.StringVar(&config.Ovs, "ovs", config.Ovs, "ovs db address:port")

	flag.StringVar(&config.NbSock, "nbsock", config.NbSock, "nb db local .sock file")
	flag.StringVar(&config.SbSock, "sbsock", config.SbSock, "sb db local .sock file")
	flag.StringVar(&config.OvsSock, "ovssock", config.OvsSock, "ovs db local .sock file")

	flag.BoolVar(&config.Sandbox, "sandbox", false, "connect to sandbox with local .sock files")
	flag.BoolVar(&config.TestEnv, "testenv", false, "enable testenv")

	flag.StringVar(&config.Hover, "hover", config.Hover, "hover url")
}

//Start iovisor-ovn Daemon
func main() {

	//Init Logger
	common.LogInit()

	//Parse Cmdline args
	flag.Parse()
	config.PrintConfig()

	//TODO manage multiple hosts (arrays/maps oh HoverDataplane)
	dataplane := hoverctl.NewDataplane()

	//Connect to hover and initialize HoverDataplane
	if err := dataplane.Init(config.Hover); err != nil {
		Log.Errorf("unable to conect to Hover %s\n%s\n", config.Hover, err)
		os.Exit(1)
	}

	//simple test enviroment (see testenv/env.go)
	if config.TestEnv {
		go testenv.TestEnv(dataplane)
	}

	//Montiors started here!
	go mainlogic.MainLogic(dataplane)

	time.Sleep(500 * time.Millisecond)

	//start simple cli
	go cli.Cli(global.Hh)

	//wait forever. if main is killed, Go kills all other goroutines
	quit := make(chan bool)
	<-quit
}
