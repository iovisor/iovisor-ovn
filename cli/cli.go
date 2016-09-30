// Inline command line interface for debug purposes
// in future this cli will be a separate go program that connects to the main iovisor-ovn daemon
// in future this cli will use a cli go library (e.g. github.com/urfave/cli )

package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/netgroup-polito/iovisor-ovn/bpf"
	"github.com/netgroup-polito/iovisor-ovn/hoverctl"
	"github.com/netgroup-polito/iovisor-ovn/ovnmonitor"
	"github.com/netgroup-polito/iovisor-ovn/testenv"
)

func Cli(hh *ovnmonitor.HandlerHandler) {
	dataplane := hh.Dataplane
	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("cli@iov-ovn$")
		line, _ := reader.ReadString('\n')

		line = TrimSuffix(line, "\n")
		args := strings.Split(line, " ")

		if len(args) >= 1 {
			switch args[0] {
			case "test":
				fmt.Printf("\ntest\n\n")
				testenv.TestLinkPostDelete(dataplane)
			case "nb":
				if len(args) >= 1 {
					if len(args) == 1 {
						ovnmonitor.PrintNb(hh)
					} else {
						if len(args) == 3 {
							switch args[2] {
							case "ls":
								ovnmonitor.PrintNbLogicalSwitch(hh, args[2])
							case "lsp":
								ovnmonitor.PrintNbLogicalSwitchPort(hh, args[2])
							default:
								PrintNbUsage()
							}
						} else {
							PrintNbUsage()
						}
					}
				} else {
					PrintNbUsage()
				}
				fmt.Printf("\nNorthBound DB\n\n")
				ovnmonitor.PrintNb(hh)
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
								_, module := hoverctl.ModulePOST(dataplane, "bpf", "Switch", bpf.Switch)
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
				fmt.Println("\nInvalid Command\n\n")
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

func PrintNbUsage() {
	fmt.Printf("\nNB Usage\n\n")
	fmt.Printf("	nb                  print the whole NorthBound\n")
	fmt.Printf("	nb ls   <ls-name>   print the Logical Switch table\n")
	fmt.Printf("	nb lsp  <lsp-name>  print the Logical Switch Port table\n")
}

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

func PrintHelp() {
	fmt.Printf("\n")
	fmt.Printf("IOVisor-OVN Command Line Interface HELP\n\n")
	fmt.Printf("	interfaces, i    prints /external_interfaces/\n")
	fmt.Printf("	modules, m       prints /modules/\n")
	fmt.Printf("	links, l         prints /links/\n")
	fmt.Printf("	table, t         prints tables\n\n")
	fmt.Printf("	nb               prints NorthBound database local structs\n\n")
	fmt.Printf("	help, h          print help\n")
	fmt.Printf("\n")
	PrintModulesUsage()
	PrintLinksUsage()
	PrintTableUsage()
	PrintNbUsage()
}
