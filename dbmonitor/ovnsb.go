package dbmonitor

import "github.com/socketplane/libovsdb"

func MonitorOvnSb() {

	//handler: one for each monitor instance
	handler := MonitorHandler{}

	//channel to notificate someone with new TableUpdates
	handler.update = make(chan *libovsdb.TableUpdates)
	//cache contan a map between string and libovsdb.Row
	cache := make(map[string]map[string]libovsdb.Row)
	handler.cache = &cache

	ovnsbdb_sock := "/home/matteo/ovs/tutorial/sandbox/ovnsb_db.sock"
	ovnsb, err := libovsdb.ConnectWithUnixSocket(ovnsbdb_sock)

	handler.db = ovnsb

	if err != nil {
		log.Errorf("unable to Connect to %s - %s\n", ovnsbdb_sock, err)
		return
	}

	log.Noticef("starting ovn sb db monitor @ %s\n", ovnsbdb_sock)

	var notifier MyNotifier
	notifier.handler = &handler
	ovnsb.Register(notifier)

	//TODO change db
	var ovnSbDb_name = "OVN_Southbound"
	initial, err := ovnsb.MonitorAll(ovnSbDb_name, "")
	if err != nil {
		log.Errorf("unable to Monitor %s - %s\n", ovnSbDb_name, err)
		return
	}
	PopulateCache(&handler, *initial)

	ovnSbMonitor(&handler)
	<-handler.quit

	return
}

func ovnSbMonitor(h *MonitorHandler) {
	printTable := make(map[string]int)
	printTable["Port_Binding"] = 1
	//printTable["Logical_Switch_Port"] = 1

	for {
		select {
		case currUpdate := <-h.update:
			//manage case of new update from db

			//for debug purposes, print the new rows added or modified
			//a copy of the whole db is in cache.

			for table, tableUpdate := range currUpdate.Updates {
				if _, ok := printTable[table]; ok {

					log.Noticef("update table: %s\n", table)
					for uuid, row := range tableUpdate.Rows {
						log.Noticef("UUID     : %s\n", uuid)

						newRow := row.New
						PrintRow(newRow)
					}
				}
			}
		}
	}
}
