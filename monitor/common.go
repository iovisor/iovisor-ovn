package monitor

import (
	"reflect"

	l "github.com/op/go-logging"

	"github.com/socketplane/libovsdb"
)

var log = l.MustGetLogger("politoctrl")

type MonitorHandler struct {
	quit   chan bool
	update chan *libovsdb.TableUpdates
	cache  *map[string]map[string]libovsdb.Row
	db     *libovsdb.OvsdbClient
}

func PrintRow(row libovsdb.Row) {
	for key, value := range row.Fields {
		log.Debugf("%20s : %s\n", key, value)
	}
}

//DUMP ALL THE DB
func PrintCache(h *MonitorHandler) {
	var cache = *h.cache
	log.Noticef("print all tables in cache\n")
	for tableName, table := range cache {
		log.Noticef("%20s:%s\n", "TABLE", tableName)
		for uuid, row := range table {
			log.Noticef("%20s:%s\n", "UUID", uuid)
			PrintRow(row)
		}
	}
}

//TODO
//DUMP TABLE X

func PopulateCache(h *MonitorHandler, updates libovsdb.TableUpdates) {
	var cache = *h.cache
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

type MyNotifier struct {
	handler *MonitorHandler
}

func (n MyNotifier) Update(context interface{}, tableUpdates libovsdb.TableUpdates /*, h *MonitorHandler*/) {
	PopulateCache(n.handler, tableUpdates)
	n.handler.update <- &tableUpdates
}
func (n MyNotifier) Locked([]interface{}) {
}
func (n MyNotifier) Stolen([]interface{}) {
}
func (n MyNotifier) Echo([]interface{}) {
}
func (n MyNotifier) Disconnected(client *libovsdb.OvsdbClient) {
}
