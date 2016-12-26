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
package ovnmonitor

import (
	"fmt"
	"os"
	"strconv"

	"github.com/olekukonko/tablewriter"
)

func PrintLogicalSwitchByName(name string, db *OvnDB) {
	if sw, ok := db.Switches[name]; ok {
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"SWITCH-UUID", "NAME", "MODIFIED"})
		table.Append([]string{sw.uuid, sw.Name, strconv.FormatBool(sw.Modified)})
		table.Render()

		table = tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"PORTS-UUID", "NAME", "IFACE", "TYPE", "RPORT", "MODIFIED"})
		for _, port := range sw.Ports {
			table.Append([]string{port.uuid, port.Name, port.IfaceName, port.Type, port.RouterPort, strconv.FormatBool(port.Modified)})
		}
		table.Render()
	}
}

func PrintLogicalRouterByName(name string, db *OvnDB) {
	table := tablewriter.NewWriter(os.Stdout)
	if r, ok := db.Routers[name]; ok {
		table.SetHeader([]string{"ROUTERS-UUID", "NAME", "MODIFIED", "ENABLED"})
		table.Append([]string{r.uuid, r.Name, strconv.FormatBool(r.Modified), strconv.FormatBool(r.Enabled)})
		table.Render()

		table = tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"PORTS-UUID", "NAME", "NETWORKS", "MAC", "MODIFIED", "ENABLED"})
		for _, port := range r.Ports {
			table.Append([]string{port.uuid, port.Name, port.Networks, port.Mac, strconv.FormatBool(port.Modified), strconv.FormatBool(port.Enabled)})
		}
		table.Render()
	}
}

func PrintOvsInterfaces(db *OvnDB) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"INTERFACES-UUID", "NAME", "EXTERNAL-IDS"})
	for _, iface := range db.ovsInterfaces {
		table.Append([]string{iface.uuid, iface.Name, iface.ExternalIdIface})
	}
	table.Render()
}

func PrintLogicalSwitches(verbose bool, db *OvnDB) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"SWITCHES-UUID", "NAME", "MODIFIED"})
	for _, sw := range db.Switches {
		table.Append([]string{sw.uuid, sw.Name, strconv.FormatBool(sw.Modified)})
	}
	table.Render()

	if verbose {
		for swname, _ := range db.Switches {
			fmt.Printf("\n")
			PrintLogicalSwitchByName(swname, db)
		}
	}
}

func PrintLogicalRouters(verbose bool, db *OvnDB) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"ROUTERS-UUID", "NAME", "MODIFIED", "ENABLED"})
	for _, r := range db.Routers {
		table.Append([]string{r.uuid, r.Name, strconv.FormatBool(r.Modified), strconv.FormatBool(r.Enabled)})
	}
	table.Render()

	if verbose {
		for rname, _ := range db.Routers {
			fmt.Printf("\n")
			PrintLogicalRouterByName(rname, db)
		}
	}
}

func PrintOvnMonitor(verbose bool, db *OvnDB) {
	fmt.Printf("\nLogical Switches\n\n")
	PrintLogicalSwitches(verbose, db)
	fmt.Printf("\nLogical Routers\n\n")
	PrintLogicalRouters(verbose, db)
	fmt.Printf("\nInterfaces\n\n")
	PrintOvsInterfaces(db)
}
