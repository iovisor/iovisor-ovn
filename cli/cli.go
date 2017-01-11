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
// Inline command line interface for debug purposes
// in future this cli will be a separate go program that connects to the main iovisor-ovn daemon
// in future this cli will use a cli go library (e.g. github.com/urfave/cli )

package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/netgroup-polito/iovisor-ovn/hoverctl"
	"github.com/netgroup-polito/iovisor-ovn/iomodules/l2switch"
	"github.com/netgroup-polito/iovisor-ovn/mainlogic"
	"github.com/netgroup-polito/iovisor-ovn/ovnmonitor"
)

func Cli(dataplaneref *hoverctl.Dataplane) {
	db := &mainlogic.Mon.DB
	dataplane := dataplaneref
	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("cli@iov-ovn$ ")
		line, _ := reader.ReadString('\n')

		line = TrimSuffix(line, "\n")
		args := strings.Split(line, " ")

		if len(args) >= 1 {
			switch args[0] {
			case "mainlogic", "ml":
				if len(args) >= 2 {
					switch args[1] {
					case "-v":
						fmt.Printf("\nMainLogic (verbose)\n\n")
						mainlogic.PrintMainLogic(true)
					case "switch":
						if len(args) >= 3 {
							fmt.Printf("\nMainLogic Switch %s\n\n", args[2])
							mainlogic.PrintL2Switch(args[2])
						} else {
							fmt.Printf("\nMainLogic Switches \n\n")
							mainlogic.PrintL2Switches(true)
						}
					case "router":
						if len(args) >= 3 {
							fmt.Printf("\nMainLogic Router %s\n\n", args[2])
							mainlogic.PrintRouter(args[2])
						} else {
							fmt.Printf("\nMainLogic Routers \n\n")
							mainlogic.PrintRouters(true)
						}
					}
				} else {
					fmt.Printf("\nMainLogic\n\n")
					mainlogic.PrintMainLogic(false)
				}

			case "ovnmonitor", "ovn", "o":
				if len(args) >= 2 {
					switch args[1] {
					case "-v":
						fmt.Printf("\nOvnMonitor (verbose)\n\n")
						ovnmonitor.PrintOvnMonitor(true, db)
					case "switch", "s":
						if len(args) >= 3 {
							fmt.Printf("\nOvnMonitor Logical Switch %s\n\n", args[2])
							ovnmonitor.PrintLogicalSwitchByName(args[2], db)
						} else {
							fmt.Printf("\nOvnMonitor Logical Switches \n\n")
							ovnmonitor.PrintLogicalSwitches(true, db)
						}
					case "router", "r":
						if len(args) >= 3 {
							fmt.Printf("\nOvnMonitor Logical Router %s\n\n", args[2])
							ovnmonitor.PrintLogicalRouterByName(args[2], db)
						} else {
							fmt.Printf("\nOvnMonitor Logical Routers \n\n")
							ovnmonitor.PrintLogicalRouters(true, db)
						}
					case "interface", "i":
						fmt.Printf("\nOvnMonitor Ovs Interfaces\n\n")
						ovnmonitor.PrintOvsInterfaces(db)
					}
				} else {
					fmt.Printf("\nOvn Monitor\n\n")
					ovnmonitor.PrintOvnMonitor(false, db)
				}

			case "interfaces", "i":
				fmt.Printf("\nInterfaces\n\n")
				_, external_interfaces := hoverctl.ExternalInterfacesListGET(dataplane)
				hoverctl.ExternalInterfacesListPrint(external_interfaces)
			case "modules", "m":
				if len(args) >= 2 {
					switch args[1] {
					case "get":
						switch len(args) {
						case 2:
							fmt.Printf("\nModules GET\n\n")
							_, modules := hoverctl.ModuleListGET(dataplane)
							hoverctl.ModuleListPrint(modules)
						case 3:
							fmt.Printf("\nModules GET\n\n")
							_, module := hoverctl.ModuleGET(dataplane, args[2])
							hoverctl.ModulePrint(module)
						default:
							PrintModulesUsage()
						}
					case "post":
						switch len(args) {
						case 3:
							fmt.Printf("\nModules POST\n\n")
							if args[2] == "switch" {
								_, module := hoverctl.ModulePOST(dataplane, "bpf", "Switch", l2switch.SwitchSecurityPolicy)
								hoverctl.ModulePrint(module)
							} else {
								//TODO Print modules list
							}
						default:
							PrintModulesUsage()
						}
					case "delete":
						switch len(args) {
						case 3:
							fmt.Printf("\nModules DELETE\n\n")
							hoverctl.ModuleDELETE(dataplane, args[2])
						default:
							PrintModulesUsage()
						}
					default:
						PrintModulesUsage()
					}
				} else {
					PrintModulesUsage()
				}
			case "links", "l":
				if len(args) >= 2 {
					switch args[1] {
					case "get":
						switch len(args) {
						case 2:
							fmt.Printf("\nLinks GET\n\n")
							_, links := hoverctl.LinkListGet(dataplane)
							hoverctl.LinkListPrint(links)
						case 3:
							fmt.Printf("\nLinks GET\n\n")
							_, link := hoverctl.LinkGET(dataplane, args[2])
							hoverctl.LinkPrint(link)
						default:
							PrintLinksUsage()
						}
					case "post":
						switch len(args) {
						case 4:
							fmt.Printf("\nLinks POST\n\n")
							_, link := hoverctl.LinkPOST(dataplane, args[2], args[3])
							hoverctl.LinkPrint(link)
						default:
							PrintLinksUsage()
						}
					case "delete":
						switch len(args) {
						case 3:
							fmt.Printf("\nLinks DELETE\n\n")
							hoverctl.LinkDELETE(dataplane, args[2])
						default:
							PrintLinksUsage()
						}
					default:
						PrintLinksUsage()
					}
				} else {
					PrintLinksUsage()
				}
			case "table", "t":
				if len(args) >= 2 {
					switch args[1] {
					case "get":
						switch len(args) {
						case 2:
							fmt.Printf("\nTable GET\n\n")
							_, modules := hoverctl.ModuleListGET(dataplane)
							for moduleName, _ := range modules {
								fmt.Printf("**MODULE** -> %s\n", moduleName)
								_, tables := hoverctl.TableListGET(dataplane, moduleName)
								for _, tablename := range tables {
									fmt.Printf("Table *%s*\n", tablename)
									_, table := hoverctl.TableGET(dataplane, moduleName, tablename)
									hoverctl.TablePrint(table)
								}
							}
						case 3:
							fmt.Printf("\nTable GET\n\n")
							_, tables := hoverctl.TableListGET(dataplane, args[2])
							for _, tablename := range tables {
								fmt.Printf("Table *%s*\n", tablename)
								_, table := hoverctl.TableGET(dataplane, args[2], tablename)
								hoverctl.TablePrint(table)
							}
						case 4:
							fmt.Printf("\nTable GET\n\n")
							_, table := hoverctl.TableGET(dataplane, args[2], args[3])
							hoverctl.TablePrint(table)
						case 5:
							fmt.Printf("\nTable GET\n\n")
							_, tableEntry := hoverctl.TableEntryGET(dataplane, args[2], args[3], args[4])
							hoverctl.TableEntryPrint(tableEntry)
						default:
							PrintTableUsage()
						}
					case "put":
						if len(args) == 6 {
							fmt.Printf("\nTable PUT\n\n")
							_, tableEntry := hoverctl.TableEntryPUT(dataplane, args[2], args[3], args[4], args[5])
							hoverctl.TableEntryPrint(tableEntry)
						} else {
							PrintTableUsage()
						}
					case "post":
						if len(args) == 6 {
							fmt.Printf("\nTable POST\n\n")
							_, tableEntry := hoverctl.TableEntryPOST(dataplane, args[2], args[3], args[4], args[5])
							hoverctl.TableEntryPrint(tableEntry)
						} else {
							PrintTableUsage()
						}
					case "delete":
						if len(args) == 5 {
							fmt.Printf("\nTable DELETE\n\n")
							hoverctl.TableEntryDELETE(dataplane, args[2], args[3], args[4])
						} else {
							PrintTableUsage()
						}
					default:
						PrintTableUsage()
					}
				} else {
					PrintTableUsage()
				}
			case "help", "h":
				PrintHelp()

			case "":
			default:
				fmt.Printf("\nInvalid Command\n\n")
				PrintHelp()
			}
		}
		fmt.Printf("\n")
	}
}

func TrimSuffix(s, suffix string) string {
	if strings.HasSuffix(s, suffix) {
		s = s[:len(s)-len(suffix)]
	}
	return s
}
