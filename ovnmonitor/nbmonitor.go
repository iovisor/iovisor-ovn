package ovnmonitor

import (
	"github.com/socketplane/libovsdb"
)

//start to monitor Ovn Nb databases
//For now, only one db, in future db string passedasparameter
func MonitorOvnNb() (h *MonitorHandler) {

	//handler: one for each monitor instance
	handler := MonitorHandler{}

	//channel to notificate someone with new TableUpdates
	handler.Update = make(chan *libovsdb.TableUpdates)

	//channel buffered to notify the logic of new changes
	handler.Bufupdate = make(chan string, 10000)

	//Channel buffered to notify the logic fo new changes
	handler.MainLogicNotification = make(chan string, 100)

	//cache contan a map between string and libovsdb.Row
	cache := make(map[string]map[string]libovsdb.Row)
	handler.Cache = &cache

	//Sandbox Environment
	ovnnbdb_sock := "/home/matteo/ovs/tutorial/sandbox/ovnnb_db.sock"
	ovnnb, err := libovsdb.ConnectWithUnixSocket(ovnnbdb_sock)

	//Openstack Real Environment
	// ip := "127.0.0.1"
	// port := 6641
	// ovnnbdb_sock := ip + ":" + strconv.Itoa(port)
	// ovnnb, err := libovsdb.Connect(ip, port)

	handler.Db = ovnnb

	if err != nil {
		log.Errorf("unable to Connect to %s - %s\n", ovnnbdb_sock, err)
		return
	}

	log.Noticef("starting ovn nb db monitor @ %s\n", ovnnbdb_sock)

	var notifier MyNotifier
	notifier.handler = &handler
	ovnnb.Register(notifier)

	//TODO change db
	var ovnNbDb_name = "OVN_Northbound"
	initial, err := ovnnb.MonitorAll(ovnNbDb_name, "")
	if err != nil {
		log.Errorf("unable to Monitor %s - %s\n", ovnNbDb_name, err)
		return
	}
	PopulateCache(&handler, *initial)

	//Receive all update & populate cache
	go NbLogicInit(&handler)
	go ovnNbMonitorFilter(&handler)
	//	<-handler.Quit
	h = &handler
	return
}

func ovnNbMonitorFilter(h *MonitorHandler) {
	printTable := make(map[string]int)
	printTable["Logical_Switch"] = 1
	printTable["Logical_Switch_Port"] = 1

	for {
		select {
		case currUpdate := <-h.Update:
			//manage case of new update from db

			//for debug purposes, print the new rows added or modified
			//a copy of the whole db is in cache.

			for table, _ /*tableUpdate*/ := range currUpdate.Updates {
				if _, ok := printTable[table]; ok {
					//Notify nblogic to update db structures!
					h.Bufupdate <- table

					// log.Noticef("update table: %s\n", table)
					// for uuid, row := range tableUpdate.Rows {
					// 	log.Noticef("UUID     : %s\n", uuid)
					//
					// 	newRow := row.New
					// 	PrintRow(newRow)
					// }
				}
			}
		}
	}
}
