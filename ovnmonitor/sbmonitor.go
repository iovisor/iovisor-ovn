package ovnmonitor

import "github.com/socketplane/libovsdb"

func MonitorOvnSb() {

	//handler: one for each monitor instance
	handler := MonitorHandler{}

	//channel to notificate someone with new TableUpdates
	handler.Update = make(chan *libovsdb.TableUpdates)
	//cache contan a map between string and libovsdb.Row

	//channel buffered to notify the logic of new changes
	handler.Bufupdate = make(chan string, 10000)

	cache := make(map[string]map[string]libovsdb.Row)
	handler.Cache = &cache

	//Sandbox Environment
	ovnsbdb_sock := "/home/matteo/ovs/tutorial/sandbox/ovnsb_db.sock"
	ovnsb, err := libovsdb.ConnectWithUnixSocket(ovnsbdb_sock)

	// //Openstack Real Environment
	// ip := "127.0.0.1"
	// port := 6642
	// ovnsbdb_sock := ip + ":" + strconv.Itoa(port)
	// ovnsb, err := libovsdb.Connect(ip, port)

	handler.Db = ovnsb

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
	<-handler.Quit

	return
}

func ovnSbMonitor(h *MonitorHandler) {
	printTable := make(map[string]int)
	// printTable["Port_Binding"] = 1
	// printTable["Chassis"] = 1

	for {
		select {
		case currUpdate := <-h.Update:
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
