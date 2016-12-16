package main

import (
	"net"
	"fmt"
	)



func getInterfaceMac(iface string) string {

	ifc, err := net.InterfaceByName(iface)
	if err != nil {
		return ""
	}

	return ifc.HardwareAddr.String()
}


func main() {
	fmt.Println(getInterfaceMac("veth1"))
}
