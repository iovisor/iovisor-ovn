package ovnmonitor

import "github.com/socketplane/libovsdb"

func OvsLogicInit(h *MonitorHandler) *Ovs_Database {
	//Init Northbound Database Struct
	ovs := Ovs_Database{}
	ovs.Interface = make(map[string]*Interface_Item)
	//Launch goroutine to react to new nb events on buffered channel
	//*h.ovsDatabase = &ovs
	h.OvsDatabase = &ovs
	go OvsLogic(h, &ovs)
	return &ovs
}

func OvsLogic(h *MonitorHandler, ovs *Ovs_Database) {
	//buffered channel, give me the opportunity to get notified on every change in ovnnb
	//but I base my logic on cache-localdbstructs compare, every time
	for {
		select {
		case tableUpdate := <-h.BufupdateOvs:
			//log.Noticef("**NB Notification on Table %s\n", tableUpdate)
			//select on what cache table perform updates
			switch tableUpdate {

			/****LOGICAL_SWITCH***/
			case "Interface":
				//PrintCacheTable(h, tableUpdate)
				//cache lookup
				var cache = *h.Cache

				/*****Interface TABLE********/
				table, _ := cache["Interface"]
				for _, row := range table {

					/*****Interface ITEM********/
					name := row.Fields["name"].(string)
					//external_ids := row.Fields["external_ids"]
					ifaceIDstr := ""

					if extIDs, ok := row.Fields["external_ids"]; ok {
						extIDMap := extIDs.(libovsdb.OvsMap).GoMap
						if ifaceID, ok := extIDMap["iface-id"]; ok {
							ifaceIDstr = ifaceID.(string)
							//PrintTypeDebug(ifaceID)
							//log.Errorf("%s", ifaceID)
							//log.Errorf("%s\n", ifaceID)
						}
					}

					//PrintTypeDebug(external_ids)
					//log.Debugf("external_ids:%s\n", external_ids)

					if iface, ok := ovs.Interface[name]; ok {
						/*****Logical_Switch name PRESENT IN MAP *******/
						//log.Notice("**LS PRESENT IN MAP**")

						iface.Name = name
						//TODO iface.external_ids
						iface.IfaceId = ifaceIDstr
						//PortsToMap(ports, &logicalSwitch.PortsUUID)

						log.Warningf("IF(update):%+v\n", iface)
						h.MainLogicNotification <- "IF (update)"
					} else {
						/*****Logical_Switch name *NOT* PRESENT IN MAP *******/
						//Create Logical_Switch in map
						//log.Notice("**LS /NOT/ PRESENT IN MAP**")

						iface := Interface_Item{}

						iface.Name = name
						//logicalSwitch.PortsUUID = make(map[string]string)
						//PortsToMap(ports, &logicalSwitch.PortsUUID)
						iface.IfaceId = ifaceIDstr
						iface.Up = false

						ovs.Interface[name] = &iface
						log.Warningf("IF(  add ):%+v\n", iface)
						h.MainLogicNotification <- "IF (added)"
					}
				}
			}
		}
	}
}
