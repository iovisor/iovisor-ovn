package main

import (
	"flag"
	"os"
	"time"

	"github.com/netgroup-polito/iovisor-ovn/config"
	"github.com/netgroup-polito/iovisor-ovn/hoverctl"
	"github.com/netgroup-polito/iovisor-ovn/ovnmonitor"
	"github.com/netgroup-polito/iovisor-ovn/testenv"
	l "github.com/op/go-logging"

	"github.com/netgroup-polito/iovisor-ovn/cli"
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

	//Init Global Handler
	globalHandler := ovnmonitor.HandlerHandler{}

	//TODO Check how many dataplanes should be initialized.
	//For now (single Hypervisor) only one

	//In future (multiple hypervisor) I can read this information from config package
	//Or read the information from the Main Nb database table (that points to hypervisors)

	dataplane := hoverctl.NewDataplane()

	//Connect to hover and initialize HoverDataplane
	if err := dataplane.Init(config.Hover); err != nil {
		Log.Errorf("unable to conect to Hover %s\n%s\n", config.Hover, err)
		os.Exit(1)
	}

	globalHandler.Dataplane = dataplane

	//simple test enviroment (see testenv/env.go)
	if config.TestEnv {
		go testenv.TestEnv(dataplane)
	}

	//Montiors started here!
	go mainlogic.MainLogic(&globalHandler)

	time.Sleep(500 * time.Millisecond)

	//start simple cli
	go cli.Cli(&globalHandler)

	//wait forever. if main is killed, Go kills all other goroutines
	quit := make(chan bool)
	<-quit
}
