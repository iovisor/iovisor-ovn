package monitor

import (
	"reflect"

	l "github.com/op/go-logging"

	"github.com/socketplane/libovsdb"
)

var log = l.MustGetLogger("politoctrl")

var quit chan bool
var update chan *libovsdb.TableUpdates
var cache map[string]map[string]libovsdb.Row

//monitor ovs databse changes

func MonitorOvsDb() {

	//channel to notificate someone with new TableUpdates
	update = make(chan *libovsdb.TableUpdates)
	//cache contan a map between string and libovsdb.Row
	cache = make(map[string]map[string]libovsdb.Row)

	ovsdb_sock := "/home/matteo/ovs/tutorial/sandbox/db.sock"
	ovs, err := libovsdb.ConnectWithUnixSocket(ovsdb_sock)

	if err != nil {
		log.Errorf("unable to Connect to %s - %s\n", ovsdb_sock, err)
		return
	}

	log.Infof("starting ovs monitor @ %s\n", ovsdb_sock)

	var notifier myNotifier
	ovs.Register(notifier)

	initial, _ := ovs.MonitorAll("Open_vSwitch", "")
	populateCache(*initial)

	ovsMonitor(ovs)
	<-quit

	return
}

func PrintRow(row libovsdb.Row) {
	for key, value := range row.Fields {
		log.Debugf("%20s : %s\n", key, value)
	}
}

func ovsMonitor(ovs *libovsdb.OvsdbClient) {
	for {
		select {
		case currUpdate := <-update:
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

/*
  if _, ok := newRow.Fields["name"]; ok {
    name := newRow.Fields["name"].(string)
    log.Debugf("update name : %s\n", name)
*/
/*
   if name == "stop" {
     fmt.Println("Bridge stop detected : ", uuid)
     ovs.Disconnect()
     quit <- true
   }
*/

func populateCache(updates libovsdb.TableUpdates) {
	for table, tableUpdate := range updates.Updates {
		if _, ok := cache[table]; !ok {
			cache[table] = make(map[string]libovsdb.Row)

		}
		for uuid, row := range tableUpdate.Rows {
			empty := libovsdb.Row{}
			if !reflect.DeepEqual(row.New, empty) {
				cache[table][uuid] = row.New
			} else {
				delete(cache[table], uuid)
			}
		}
	}
}

type myNotifier struct {
}

func (n myNotifier) Update(context interface{}, tableUpdates libovsdb.TableUpdates) {
	populateCache(tableUpdates)
	update <- &tableUpdates
}
func (n myNotifier) Locked([]interface{}) {
}
func (n myNotifier) Stolen([]interface{}) {
}
func (n myNotifier) Echo([]interface{}) {
}
func (n myNotifier) Disconnected(client *libovsdb.OvsdbClient) {
}
