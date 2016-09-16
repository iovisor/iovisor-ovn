package ovnmonitor

import (
	"errors"
	"reflect"

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

						log.Warningf("LS(update):%+v\n", logicalSwitch)
					} else {
						/*****Logical_Switch name *NOT* PRESENT IN MAP *******/
						//Create Logical_Switch in map
						//log.Notice("**LS /NOT/ PRESENT IN MAP**")

						logicalSwitch := Logical_Switch_Item{}

						logicalSwitch.Name = name
						logicalSwitch.PortsUUID = make(map[string]string)
						PortsToMap(ports, &logicalSwitch.PortsUUID)

						nb.Logical_Switch[name] = &logicalSwitch
						log.Warningf("LS(  add ):%+v\n", logicalSwitch)
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

					if logicalSwitchPort, ok := nb.Logical_Switch_Port[name]; ok {
						/*****Logical_Switch_Port name PRESENT IN MAP *******/
						//log.Notice("**LS-PORT PRESENT IN MAP**")

						logicalSwitchPort.Name = name
						logicalSwitchPort.UUID = uuid
						//Addresses to slice
						PrintTypeDebug(addresses)

						//PortsToMap(ports, &logicalSwitch.Ports)

						log.Warningf("LP(update):%+v\n", logicalSwitchPort)
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
						PrintTypeDebug(addresses)
						//logicalSwitchPort.Addresses = ovsStringSetToSlice(addresses)

						nb.Logical_Switch_Port[name] = &logicalSwitchPort
						log.Warningf("LP(  add ):%+v\n", logicalSwitchPort)
					}

					log.Debugf("Port Lookup: %s -> switch: %s\n", name, PortLookup(nb, name))
				}
			}
		}
	}
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
