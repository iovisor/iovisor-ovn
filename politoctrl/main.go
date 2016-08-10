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

package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/mbertrone/politoctrl"
)

var listenSocket string
var hoverUrl string
var helpFlag bool
var port1 string
var port2 string

func init() {
	const (
		hoverDefault        = ""
		hoverHelp           = "Local hover URL"
		listenSocketDefault = "127.0.0.1:5001"
		listenSocketHelp    = "address:port to listen for updates"
		portHelp            = "Port/Interface to connect"
		port1Default        = "i:veth1_"
		port2Default        = "i:veth2_"
	)
	flag.StringVar(&hoverUrl, "hover", hoverDefault, hoverHelp)
	flag.StringVar(&listenSocket, "listen", listenSocketDefault, listenSocketHelp)
	flag.BoolVar(&helpFlag, "h", false, "print this help")
	flag.StringVar(&port1, "port1", port1Default, portHelp)
	flag.StringVar(&port2, "port2", port2Default, portHelp)

	flag.Usage = func() {
		//TODO manage multiple hover clients
		fmt.Printf("Usage: %s -hover http://localhost:5000\n", filepath.Base(os.Args[0]))
		fmt.Printf(" -hover   URL       %s (default=%s)\n", hoverHelp, hoverDefault)
		fmt.Printf(" -listen  ADDR:PORT %s (default=%s)\n", listenSocketHelp, listenSocketDefault)
		fmt.Printf(" -port1   PORT      %s (default=%s)\n", portHelp, port1Default)
		fmt.Printf(" -port2   PORT      %s (default=%s)\n", portHelp, port2Default)
	}
}

func main() {
	flag.Parse()
	if helpFlag {
		flag.Usage()
		os.Exit(0)
	}
	if len(hoverUrl) == 0 {
		fmt.Println("Missing argument -hover")
		flag.Usage()
		os.Exit(1)
	}

	srv, err := politoctrl.NewServer(hoverUrl, port1, port2)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	politoctrl.Info.Printf("polito-ctrl Server listening on %s\n", listenSocket)
	http.ListenAndServe(listenSocket, srv.Handler())
}
