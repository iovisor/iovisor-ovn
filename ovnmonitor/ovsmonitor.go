package ovnmonitor

import "github.com/socketplane/libovsdb"

func MonitorOvsDb() {

	//handler: one for each monitor instance
	handler := MonitorHandler{}

	//channel to notificate someone with new TableUpdates
	handler.Update = make(chan *libovsdb.TableUpdates)

	//channel buffered to notify the logic of new changes
	handler.Bufupdate = make(chan string, 10000)

	//cache contan a map between string and libovsdb.Row
	cache := make(map[string]map[string]libovsdb.Row)
	handler.Cache = &cache

	// Sandbox Real Environment
	ovsdb_sock := "/home/matteo/ovs/tutorial/sandbox/db.sock"
	ovs, err := libovsdb.ConnectWithUnixSocket(ovsdb_sock)

	//Openstack Real Environment
	// ip := "127.0.0.1"
	// port := 6640
	// ovsdb_sock := ip + ":" + strconv.Itoa(port)
	// ovs, err := libovsdb.Connect(ip, port)

	handler.Db = ovs

	if err != nil {
		log.Errorf("unable to Connect to %s - %s\n", ovsdb_sock, err)
		return
	}

	log.Noticef("starting ovs local monitor @ %s\n", ovsdb_sock)

	var notifier MyNotifier
	notifier.handler = &handler
	ovs.Register(notifier)

	var ovsDb_name = "Open_vSwitch"
	initial, err := handler.Db.MonitorAll(ovsDb_name, "")
	if err != nil {
		log.Errorf("unable to Monitor %s - %s\n", ovsDb_name, err)
		return
	}
	PopulateCache(&handler, *initial)

	ovsMonitor(&handler)
	<-handler.Quit

	return
}

func ovsMonitor(h *MonitorHandler) {
	printTable := make(map[string]int)
	printTable["Interface"] = 1
	printTable["Port"] = 1
	printTable["Bridge"] = 1

	for {
		select {
		case currUpdate := <-h.Update:
			//PrintCache(h)

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
