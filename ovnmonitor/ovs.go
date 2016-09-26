package ovnmonitor

type Ovs_Database struct {
	Interface map[string]*Interface_Item //Interface 			name
}

type Interface_Item struct {
	Name string
	//we should have to check if it's br-int or not ... in the other logic
	//External_Ids map[string]string
	IfaceId     string
	Up          bool
	IfaceNumber int
	IfaceFd     int
}
