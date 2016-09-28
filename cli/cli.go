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

		if len(args) >= 1 {
			switch args[0] {
			case "test", "t":
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
				_, links := hoverctl.LinkListGet(dataplane)
				hoverctl.LinkListPrint(links)
				//fmt.Printf("%+v\n", links)
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

func PrintHelp() {
	fmt.Printf("	interfaces, i    prints /external_interfaces/\n")
	fmt.Printf("	modules, m       prints /modules/\n")
	fmt.Printf("	links, l         prints /links/\n")
	fmt.Printf("\n	help, h             print help!\n")
}
