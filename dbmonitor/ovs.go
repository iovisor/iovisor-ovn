package dbmonitor

import "github.com/socketplane/libovsdb"

func MonitorOvsDb() {

	//handler: one for each monitor instance
	handler := MonitorHandler{}

	//channel to notificate someone with new TableUpdates
	handler.update = make(chan *libovsdb.TableUpdates)
	//cache contan a map between string and libovsdb.Row
	cache := make(map[string]map[string]libovsdb.Row)
	handler.cache = &cache

	ovsdb_sock := "/home/matteo/ovs/tutorial/sandbox/db.sock"
	ovs, err := libovsdb.ConnectWithUnixSocket(ovsdb_sock)

	handler.db = ovs

	if err != nil {
		log.Errorf("unable to Connect to %s - %s\n", ovsdb_sock, err)
		return
	}

	log.Noticef("starting ovs monitor @ %s\n", ovsdb_sock)

	var notifier MyNotifier
	notifier.handler = &handler
	ovs.Register(notifier)

	var ovsDb_name = "Open_vSwitch"
	initial, err := handler.db.MonitorAll(ovsDb_name, "")
	if err != nil {
		log.Errorf("unable to Monitor %s - %s\n", ovsDb_name, err)
		return
	}
	PopulateCache(&handler, *initial)

	ovsMonitor(&handler)
	<-handler.quit

	return
}

func ovsMonitor(h *MonitorHandler) {
	for {
		select {
		case currUpdate := <-h.update:
			//PrintCache(h)

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
