package ovnmonitor

type Ovs_Database struct {
	Interface map[string]*Interface_Item //Interface 			name
}

type Interface_Item struct {
	Name         string
	External_Ids map[string]string
}
