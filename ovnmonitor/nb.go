package ovnmonitor

type Nb_Database struct {
	Logical_Switch      map[string]*Logical_Switch_Item      //Logical_Switch 			name
	Logical_Switch_Port map[string]*Logical_Switch_Port_Item //Logical_Switch_Port name
}

type Logical_Switch_Item struct {
	Name      string
	PortsUUID map[string]string
}

type Logical_Switch_Port_Item struct {
	UUID      string
	Name      string
	Addresses []string
}

//return the name of the switch a port belongs to
func PortLookup(nb *Nb_Database, portName string) string {
	//TODO implement
	uuid := nb.Logical_Switch_Port[portName].UUID
	for _, lsptr := range nb.Logical_Switch {
		ls := *lsptr
		if _, ok := ls.PortsUUID[uuid]; ok {
			return ls.Name
		}
	}
	return ""
}

//Immediate Lookup @ Logical_Switch name
//Immediate Lookup @ Logical_Switch_Port (that contains reference to )
