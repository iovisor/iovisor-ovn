package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	l "github.com/op/go-logging"

	"github.com/netgroup-polito/iovisor-ovn/cli"
	"github.com/netgroup-polito/iovisor-ovn/common"
	"github.com/netgroup-polito/iovisor-ovn/hoverctl"
	"github.com/netgroup-polito/iovisor-ovn/mainlogic"
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

//Start iovisor-ovn Daemon
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
	dataplane := hoverctl.NewDataplane()

	//Init Bpf loadc files
	//bpf.Init()

	//Connect to hover and initialize HoverDataplane
	if err := dataplane.Init(hoverUrl); err != nil {
		Log.Errorf("unable to conect to Hover %s\n%s\n", hoverUrl, err)
		os.Exit(1)
	}

	//simple test enviroment (see testenv/env.go)
	//go testenv.TestEnv(dataplane)

	//Montiors started here!
	mainlogic.MainLogic(dataplane)

	time.Sleep(500 * time.Millisecond)
	//start simple cli
	go cli.Cli(dataplane)

	//wait forever. if main is killed, Go kills all other goroutines
	quit := make(chan bool)
	<-quit
}
