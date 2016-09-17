package ovnmonitor

func OvsLogicInit(h *MonitorHandler) *Ovs_Database {
	//Init Northbound Database Struct
	ovs := Ovs_Database{}
	ovs.Interface = make(map[string]*Interface_Item)
	//Launch goroutine to react to new nb events on buffered channel
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
					external_ids := row.Fields["external_ids"]
					log.Warningf("external_ids:%s\n", external_ids)

					if iface, ok := ovs.Interface[name]; ok {
						/*****Logical_Switch name PRESENT IN MAP *******/
						//log.Notice("**LS PRESENT IN MAP**")

						iface.Name = name
						//TODO iface.external_ids
						//PortsToMap(ports, &logicalSwitch.PortsUUID)

						log.Warningf("IF(update):%+v\n", iface)
					} else {
						/*****Logical_Switch name *NOT* PRESENT IN MAP *******/
						//Create Logical_Switch in map
						//log.Notice("**LS /NOT/ PRESENT IN MAP**")

						iface := Interface_Item{}

						iface.Name = name
						//logicalSwitch.PortsUUID = make(map[string]string)
						//PortsToMap(ports, &logicalSwitch.PortsUUID)

						ovs.Interface[name] = &iface
						log.Warningf("IF(  add ):%+v\n", iface)
					}
				}
			}
		}
	}
}
