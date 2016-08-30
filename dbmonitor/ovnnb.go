package dbmonitor

import "github.com/socketplane/libovsdb"

func MonitorOvnNb() {

	//handler: one for each monitor instance
	handler := MonitorHandler{}

	//channel to notificate someone with new TableUpdates
	handler.update = make(chan *libovsdb.TableUpdates)
	//cache contan a map between string and libovsdb.Row
	cache := make(map[string]map[string]libovsdb.Row)
	handler.cache = &cache

	ovnnbdb_sock := "/home/matteo/ovs/tutorial/sandbox/ovnnb_db.sock"
	ovnnb, err := libovsdb.ConnectWithUnixSocket(ovnnbdb_sock)

	handler.db = ovnnb

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

	ovnNbMonitor(&handler)
	<-handler.quit

	return
}

func ovnNbMonitor(h *MonitorHandler) {
	for {
		select {
		case currUpdate := <-h.update:
			//manage case of new update from db

			//for debug purposes, print the new rows added or modified
			//a copy of the whole db is in cache.

			for table, tableUpdate := range currUpdate.Updates {
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
