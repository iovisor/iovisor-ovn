package ovnmonitor

type Nb_Database struct {
	Logical_Switch      map[string]*Logical_Switch_Item      //Logical_Switch 			name
	Logical_Switch_Port map[string]*Logical_Switch_Port_Item //Logical_Switch_Port name
}

type Logical_Switch_Item struct {
	Name       string
	PortsUUID  map[string]string
	ModuleId   string
	PortsArray [9]int
	PortsCount int
	//Enabled   bool
}

type Logical_Switch_Port_Item struct {
	UUID      string
	Name      string
	Addresses []string
}

//return the name of the switch a port belongs to
func PortLookup(nb *Nb_Database, portName string) string {

	if port, ok := nb.Logical_Switch_Port[portName]; ok {
		uuid := port.UUID
		for _, lsptr := range nb.Logical_Switch {
			ls := *lsptr
			if _, ok := ls.PortsUUID[uuid]; ok {
				return ls.Name
			}
		}
	}
	return ""
}

func FindFirtsFreeLogicalPort(logicalSwitch *Logical_Switch_Item) int {
	for i := 1; i < 9; i++ {
		if logicalSwitch.PortsArray[i] == 0 {
			return i
		}
	}
	return 0
}

//Immediate Lookup @ Logical_Switch name
//Immediate Lookup @ Logical_Switch_Port (that contains reference to )
