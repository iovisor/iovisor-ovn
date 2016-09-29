package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/netgroup-polito/iovisor-ovn/hoverctl"
)

func Cli(dataplane *hoverctl.Dataplane) {
	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("cli@iov-ovn$")
		line, _ := reader.ReadString('\n')

		line = TrimSuffix(line, "\n")
		args := strings.Split(line, " ")
		// fmt.Printf("/%+v/\n", args)

		//TODO parse other commands
		if len(args) >= 1 {
			switch args[0] {
			case "test":
				fmt.Printf("test...\n")
				//testenv.TestSwitch2ifc(dataplane, "i:veth1_", "i:veth2_")
			case "interfaces", "i":
				fmt.Printf("Interfaces:\n")
				_, external_interfaces := hoverctl.ExternalInterfacesListGET(dataplane)
				hoverctl.ExternalInterfacesListPrint(external_interfaces)
				// fmt.Printf("%+v\n", external_interfaces)
			case "modules", "m":
				fmt.Printf("Modules:\n")
				_, modules := hoverctl.ModuleListGET(dataplane)
				hoverctl.ModuleListPrint(modules)
				// fmt.Printf("%+v\n", modules)
			case "links", "l":
				fmt.Printf("Links:\n")
				if len(args) >= 2 {
					switch args[1] {
					case "get":
						switch len(args) {
						case 2:
							_, links := hoverctl.LinkListGet(dataplane)
							hoverctl.LinkListPrint(links)
						case 3:
							_, link := hoverctl.LinkGET(dataplane, args[2])
							hoverctl.LinkPrint(link)
						default:
							PrintLinksUsage()
						}
					case "post":
						switch len(args) {
						case 4:
							_, link := hoverctl.LinkPOST(dataplane, args[2], args[3])
							hoverctl.LinkPrint(link)
						default:
							PrintLinksUsage()
						}
					case "delete":
						switch len(args) {
						case 3:
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
				//fmt.Printf("%+v\n", links)
			case "table", "t":
				if len(args) >= 2 {
					switch args[1] {
					case "get":
						switch len(args) {
						case 2:
							fmt.Printf("2-> table get\n")
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
							fmt.Printf("3-> table get <module-id>\n")
							_, tables := hoverctl.TableListGET(dataplane, args[2])
							for _, tablename := range tables {
								fmt.Printf("Table *%s*\n", tablename)
								_, table := hoverctl.TableGET(dataplane, args[2], tablename)
								hoverctl.TablePrint(table)
							}
						case 4:
							fmt.Printf("4-> table get <module-id> <table-id>\n")
							_, table := hoverctl.TableGET(dataplane, args[2], args[3])
							hoverctl.TablePrint(table)
						case 5:
							fmt.Printf("5-> table get <module-id> <table-id> <entry-key>\n")
							_, tableEntry := hoverctl.TableEntryGET(dataplane, args[2], args[3], args[4])
							hoverctl.TableEntryPrint(tableEntry)
						default:
							PrintTableUsage()
						}
					case "put":
						if len(args) == 6 {
							_, tableEntry := hoverctl.TableEntryPUT(dataplane, args[2], args[3], args[4], args[5])
							hoverctl.TableEntryPrint(tableEntry)
						} else {
							PrintTableUsage()
						}
					case "delete":
						if len(args) == 5 {
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
				fmt.Println("invalid command")
				PrintHelp()
			}
		}
	}
}

func TrimSuffix(s, suffix string) string {
	if strings.HasSuffix(s, suffix) {
		s = s[:len(s)-len(suffix)]
	}
	return s
}

func PrintTableUsage() {
	fmt.Printf("Table Usage:\n")
	fmt.Printf("table get\n")
	fmt.Printf("table get <module-id>\n")
	fmt.Printf("table get <module-id> <table-id>\n")
	fmt.Printf("table get <module-id> <table-id> <entry-key>\n")
	fmt.Printf("table put <module-id> <table-id> <entry-key> <entry-value>\n")
	fmt.Printf("table delete <module-id> <table-id> <entry-key> <entry-value>\n")
}

func PrintLinksUsage() {
	fmt.Printf("Links Usage:\n")
	fmt.Printf("links get\n")
	fmt.Printf("links get <link-id>\n")
	fmt.Printf("links delete <link-id>\n")
	fmt.Printf("links post <from> <to>\n")
}

func PrintHelp() {
	fmt.Printf("	interfaces, i    prints /external_interfaces/\n")
	fmt.Printf("	modules, m       prints /modules/\n")
	fmt.Printf("	links, l         prints /links/\n")
	fmt.Printf("	table, t         prints tables\n")
	fmt.Printf("\n	help, h             print help\n")
}
