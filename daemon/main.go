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
//TODO Cmdline
//TODO OVN Manager

package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	l "github.com/op/go-logging"

	"github.com/mbertrone/politoctrl/bpf"
	"github.com/mbertrone/politoctrl/helper"
)

var listenSocket string
var hoverUrl string
var helpFlag bool
var Log = l.MustGetLogger("politoctrl")
var format = l.MustStringFormatter(`%{color}%{time:15:04:05.000} %{shortfunc} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}`)

func init() {
	const (
		hoverDefault        = ""
		hoverHelp           = "Local hover URL"
		listenSocketDefault = "127.0.0.1:5001"
		listenSocketHelp    = "address:port to listen for updates"
	)
	flag.StringVar(&hoverUrl, "hover", hoverDefault, hoverHelp)
	flag.StringVar(&listenSocket, "listen", listenSocketDefault, listenSocketHelp)
	flag.BoolVar(&helpFlag, "h", false, "print this help")

	flag.Usage = func() {
		//TODO manage multiple hover clients
		fmt.Printf("Usage: %s -hover http://localhost:5000\n", filepath.Base(os.Args[0]))
		fmt.Printf(" -hover   URL       %s (default=%s)\n", hoverHelp, hoverDefault)
		fmt.Printf(" -listen  ADDR:PORT %s (default=%s)\n", listenSocketHelp, listenSocketDefault)
	}
}

func main() {
	//TODO How to separate log in a different package?

	/*----LOGGING-----*/
	// For demo purposes, create two backend for os.Stderr.
	backend1 := l.NewLogBackend(os.Stderr, "", 0)
	backend2 := l.NewLogBackend(os.Stderr, "", 0)

	// For messages written to backend2 we want to add some additional
	// information to the output, including the used log level and the name of
	// the function.
	backend2Formatter := l.NewBackendFormatter(backend2, format)

	// Only errors and more severe messages should be sent to backend1
	backend1Leveled := l.AddModuleLevel(backend1)
	backend1Leveled.SetLevel(l.CRITICAL, "")

	// Set the backends to be used.
	l.SetBackend(backend1Leveled, backend2Formatter)
	/*---------------------*/

	//Start Polito Controller
	//Parse Cmdline args
	//TODO start without one simgle hover, but can change hover(s) connections

	flag.Parse()
	if helpFlag {
		flag.Usage()
		os.Exit(0)
	}
	if len(hoverUrl) == 0 {
		Log.Error("missing argument -hover")
		flag.Usage()
		os.Exit(1)
	}

	//Connect to hover and initialize HoverDataplane
	//TODO manage multiple hosts (arrays/maps oh HoverDataplane)

	dataplane := helper.NewDataplane()

	if err := dataplane.Init(hoverUrl); err != nil {
		Log.Error("unable to conect to Hover %s\n%s\n", hoverUrl, err)
		os.Exit(1)
	}

	_, sw := helper.ModulePOST(dataplane, "bpf", "DummySwitch", bpf.DummySwitch2count)
	/*	_, l1 := */ helper.LinkPOST(dataplane, "i:veth1_", sw.Id)
	/*	_, l2 := */ helper.LinkPOST(dataplane, "i:veth2_", sw.Id)
	/*	_, l3 := */ helper.LinkPOST(dataplane, "i:veth3_", sw.Id)

	time.Sleep(time.Second * 8)

	_, kv := helper.TableEntryGET(dataplane, sw.Id, "count", "0x1")

	fmt.Printf("key: %s value: %s\n", kv.Key, kv.Value)

	err, kv := helper.TableEntryPUT(dataplane, sw.Id, "count", "0x1", "0x9")
	if err != nil {
		fmt.Printf("%s\n", err)
	}
	fmt.Printf("key: %s value: %s\n", kv.Key, kv.Value)

	_, kv = helper.TableEntryGET(dataplane, sw.Id, "count", "0x1")

	fmt.Printf("key: %s value: %s\n", kv.Key, kv.Value)

	_, kv = helper.TableEntryDELETE(dataplane, sw.Id, "count", "0x1")

	fmt.Printf("key: %s value: %s\n", kv.Key, kv.Value)
	_, kv = helper.TableEntryGET(dataplane, sw.Id, "count", "0x1")

	fmt.Printf("key: %s value: %s\n", kv.Key, kv.Value)

	/*
		fmt.Printf("id: %s\nfrom: %s\nto: %s\n", l.Id, l.From, l.To)
		errore, m := helper.ModuleListGET(dataplane)
		fmt.Printf("e:%s\n", errore)

		o, _ := json.Marshal(m.ListModules)
		fmt.Println(string(o))
	*/

	/*
		err, l := helper.LinkGET(dataplane, l1.Id)
		if err != nil {
			fmt.Printf("error: %s\n", err)
		}
		out, _ := json.Marshal(l)
		fmt.Println(string(out))
	*/

	/*
		helper.LinkDELETE(dataplane, l1.Id)
		time.Sleep(time.Second * 2)
		helper.LinkDELETE(dataplane, l2.Id)
		time.Sleep(time.Second * 2)
		helper.LinkDELETE(dataplane, l3.Id)
		time.Sleep(time.Second * 2)
	*/
	//		helper.ModuleDELETE(dataplane, sw.Id)

	/*
		for i := 0; i < 10; i++ {
			reader := bufio.NewReader(os.Stdin)
			fmt.Print("$ ")
			text, _ := reader.ReadString('\n')
			fmt.Println(text)
			fmt.Println(common.Version)
			cli.Parse(text)

		}

		//non mi serve un nuovo server!
		//mi serve per ora un singolo dataplane

		//poi loop cmdline

		srv, err := politoctrl.NewServer(hoverUrl)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}


		politoctrl.Info.Printf("polito-ctrl Server listening on %s\n", listenSocket)
		http.ListenAndServe(listenSocket, srv.Handler())
		//class that manages cli requests
		//infinite loop
		//a := "ciao"
	*/
}
