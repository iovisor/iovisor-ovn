package ovnmonitor

//Idea: new notification triggers this method.
//But the logic is based on cache
//We compare cache module with local db structs!
//And we can triggers new IOVisor events

func NbLogicInit(h *MonitorHandler) *Nb_Database {
	nb := Nb_Database{}
	go NbLogic(h, &nb)
	return &nb
}

func NbLogic(h *MonitorHandler, nb *Nb_Database) {

	for {
		select {
		case tableUpdate := <-h.Bufupdate:
			log.Noticef("Notification on Table %s\n", tableUpdate)

			switch tableUpdate {
			case "Logical_Switch":
				log.Notice("Logical_Switch")
				PrintCacheTable(h, tableUpdate)
				//use the local struct to perform a check_>
				//insert, delete, update?
			case "Logical_Switch_Port":
				log.Notice("Logical_Switch_Port")
				PrintCacheTable(h, tableUpdate)
			}
		}
	}
}
