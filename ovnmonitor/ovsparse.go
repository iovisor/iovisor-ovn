package ovnmonitor

import "github.com/socketplane/libovsdb"

func OvsParseInit(h *MonitorHandler) *Ovs_Database {
	//Init Northbound Database Struct
	ovs := Ovs_Database{}
	// ovs.Interface = make(map[string]*Interface_Item)
	ovs.Clear()
	//Launch goroutine to react to new nb events on buffered channel
	//*h.ovsDatabase = &ovs
	h.OvsNewDatabase = &ovs
	go OvsParse(h, &ovs)
	return &ovs
}

//This Method is in charge to convert the libovsdb struct into my struct representation
//Cache (in ovsdblib format) -> Cache (in my format: struct OvsDatabase)
//
//h.BufupdateOvs is a buffered channel
//that notify me every time there is a change in a table (with the table name)
//
//The main logic is moved into /mainlogic/*
//
//Notify the main logic @ each change in the Database
func OvsParse(h *MonitorHandler, ovs *Ovs_Database) {

	//TODO check cuncurrency
	//maybe the channel to notify the mainlogic should not be buffered, but unbuffered
	//The racing condition is:
	//What if mainlogic try to read a newOvsDatabase, but newOvsDatabase is overrided by this method ?

	for {
		select {
		case tableUpdate := <-h.BufupdateOvs:

			//Mutex: also the mainlogic contains the reference to the OvsDatabase
			//Be sure the structs are consistent!
			h.RWMutex.Lock()

			//select on what cache table perform updates
			switch tableUpdate {

			/****INTERFACES***/
			case "Interface":
				var cache = *h.Cache

				//Init Ovs_Database (Erase & re-initialize maps)
				ovs.Clear()

				/*****Interface TABLE********/
				table, _ := cache["Interface"]
				for _, row := range table {

					/*****Interface ITEM********/
					iface := Interface_Item{}
					iface.Name = row.Fields["name"].(string)

					if extIDs, ok := row.Fields["external_ids"]; ok {
						extIDMap := extIDs.(libovsdb.OvsMap).GoMap
						if ifaceID, ok := extIDMap["iface-id"]; ok {
							iface.IfaceIdExternalIds = ifaceID.(string)
						}
					}
					//log.Noticef("%+v\n", iface)

					ovs.Interface[iface.Name] = &iface
				}
				h.RWMutex.Unlock()
				h.MainLogicNotification <- "Ovs.Interfaces"
			}
		}
	}
}
