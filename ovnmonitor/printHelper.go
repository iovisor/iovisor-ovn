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
		fmt.Printf("%12s %s\n%12s %s\n%12s %d\n%12s %+v\n%12s %t\n", "SwitchName:", ls.Name, "ModuleId:", ls.ModuleId, "PortsArray:", ls.PortsArray, "PortsCount:", ls.PortsCount, "ToRemove:", ls.ToRemove)
		fmt.Printf("%12s\n", "portsUUID:")
		for k, _ := range ls.PortsUUID {
			fmt.Printf("%12s %20s", "", k)
			PrintNbLogicalSwitchPort(hh, k)
			fmt.Printf("\n")
		}
	}
}

func PrintNbLogicalSwitchPort(hh *HandlerHandler, logical_switch_port_name string) {
	lsp, ok := hh.Nb.NbDatabase.Logical_Switch_Port[logical_switch_port_name]
	if ok {
		fmt.Printf("%12s %s\n%12s %s\n%12s %s\n%12s %s\n", "UUID:", lsp.UUID, "Name:", lsp.Name, "Addresses:", lsp.Addresses, "PortSecutiry:", lsp.PortSecutiry)
		fmt.Printf("%12s %s\n%12s %s\n%12s %s\n%12s %s\n%12s %t\n\n", "SecurityMac:", lsp.SecurityMacStr, "SecurityIp:", lsp.SecurityIpStr, "LSwitchName:", lsp.LogicalSwitchName, "IfaceRef:", lsp.InterfaceReference, "ToRemove:", lsp.ToRemove)
	}
}

func PrintOvs(hh *HandlerHandler) {
	//	fmt.Printf("loop on Logical Switches")
	fmt.Printf("\nInterfaces\n\n")
	for ifaceName, _ := range hh.Ovs.OvsDatabase.Interface {
		PrintOvsInterface(hh, ifaceName)
		fmt.Printf("\n\n")
	}
}

func PrintOvsInterface(hh *HandlerHandler, interface_name string) {
	iface, ok := hh.Ovs.OvsDatabase.Interface[interface_name]
	if ok {
		fmt.Printf("%21s: %s\n%21s: %s\n%21s: %s\n%21s: %d\n%21s: %d\n%21s: %d\n", "*Name*", iface.Name, "IfaceIdExternalIds", iface.IfaceIdExternalIds, "LinkIdHover", iface.LinkIdHover, "IfaceIdRedirectHover", iface.IfaceIdRedirectHover, "IfaceIdArrayBroadcast", iface.IfaceIdArrayBroadcast, "IfaceFd", iface.IfaceFd)

		fmt.Printf("%21s: %s\n%21s: %s\n%21s: %t\n%21s: %t", "SecurityMacString", iface.SecurityMacString, "SecurityIpString", iface.SecurityIpString, "Up", iface.Up, "ToRemove", iface.ToRemove)
	}
}
