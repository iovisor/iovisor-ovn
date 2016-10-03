package ovnmonitor

type Ovs_Database struct {
	Interface map[string]*Interface_Item //Interface 			name
}

type Interface_Item struct {
	Name string
	//we should have to check if it's br-int or not ... in the other logic
	//External_Ids map[string]string
	IfaceIdExternalIds string //Name of the Interface in External_Ids
	IfaceNumber        int    //Interface id inside hover (1,2,3....)
	IfaceFd            int    //Interface Fd inside External_Ids (42, etc...)
	Up                 bool   //Up means corresponding module is already POSTed
	LinkId             string //iomodules Link Id
	ToRemove           bool   //To remove flag
}
