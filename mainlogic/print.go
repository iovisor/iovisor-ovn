// Copyright 2017 Politecnico di Torino
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

	"github.com/olekukonko/tablewriter"
)

func PrintL2Switch(name string) {
	if sw, ok := switches[name]; ok {
		table := tablewriter.NewWriter(os.Stdout)
		if sw.swIomodule != nil {
			// table = tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"SWITCH", "MODULE-ID", "PORTS#"})
			table.Append([]string{sw.Name, sw.swIomodule.ModuleId, strconv.Itoa(len(sw.swIomodule.Interfaces))})
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
			table.SetHeader([]string{"SWITCH", "NAME", "LINK", "REDIRECT"})
			for _, iface := range sw.swIomodule.Interfaces {
				table.Append([]string{sw.Name, iface.IfaceName, iface.LinkIdHover, strconv.Itoa(iface.IfaceId)})
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
		table.Append([]string{swname, sw.swIomodule.ModuleId, strconv.Itoa(len(sw.swIomodule.Interfaces))})
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
	if _, ok := routers[name]; ok {

		if r, ok := routers[name]; ok {
			table := tablewriter.NewWriter(os.Stdout)
			if r.rIoModule != nil {
				// table = tablewriter.NewWriter(os.Stdout)
				table.SetHeader([]string{"ROUTER", "MODULE-ID", "PORTS#"})
				table.Append([]string{r.Name, r.rIoModule.ModuleId, strconv.Itoa(r.rIoModule.PortsCount)})
				table.Render()
			} else {
				// table = tablewriter.NewWriter(os.Stdout)
				table.SetHeader([]string{"ROUTER", "MODULE-ID", "PORTS#"})
				table.Append([]string{r.Name, r.rIoModule.ModuleId, " "})
				table.Render()
			}

			table = tablewriter.NewWriter(os.Stdout)
			table.SetHeader([]string{"ROUTER", "NAME", "IP", "NETMASK", "MAC"})
			for _, rp := range r.ports {
				table.Append([]string{r.Name, rp.Name, rp.IP, rp.Mask, rp.Mac})
			}
			table.Render()

			if r.rIoModule != nil {
				table = tablewriter.NewWriter(os.Stdout)
				table.SetHeader([]string{"ROUTER", "NAME", "LINK", "FD", "REDIRECT", "IP", "NETMASK", "MAC"})
				for _, iface := range r.rIoModule.Interfaces {
					table.Append([]string{r.Name, iface.IfaceName, iface.LinkIdHover, strconv.Itoa(iface.IfaceIdRedirectHover), iface.IP, iface.Netmask, iface.IP})
				}
				table.Render()
			}
		}
	}
}

func PrintRouters(verbose bool) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ROUTER", "MODULE-ID", "PORTS#"})
	for _, r := range routers {
		table.Append([]string{r.Name, r.rIoModule.ModuleId, strconv.Itoa(r.rIoModule.PortsCount)})
	}
	table.Render()

	if verbose {
		for routername, _ := range routers {
			fmt.Printf("\n")
			PrintRouter(routername)
		}
	}
}

func PrintMainLogic(verbose bool) {
	fmt.Printf("\nSwitches\n\n")
	PrintL2Switches(verbose)
	fmt.Printf("\nRouters\n\n")
	PrintRouters(verbose)
}
