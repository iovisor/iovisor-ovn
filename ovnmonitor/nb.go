package ovnmonitor

type Nb_Database struct {
	Logical_Switch      map[string]*Logical_Switch_Item
	Logical_Switch_Port map[string]*Logical_Switch_Port_Item
}

type Logical_Switch_Item struct {
	Name                 string
	Neutron_Network_Name string
	Ports                map[string]*Logical_Switch_Port_Item
}

type Logical_Switch_Port_Item struct {
	Name string
}
