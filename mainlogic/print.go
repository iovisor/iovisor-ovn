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
package mainlogic

import (
	"fmt"
	"os"
	"strconv"

	"github.com/davecgh/go-spew/spew"
	"github.com/olekukonko/tablewriter"
)

func PrintL2Switch(name string) {
	if sw, ok := switches[name]; ok {
		table := tablewriter.NewWriter(os.Stdout)
		if sw.swIomodule != nil {
			// table = tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"SWITCH", "MODULE-ID", "PORTS#"})
			table.Append([]string{sw.Name, sw.swIomodule.ModuleId, strconv.Itoa(sw.swIomodule.PortsCount)})
			table.Render()
		} else {
			// table = tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"SWITCH", "MODULE-ID", "PORTS#"})
			table.Append([]string{sw.Name, sw.swIomodule.ModuleId, " "})
			table.Render()
		}

		table = tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"SWITCH", "NAME", "IFACE"})
		for _, swp := range sw.ports {
			table.Append([]string{sw.Name, swp.Name, swp.IfaceName})
		}
		table.Render()

		if sw.swIomodule != nil {
			table = tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"SWITCH", "NAME", "LINK", "FD", "BCASTID", "REDIRECT"})
			for _, iface := range sw.swIomodule.Interfaces {
				table.Append([]string{sw.Name, iface.IfaceName, iface.LinkIdHover, strconv.Itoa(iface.IfaceFd), strconv.Itoa(iface.IfaceIdArrayBroadcast), strconv.Itoa(iface.IfaceIdRedirectHover)})
			}
			table.Render()
		}
		//TODO We need to print ports array?
		//spew.Dump(sw.swIomodule.PortsArray)
	}
}

func PrintL2Switches(verbose bool) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"SWITCHES", "MODULE-ID", "PORTS#"})
	for swname, sw := range switches {
		table.Append([]string{swname, sw.swIomodule.ModuleId, strconv.Itoa(sw.swIomodule.PortsCount)})
	}
	table.Render()

	if verbose {
		for swname, _ := range switches {
			fmt.Printf("\n")
			PrintL2Switch(swname)
		}
	}
}

func PrintRouter(name string) {
	if router, ok := routers[name]; ok {
		//TODO Improve. First implementation
		spew.Dump(router)
	}
}

func PrintRouters() {
	for routername, _ := range routers {
		PrintRouter(routername)
	}
}

func PrintMainLogic(verbose bool) {
	fmt.Printf("\nSwitches\n\n")
	PrintL2Switches(verbose)
	fmt.Printf("\nRouters\n\n")
	PrintRouters()
}
