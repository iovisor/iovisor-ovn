package ovnmonitor

type Ovs_Database struct {
	Interface map[string]*Interface_Item //Interface 			name
}

type Interface_Item struct {
	//OVN/OVS fields
	Name               string
	IfaceIdExternalIds string //Name of the Interface in External_Ids

	//Main Logic Fields
	IfaceIdRedirectHover  int    //Iface id inside hover (relative to the m:1234 the interface is attached to ...) and provided my the extended hover /links/ API
	IfaceIdArrayBroadcast int    //Interface Id in the array for broadcast (id->fd for broadcast)
	IfaceFd               int    //Interface Fd inside External_Ids (42, etc...)
	Up                    bool   //Up means corresponding (module??) link is already POSTed
	LinkIdHover           string //iomodules Link Id
	ToRemove              bool   //To remove flag
	SecurityMacString     string //if the security policy based on mac in already injected into the Iomodules, contains the string of the leaf in the table
}

func (ovs *Ovs_Database) Clear() {
	ovs.Interface = make(map[string]*Interface_Item)
}

func (iface *Interface_Item) Init() {
	iface.Up = false
	iface.ToRemove = false
}
