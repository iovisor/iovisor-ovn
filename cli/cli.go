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

	"github.com/iovisor/iovisor-ovn/hover"
	"github.com/iovisor/iovisor-ovn/iomodules/l2switch"
	"github.com/iovisor/iovisor-ovn/mainlogic"
	"github.com/iovisor/iovisor-ovn/ovnmonitor"
)

func Cli(c *hover.Client) {
	db := &mainlogic.Mon.DB
	hc := c
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
				_, external_interfaces := hc.ExternalInterfacesListGET()
				hover.ExternalInterfacesListPrint(external_interfaces)
			case "modules", "m":
				if len(args) >= 2 {
					switch args[1] {
					case "get":
						switch len(args) {
						case 2:
							fmt.Printf("\nModules GET\n\n")
							_, modules := hc.ModuleListGET()
							hover.ModuleListPrint(modules)
						case 3:
							fmt.Printf("\nModules GET\n\n")
							_, module := hc.ModuleGET(args[2])
							hover.ModulePrint(module)
						default:
							PrintModulesUsage()
						}
					case "post":
						switch len(args) {
						case 3:
							fmt.Printf("\nModules POST\n\n")
							if args[2] == "switch" {
								_, module := hc.ModulePOST("bpf", "Switch", l2switch.SwitchSecurityPolicy)
								hover.ModulePrint(module)
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
							hc.ModuleDELETE(args[2])
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
							_, links := hc.LinkListGet()
							hover.LinkListPrint(links)
						case 3:
							fmt.Printf("\nLinks GET\n\n")
							_, link := hc.LinkGET(args[2])
							hover.LinkPrint(link)
						default:
							PrintLinksUsage()
						}
					case "post":
						switch len(args) {
						case 4:
							fmt.Printf("\nLinks POST\n\n")
							_, link := hc.LinkPOST(args[2], args[3])
							hover.LinkPrint(link)
						default:
							PrintLinksUsage()
						}
					case "delete":
						switch len(args) {
						case 3:
							fmt.Printf("\nLinks DELETE\n\n")
							hc.LinkDELETE(args[2])
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
							_, modules := hc.ModuleListGET()
							for moduleName, _ := range modules {
								fmt.Printf("**MODULE** -> %s\n", moduleName)
								_, tables := hc.TableListGET(moduleName)
								for _, tablename := range tables {
									fmt.Printf("Table *%s*\n", tablename)
									_, table := hc.TableGET(moduleName, tablename)
									hover.TablePrint(table)
								}
							}
						case 3:
							fmt.Printf("\nTable GET\n\n")
							_, tables := hc.TableListGET(args[2])
							for _, tablename := range tables {
								fmt.Printf("Table *%s*\n", tablename)
								_, table := hc.TableGET(args[2], tablename)
								hover.TablePrint(table)
							}
						case 4:
							fmt.Printf("\nTable GET\n\n")
							_, table := hc.TableGET(args[2], args[3])
							hover.TablePrint(table)
						case 5:
							fmt.Printf("\nTable GET\n\n")
							_, tableEntry := hc.TableEntryGET(args[2], args[3], args[4])
							hover.TableEntryPrint(tableEntry)
						default:
							PrintTableUsage()
						}
					case "put":
						if len(args) == 6 {
							fmt.Printf("\nTable PUT\n\n")
							_, tableEntry := hc.TableEntryPUT(args[2], args[3], args[4], args[5])
							hover.TableEntryPrint(tableEntry)
						} else {
							PrintTableUsage()
						}
					case "post":
						if len(args) == 6 {
							fmt.Printf("\nTable POST\n\n")
							_, tableEntry := hc.TableEntryPOST(args[2], args[3], args[4], args[5])
							hover.TableEntryPrint(tableEntry)
						} else {
							PrintTableUsage()
						}
					case "delete":
						if len(args) == 5 {
							fmt.Printf("\nTable DELETE\n\n")
							hc.TableEntryDELETE(args[2], args[3], args[4])
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
