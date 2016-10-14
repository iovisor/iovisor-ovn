package ovnmonitor

type Ovs_Database struct {
	Interface map[string]*Interface_Item //Interface 			name
}

type Interface_Item struct {
	Name string
	//we should have to check if it's br-int or not ... in the other logic
	//External_Ids map[string]string
	IfaceIdExternalIds string //Name of the Interface in External_Ids
	IfaceNumberHover   int    //Iface number inside hover (relative to the m:1234 the interface is attached to ...)
	IfaceNumber        int    //Interface id inside hover (1,2,3....)?? Not sure.. probably iface number for broadcast inside hover
	IfaceFd            int    //Interface Fd inside External_Ids (42, etc...)
	Up                 bool   //Up means corresponding module is already POSTed
	LinkId             string //iomodules Link Id
	ToRemove           bool   //To remove flag
	SecurityMacString  string //if the security policy based on mac in already injected into the Iomodules, contains the string of the leaf in the table
}
