package ovnmonitor

type Nb_Database struct {
	Logical_Switch      map[string]*Logical_Switch_Item      //Logical_Switch 			name
	Logical_Switch_Port map[string]*Logical_Switch_Port_Item //Logical_Switch_Port name
}

type Logical_Switch_Item struct {
	//OVN fields
	Name      string
	PortsUUID map[string]string

	//Main Logic Fields
	ModuleId   string //module id inside hover
	PortsArray [9]int //[0]=empty [1..8]=contains the port allocation(with fd) for broadcast tricky implemented inside hover
	PortsCount int    //number of allocated ports
	ToRemove   bool   //mark to remove if not present into the OVN_NB_DATABASE, but not yet removed... Other data to remove before
	// Up         bool   //means corresponding module is already posted
}

type Logical_Switch_Port_Item struct {
	//OVN Fiels
	UUID         string
	Name         string
	Addresses    string
	PortSecutiry string

	//Main Logic Fields
	SecurityMacStr     string
	SecurityIpStr      string
	ToRemove           bool
	InterfaceReference string //corresponding interface. If "" emptystring I'm authorized to remove it. After Toremove is marked
}

func (nb *Nb_Database) Clear() {
	nb.Logical_Switch = make(map[string]*Logical_Switch_Item)
	nb.Logical_Switch_Port = make(map[string]*Logical_Switch_Port_Item)
}

func (nb *Nb_Database) ClearLogicalSwitch() {
	nb.Logical_Switch = make(map[string]*Logical_Switch_Item)
}

func (nb *Nb_Database) ClearLogicalSwitchPort() {
	nb.Logical_Switch_Port = make(map[string]*Logical_Switch_Port_Item)
}

func (ls *Logical_Switch_Item) Init() {
	ls.PortsUUID = make(map[string]string)
	ls.ToRemove = false
	// ls.Up = false
}

func (lsp *Logical_Switch_Port_Item) Init() {
	lsp.ToRemove = false
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
