package ovnmonitor

import "fmt"

func PrintNb(hh *HandlerHandler) {
	//	fmt.Printf("loop on Logical Switches")
	fmt.Printf("\nLogical Switch\n\n")
	for lsName, _ := range hh.Nb.NbDatabase.Logical_Switch {
		PrintNbLogicalSwitch(hh, lsName)
		fmt.Printf("\n")
	}
	fmt.Printf("\nLogical Switch Ports\n\n")
	for _, lsp := range hh.Nb.NbDatabase.Logical_Switch_Port {
		PrintNbLogicalSwitchPort(hh, lsp.Name)
	}
}

func PrintNbLogicalSwitch(hh *HandlerHandler, logical_switch_name string) {
	ls, ok := hh.Nb.NbDatabase.Logical_Switch[logical_switch_name]
	if ok {
		fmt.Printf("%12s %s\n%12s %s\n%12s %d\n%12s %+v\n", "switchName:", ls.Name, "module-id:", ls.ModuleId, "portsCount:", ls.PortsArray, "portsArray:", ls.PortsCount)
		fmt.Printf("%12s\n", "ports:")
		for k, _ := range ls.PortsUUID {
			fmt.Printf("%12s %20s --> ", "", k)
			PrintNbLogicalSwitchPort(hh, k)
			fmt.Printf("\n")
		}
	}
}

func PrintNbLogicalSwitchPort(hh *HandlerHandler, logical_switch_port_name string) {
	lsp, ok := hh.Nb.NbDatabase.Logical_Switch_Port[logical_switch_port_name]
	if ok {
		fmt.Printf("UUID: %s name: %10s addesses: %+v\n", lsp.UUID, lsp.Name, lsp.Addresses)
	}
}
