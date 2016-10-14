package ovnmonitor

import (
	"bytes"
	"errors"
	"reflect"
	"strings"

	"github.com/socketplane/libovsdb"
)

//Idea: new notification triggers this method.
//But the logic is based on cache
//We compare cache with local db structs!
//And we can triggers new IOVisor events

//Idea #2
//Only events on ovs triggers something
//Only when a physical port is attached to ovs we have to react

//Ignore:
//We ignore time between port attact in ovs (only mantained for compatibility)
//And the time when IOVisor physically connects this port!

//In general, ovn nb db state -> tells to cms when port is up!
//So if something went wrong in IOVisor-ovn we have to set it down!

//TODO initial tables
//TODO delete items

func NbLogicInit(h *MonitorHandler) *Nb_Database {
	//Init Northbound Database Struct
	nb := Nb_Database{}
	nb.Logical_Switch = make(map[string]*Logical_Switch_Item)
	nb.Logical_Switch_Port = make(map[string]*Logical_Switch_Port_Item)
	//Launch goroutine to react to new nb events on buffered channel
	//	*h.NbDatabase = &nb
	h.NbDatabase = &nb
	go NbLogic(h, &nb)
	return &nb
}

func NbLogic(h *MonitorHandler, nb *Nb_Database) {
	//buffered channel, give me the opportunity to get notified on every change in ovnnb
	//but I base my logic on cache-localdbstructs compare, every time
	for {
		select {
		case tableUpdate := <-h.Bufupdate:
			//log.Noticef("**NB Notification on Table %s\n", tableUpdate)
			//select on what cache table perform updates
			switch tableUpdate {

			/****LOGICAL_SWITCH***/
			case "Logical_Switch":
				//PrintCacheTable(h, tableUpdate)
				//cache lookup
				var cache = *h.Cache

				/*****Logical_Switch TABLE********/
				table, _ := cache["Logical_Switch"]
				for _, row := range table {

					/*****Logical_Switch ITEM********/
					name := row.Fields["name"].(string)
					ports := row.Fields["ports"]

					if logicalSwitch, ok := nb.Logical_Switch[name]; ok {
						/*****Logical_Switch name PRESENT IN MAP *******/
						//log.Notice("**LS PRESENT IN MAP**")

						logicalSwitch.Name = name
						PortsToMap(ports, &logicalSwitch.PortsUUID)

						//log.Noticef("LS(update):%+v\n", logicalSwitch)
						//h.MainLogicNotification <- "LS update"
					} else {
						/*****Logical_Switch name *NOT* PRESENT IN MAP *******/
						//Create Logical_Switch in map
						//log.Notice("**LS /NOT/ PRESENT IN MAP**")

						logicalSwitch := Logical_Switch_Item{}

						logicalSwitch.Name = name
						logicalSwitch.PortsUUID = make(map[string]string)
						PortsToMap(ports, &logicalSwitch.PortsUUID)

						nb.Logical_Switch[name] = &logicalSwitch
						log.Noticef("LogicalSwitch ADDED:\n%+v\n", logicalSwitch)
						//h.MainLogicNotification <- "LS add"
					}
				}

			case "Logical_Switch_Port":
				//log.Notice("Logical_Switch_Port")
				//PrintCacheTable(h, tableUpdate)

				//cache lookup
				var cache = *h.Cache

				/*****Logical_Switch_Port TABLE********/
				table, _ := cache["Logical_Switch_Port"]
				for uuid, row := range table {

					/*****Logical_Switch_Port ITEM********/
					name := row.Fields["name"].(string)
					addresses := row.Fields["addresses"]
					port_security := row.Fields["port_security"]

					if logicalSwitchPort, ok := nb.Logical_Switch_Port[name]; ok {
						/*****Logical_Switch_Port name PRESENT IN MAP *******/
						//log.Notice("**LS-PORT PRESENT IN MAP**")

						logicalSwitchPort.Name = name
						logicalSwitchPort.UUID = uuid
						//Addresses to slice
						logicalSwitchPort.Addresses = InterfaceToString(addresses)
						// if logicalSwitchPort.Addresses != "" {
						// 	log.Debugf("logicalSwitchPort: %s Addresses: %s\n", logicalSwitchPort.Name, logicalSwitchPort.Addresses)
						// }
						//PrintTypeDebug(port_security)
						logicalSwitchPort.PortSecutiry = InterfaceToString(port_security)
						if logicalSwitchPort.PortSecutiry != "" {
							logicalSwitchPort.SecurityMacStr = FromPortSecurityStrToMacStr(logicalSwitchPort.PortSecutiry)
							// log.Noticef("MAC:%s\n", logicalSwitchPort.SecurityMacStr)
						}

						// if logicalSwitchPort.PortSecutiry != "" {
						// 	log.Debugf("logicalSwitchPort: %s PortSecutiry: %s\n", logicalSwitchPort.Name, logicalSwitchPort.PortSecutiry)
						// }

						//PortsToMap(ports, &logicalSwitch.Ports)

						//log.Noticef("LP(update):%+v\n", logicalSwitchPort)
						//h.MainLogicNotification <- "LP update"
					} else {
						/*****Logical_Switch_Port name *NOT* PRESENT IN MAP *******/
						//Create Logical_Switch_Port in map
						//log.Notice("**LS-PORT /NOT/ PRESENT IN MAP**")

						logicalSwitchPort := Logical_Switch_Port_Item{}

						logicalSwitchPort.Name = name
						logicalSwitchPort.UUID = uuid
						//logicalSwitchPort.Addresses =// make(map[string]string)
						//PortsToMap(ports, &logicalSwitch.Ports)
						//Addresses to slice
						//PrintTypeDebug(addresses)
						logicalSwitchPort.Addresses = InterfaceToString(addresses)
						// if logicalSwitchPort.Addresses != "" {
						// 	log.Debugf("logicalSwitchPort: %s Addresses: %s\n", logicalSwitchPort.Name, logicalSwitchPort.Addresses)
						// }
						//PrintTypeDebug(port_security)
						logicalSwitchPort.PortSecutiry = InterfaceToString(port_security)
						if logicalSwitchPort.PortSecutiry != "" {
							logicalSwitchPort.SecurityMacStr = FromPortSecurityStrToMacStr(logicalSwitchPort.PortSecutiry)
							// log.Noticef("MAC:%s\n", logicalSwitchPort.SecurityMacStr)
						}
						// if logicalSwitchPort.PortSecutiry != "" {
						// 	log.Debugf("logicalSwitchPort: %s PortSecutiry: %s\n", logicalSwitchPort.Name, logicalSwitchPort.PortSecutiry)
						// }

						//logicalSwitchPort.Addresses = ovsStringSetToSlice(addresses)

						nb.Logical_Switch_Port[name] = &logicalSwitchPort
						log.Noticef("Logical Port ADDED:\n%+v\n", logicalSwitchPort)
						//h.MainLogicNotification <- "LP add"
					}

					//log.Debugf("Port Lookup: %s -> switch: %s\n", name, PortLookup(nb, name))
				}
			}
		}
	}
}

//Find the first mac and converts it into exadecimal string
//Eg:
//input str -> "192.168.1.1 08:00:27:2a:03:54"
//output str -> 0x0800272a0354
func FromPortSecurityStrToMacStr(portSecurity string) string {
	//divide portSecurity into slices " "
	slices := strings.Split(portSecurity, " ")
	for _, slice := range slices {
		//log.Debugf("slice: %s\n", slice)
		//find first mac
		if IsMac(slice) {
			//modify format
			return MacToExadecimalString(slice)
		}
	}
	return ""
}

func IsAllowedChar(s string) bool {
	for _, r := range s {
		if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F') || (r == ':')) {
			return false
		}
	}
	return true
}

//01:34:67:90:23:56
func IsMac(s string) bool {
	if len(s) != 17 {
		return false
	}
	if s[2] != ':' || s[5] != ':' || s[8] != ':' || s[11] != ':' || s[14] != ':' {
		return false
	}
	if !IsAllowedChar(s) {
		return false
	}
	return true
}

func MacToExadecimalString(s string) string {
	var buffer bytes.Buffer

	buffer.WriteString("0x")
	buffer.WriteString(s[0:2])
	buffer.WriteString(s[3:5])
	buffer.WriteString(s[6:8])
	buffer.WriteString(s[9:11])
	buffer.WriteString(s[12:14])
	buffer.WriteString(s[15:17])

	return buffer.String()
}

func PortsToMap(ports interface{}, Ports *map[string]string) {

	portsMap := *Ports
	switch ports.(type) {
	case libovsdb.UUID:
		portsMap[ports.(libovsdb.UUID).GoUUID] = ports.(libovsdb.UUID).GoUUID
	case libovsdb.OvsSet:
		for _, uuids := range ports.(libovsdb.OvsSet).GoSet {
			portsMap[uuids.(libovsdb.UUID).GoUUID] = uuids.(libovsdb.UUID).GoUUID
		}
	}
}

func PrintTypeDebug(i interface{}) {
	log.Debugf("ports: %s\n", reflect.TypeOf(i))
	log.Debugf("ports: %s\n", i)
}

func InterfaceToString(i interface{}) string {

	switch i.(type) {
	case string:
		return i.(string)
	}
	return ""
}

func ovsStringMapToMap(oMap interface{}) (map[string]string, error) {
	var ret = make(map[string]string)
	wrap, ok := oMap.([]interface{})
	if !ok {
		return nil, errors.New("ovs map outermost layer invalid")
	}
	if wrap[0] != "map" {
		return nil, errors.New("ovs map invalid identifier")
	}

	brokenMap, ok := wrap[1].([]interface{})
	if !ok {
		return nil, errors.New("ovs map content invalid")
	}
	for _, kvPair := range brokenMap {
		kvSlice, ok := kvPair.([]interface{})
		if !ok {
			return nil, errors.New("ovs map block must be a slice")
		}
		key, ok := kvSlice[0].(string)
		if !ok {
			return nil, errors.New("ovs map key must be string")
		}
		val, ok := kvSlice[1].(string)
		if !ok {
			return nil, errors.New("ovs map value must be string")
		}
		ret[key] = val
	}
	return ret, nil
}

func ovsStringSetToSlice(oSet interface{}) []string {
	var ret []string
	if t, ok := oSet.([]interface{}); ok && t[0] == "set" {
		for _, v := range t[1].([]interface{}) {
			ret = append(ret, v.(string))
		}
	} else {
		ret = append(ret, oSet.(string))
	}
	return ret
}
