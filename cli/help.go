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
// Inline command line interface for debug purposes
// in future this cli will be a separate go program that connects to the main iovisor-ovn daemon
// in future this cli will use a cli go library (e.g. github.com/urfave/cli )

package cli

import "fmt"

func PrintTableUsage() {
	fmt.Printf("\nTable Usage\n\n")
	fmt.Printf("	table get\n")
	fmt.Printf("	table get <module-id>\n")
	fmt.Printf("	table get <module-id> <table-id>\n")
	fmt.Printf("	table get <module-id> <table-id> <entry-key>\n")
	fmt.Printf("	table put <module-id> <table-id> <entry-key> <entry-value>\n")
	fmt.Printf("	table delete <module-id> <table-id> <entry-key> <entry-value>\n")
}

func PrintLinksUsage() {
	fmt.Printf("\nLinks Usage\n\n")
	fmt.Printf("	links get\n")
	fmt.Printf("	links get <link-id>\n")
	fmt.Printf("	links post <from> <to>\n")
	fmt.Printf("	links delete <link-id>\n")
}

func PrintModulesUsage() {
	fmt.Printf("\nModules Usage\n\n")
	fmt.Printf("	modules get\n")
	fmt.Printf("	modules get <module-id>\n")
	fmt.Printf("	modules post <module-name>\n")
	fmt.Printf("	modules delete <module-id>\n")
}

func PrintMainLogicUsage() {
	fmt.Printf("\nMainLogic Usage\n\n")
	fmt.Printf("	mainlogic (-v)\n")
	fmt.Printf("	mainlogic switch (<switch-name>)\n")
	fmt.Printf("	mainlogic router (<router-name>)\n")
}

func PrintOvnMonitorUsage() {
	fmt.Printf("\nOvnMonitor Usage\n\n")
	fmt.Printf("	ovnmonitor (-v)\n")
	fmt.Printf("	ovnmonitor switch (<switch-name>)\n")
	fmt.Printf("	ovnmonitor router (<router-name>)\n")
	fmt.Printf("	ovnmonitor interface\n")
}

func PrintHelp() {
	fmt.Printf("\n")
	fmt.Printf("IOVisor-OVN Command Line Interface HELP\n\n")
	fmt.Printf("	interfaces, i    prints /external_interfaces/\n")
	fmt.Printf("	modules, m       prints /modules/\n")
	fmt.Printf("	links, l         prints /links/\n")
	fmt.Printf("	table, t         prints tables\n\n")
	fmt.Printf("	mainlogic, ml    prints mainlogic\n")
	fmt.Printf("	ovnmonitor, ovn  prints ovnmonitor\n\n")

	fmt.Printf("	help, h          print help\n")
	fmt.Printf("\n")
	PrintModulesUsage()
	PrintLinksUsage()
	PrintTableUsage()
	PrintMainLogicUsage()
	PrintOvnMonitorUsage()
}
