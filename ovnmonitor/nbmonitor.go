package ovnmonitor

import (
	"github.com/netgroup-polito/iovisor-ovn/config"
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

	ovnnbdb_sock := ""
	if config.Sandbox == true {
		//Sandbox Environment
		ovnnbdb_sock = config.NbSock
		ovnnb, err := libovsdb.ConnectWithUnixSocket(config.NbSock)
		handler.Db = ovnnb
		if err != nil {
			log.Errorf("unable to Connect to %s - %s\n", config.NbSock, err)
			return
		}
	} else {
		//Openstack Real Environment
		ovnnbdb_sock = config.Nb
		ovnnb, err := libovsdb.Connect(config.FromStringToIpPort(config.Nb))
		handler.Db = ovnnb
		if err != nil {
			log.Errorf("unable to Connect to %s - %s\n", ovnnbdb_sock, err)
			return
		}
	}

	log.Noticef("starting ovn nb db monitor @ %s\n", ovnnbdb_sock)

	var notifier MyNotifier
	notifier.handler = &handler
	handler.Db.Register(notifier)

	//TODO change db
	var ovnNbDb_name = "OVN_Northbound"
	initial, err := handler.Db.MonitorAll(ovnNbDb_name, "")
	if err != nil {
		log.Errorf("unable to Monitor %s - %s\n", ovnNbDb_name, err)
		return
	}
	PopulateCache(&handler, *initial)

	//Receive all update & populate cache
	go NbParseInit(&handler)
	go ovnNbMonitorFilter(&handler)
	//	<-handler.Quit
	h = &handler
	return
}

//All the changes notification from ovsdblib are processed by this function.
//Only notification on filtered tables are considered.
//A Notification for each tablename modfied is propagated to XXXLogic() through h.Bufupdate Channel!
func ovnNbMonitorFilter(h *MonitorHandler) {
	printTable := make(map[string]int)
	printTable["Logical_Switch"] = 1
	printTable["Logical_Switch_Port"] = 1

	for {
		select {
		case currUpdate := <-h.Update:
			//manage case of new update from db

			if config.PrintOvnNb {
				PrintCache(h)
			}

			//for debug purposes, print the new rows added or modified
			//a copy of the whole db is in cache.

			for table, tableUpdate := range currUpdate.Updates {
				if _, ok := printTable[table]; ok {
					//Notify nblogic to update db structures!
					h.Bufupdate <- table

					if config.PrintOvnNbChanges {
						log.Noticef("update table: %s\n", table)
						for uuid, row := range tableUpdate.Rows {
							log.Noticef("UUID     : %s\n", uuid)

							newRow := row.New
							PrintRow(newRow)
						}
					}

					if config.PrintOvnNbFilteredTables {
						PrintCacheTable(h, table)
					}

				}
			}
		}
	}
}
