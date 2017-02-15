// Copyright 2016 Politecnico di Torino
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package ovnmonitor

import (
	"reflect"
	"sync"

	"github.com/iovisor/iovisor-ovn/config"
	"github.com/socketplane/libovsdb"

	l "github.com/op/go-logging"
)

var log = l.MustGetLogger("ovnmonitor")

type NotificationHandler interface {
	Update(db *OvnDB)
}

// A logical Switch in the OVN northbound DB, contains only the data that is
// necessary for our implementation
type LogicalSwitch struct {
	uuid     string // UUID of this row in the database
	Name     string // Name of the LogicalSwitch
	Ports    map[string]*LogicalSwitchPort
	Modified bool // was it modified in the last update?
}

// A logical port in a switch in the OVNnorthbound DB
type LogicalSwitchPort struct {
	uuid      string
	parent    *LogicalSwitch
	Name      string
	IfaceName string // name of the virtual interface in the host. if empty
	//means that the interfaces has not been bound

	Type       string // interface type: "" -> VIF, "router" connected to a router
	RouterPort string // name of the router interface that the switch is connected to
	// TODO: Addresses and port security

	Modified bool // modified in the last update?
}

// An interface in an ovs bridge instance
type OvsInterface struct {
	uuid            string
	Name            string // ame of the port: e.g tap0123
	ExternalIdIface string // external id of that interface if set

	LogicalPort *LogicalSwitchPort // point to the port that owns this interface
}

// A logical router in the NB db
type LogicalRouter struct {
	uuid  string
	Name  string
	Ports map[string]*LogicalRouterPort
	// TODO: static routes
	Enabled  bool
	Modified bool // was it modified in the last update?
}

// A port in a logical router
type LogicalRouterPort struct {
	uuid     string
	parent   *LogicalRouter
	Name     string
	Mac      string
	Networks string // TODO: convert to array in the future
	Enabled  bool
	Modified bool // was it modified in the last update?
}

type OvnDB struct {
	Switches map[string]*LogicalSwitch // switches indexed by name
	Routers  map[string]*LogicalRouter // routers indexed by name

	// Private: this fields are only used by ovnmonitor
	// Contains all the logical switches present in the NB DB. Indexed by UUID
	logicalSwitches map[string]*LogicalSwitch

	// Contains all the logical ports present in the NB DB. Indexed by UUID
	logicalSwitchPorts map[string]*LogicalSwitchPort

	// Contains all the interfaces present in OVS DB. Indexed by UUID
	ovsInterfaces map[string]*OvsInterface

	// Contains all the routers present in the NB DB.  Indexed by UUID
	logicalRouters map[string]*LogicalRouter

	// Contains all the logical router ports.  Indexed by UUID
	logicalRouterPorts map[string]*LogicalRouterPort
}

func CreateMonitor() *OVNMonitor {
	mon := new(OVNMonitor)
	mon.DB.Switches = make(map[string]*LogicalSwitch)
	mon.DB.logicalSwitches = make(map[string]*LogicalSwitch)
	mon.DB.logicalSwitchPorts = make(map[string]*LogicalSwitchPort)
	mon.DB.ovsInterfaces = make(map[string]*OvsInterface)
	mon.DB.Routers = make(map[string]*LogicalRouter)
	mon.DB.logicalRouters = make(map[string]*LogicalRouter)
	mon.DB.logicalRouterPorts = make(map[string]*LogicalRouterPort)
	return mon
}

type OVNMonitor struct {
	DB        OvnDB
	ovsClient *libovsdb.OvsdbClient
	nbClient  *libovsdb.OvsdbClient

	handler NotificationHandler
}

func (o *OVNMonitor) Connect() (db *OvnDB, error bool) {
	// first connect to the NB DB
	ovnnbdb_sock := ""
	if config.Sandbox == true {
		// Sandbox Environment
		ovnnbdb_sock = config.NbSock
		nbClient, err := libovsdb.ConnectWithUnixSocket(config.NbSock)

		if err != nil {
			log.Errorf("unable to Connect to %s - %s\n", ovnnbdb_sock, err)
			return nil, false
		}

		o.nbClient = nbClient
	} else {
		// Openstack Real Environment
		ovnnbdb_sock = config.Nb
		nbClient, err := libovsdb.Connect(config.FromStringToIpPort(config.Nb))

		if err != nil {
			log.Errorf("unable to Connect to %s - %s\n", ovnnbdb_sock, err)
			return nil, false
		}

		o.nbClient = nbClient
	}

	// connect to the ovs db
	ovsdb_sock := ""
	if config.Sandbox == true {
		// Sandbox Real Environment
		ovsdb_sock = config.OvsSock
		ovsClient, err := libovsdb.ConnectWithUnixSocket(config.OvsSock)

		if err != nil {
			log.Errorf("unable to Connect to %s - %s\n", ovsdb_sock, err)
			return nil, false
		}

		o.ovsClient = ovsClient

	} else {
		// Openstack Real Environment
		ovsdb_sock = config.Ovs
		ovsClient, err := libovsdb.Connect(config.FromStringToIpPort(config.Ovs))

		if err != nil {
			log.Errorf("unable to Connect to %s - %s\n", ovsdb_sock, err)
			return nil, false
		}

		o.ovsClient = ovsClient
	}

	// register notifiers
	var notifier MyNotifier
	notifier.monitor = o
	notifier.mutex = new(sync.Mutex)
	o.ovsClient.Register(notifier)
	o.nbClient.Register(notifier)

	initialNb, err := o.nbClient.MonitorAll("OVN_Northbound", "")
	if err != nil {
		log.Errorf("unable to Monitor OVN_Northbound - %s\n", err)
		return nil, false
	}

	UpdateDB(&o.DB, *initialNb)

	initialOvs, err := o.ovsClient.MonitorAll("Open_vSwitch", "")
	if err != nil {
		log.Errorf("unable to Monitor Open_vSwitch - %s\n", err)
		return nil, false
	}

	UpdateDB(&o.DB, *initialOvs)

	return &o.DB, true
}

func (o *OVNMonitor) Register(handler NotificationHandler) {

	// TODO: Add mutex

	o.handler = handler
}

// TODO: Add unregister function
func UpdateDB(db *OvnDB, updates libovsdb.TableUpdates) {
	log.Noticef("UpdateDB() init\n")

	// I know that you can think that this is a silly imlementation because the
	// three conditionals could be places on the same loop.  But It is not true,
	// in this way we can guarantee that all the ports are processed before all
	// the switchs, this avoids having a switch with a reference to a non-existing
	// port

	for table, tableUpdate := range updates.Updates {
		switch table {
		case "Logical_Switch_Port":
			for uuid, row := range tableUpdate.Rows {
				ProcessLogicalSwitchPort(db, uuid, row)
			}
		}
	}

	for table, tableUpdate := range updates.Updates {
		switch table {
		case "Logical_Switch":
			for uuid, row := range tableUpdate.Rows {
				ProcessLogicalSwitch(db, uuid, row)
			}
		}
	}

	for table, tableUpdate := range updates.Updates {
		switch table {
		case "Interface":
			for uuid, row := range tableUpdate.Rows {
				ProcessInterface(db, uuid, row)
			}
		}
	}

	for table, tableUpdate := range updates.Updates {
		switch table {
		case "Logical_Router_Port":
			for uuid, row := range tableUpdate.Rows {
				ProcessLogicalRouterPort(db, uuid, row)
			}
		}
	}

	for table, tableUpdate := range updates.Updates {
		switch table {
		case "Logical_Router":
			for uuid, row := range tableUpdate.Rows {
				ProcessLogicalRouter(db, uuid, row)
			}
		}
	}

	log.Noticef("UpdateDB() finish\n")
}

func ProcessLogicalSwitch(db *OvnDB, uuid string, row libovsdb.RowUpdate) {
	log.Noticef("ProcessLogicalSwitch()")
	if sw, ok := db.logicalSwitches[uuid]; ok { // the switch is already on the db
		empty := libovsdb.Row{}
		if reflect.DeepEqual(row.New, empty) {
			delete(db.logicalSwitches, uuid)
			delete(db.Switches, sw.Name)
		} else { // update switch
			ParseLogicalSwitch(sw, row.New)
			UpdateSwitchPorts(db, sw, row.New.Fields["ports"])
			sw.Modified = true
		}
	} else { // create new switch
		sw := new(LogicalSwitch)
		sw.uuid = uuid
		ParseLogicalSwitch(sw, row.New)
		UpdateSwitchPorts(db, sw, row.New.Fields["ports"])
		sw.Modified = true
		db.logicalSwitches[uuid] = sw
		db.Switches[sw.Name] = sw
	}
}

func ParseLogicalSwitch(s *LogicalSwitch, row libovsdb.Row) {
	s.Name = row.Fields["name"].(string)
}

func UpdateSwitchPorts(db *OvnDB, sw *LogicalSwitch, ports interface{}) {
	// firstly convert the port list into a more handable data structure
	portsMap := make(map[string]string)
	switch ports.(type) {
	case libovsdb.UUID:
		portsMap[ports.(libovsdb.UUID).GoUUID] = ports.(libovsdb.UUID).GoUUID
	case libovsdb.OvsSet:
		for _, uuids := range ports.(libovsdb.OvsSet).GoSet {
			portsMap[uuids.(libovsdb.UUID).GoUUID] = uuids.(libovsdb.UUID).GoUUID
		}
	}

	// secondly update pointers to ports inside the switch
	sw.Ports = make(map[string]*LogicalSwitchPort)

	for uuid, _ := range portsMap {
		port := db.logicalSwitchPorts[uuid]

		if port == nil {
			log.Noticef("OMG, we have a port without a port")
		}
		sw.Ports[port.Name] = port
		port.parent = sw
	}
}

func ProcessLogicalSwitchPort(db *OvnDB, uuid string, row libovsdb.RowUpdate) {
	log.Noticef("ProcessLogicalSwitchPort()")
	if port, ok := db.logicalSwitchPorts[uuid]; ok { // the port is already on the db
		empty := libovsdb.Row{}
		if reflect.DeepEqual(row.New, empty) {
			delete(db.logicalSwitchPorts, uuid)
		} else { // update port
			ParseLogicalSwitchPort(port, row.New)
			port.Modified = true
			if port.parent != nil {
				port.parent.Modified = true
			}
		}
	} else { // new logical switch port
		port := new(LogicalSwitchPort)
		port.uuid = uuid
		ParseLogicalSwitchPort(port, row.New)
		UpdateBoundingPort(&db.ovsInterfaces, port)
		port.Modified = true
		db.logicalSwitchPorts[uuid] = port
	}
}

func ParseLogicalSwitchPort(s *LogicalSwitchPort, row libovsdb.Row) {
	s.Name = row.Fields["name"].(string)
	s.Type = row.Fields["type"].(string)

	if s.Type == "router" {
		// TODO: are security checks needed?
		mymap := row.Fields["options"].(libovsdb.OvsMap).GoMap
		s.RouterPort = mymap["router-port"].(string)
	}
	// addresses and port security to do
}

func ProcessInterface(db *OvnDB, uuid string, row libovsdb.RowUpdate) {

	log.Noticef("ProcessInterface()")

	if iface, ok := db.ovsInterfaces[uuid]; ok { // the interface is already on the db
		empty := libovsdb.Row{}

		if reflect.DeepEqual(row.New, empty) { // deleted
			// If the interface has been deleted it is necessary to remove the
			// bounding (if existed)
			CleanBounding(iface)
			delete(db.ovsInterfaces, uuid)
		} else { //  modified
			// for us there is just a single field that is important in the
			// interface, this is the external id. To save computing time, check
			// only if it has changed
			//
			// TODO: Can the name of an interface change?

			newIface := new(OvsInterface)
			ParseOvsInterface(newIface, row.New)

			if iface.ExternalIdIface != newIface.ExternalIdIface {
				// the interface is now bound do a different port. Update it
				CleanBounding(iface)
				iface.ExternalIdIface = newIface.ExternalIdIface
				UpdateBounding(&db.logicalSwitchPorts, iface)
			}
		}

	} else {
		x := new(OvsInterface)
		x.uuid = uuid
		ParseOvsInterface(x, row.New)
		UpdateBounding(&db.logicalSwitchPorts, x)
		db.ovsInterfaces[uuid] = x
	}
}

// looks if there is a port that matches the interface external id

func UpdateBounding(ports *map[string]*LogicalSwitchPort, iface *OvsInterface) {
	if iface.ExternalIdIface == "" {
		return
	}

	for _, port := range *ports {
		if port.Name == iface.ExternalIdIface {
			// voila! we have  a match

			iface.LogicalPort = port
			port.IfaceName = iface.Name
			port.Modified = true
			port.parent.Modified = true

			log.Noticef("Voila! port '%s' uses inteface: '%s'", port.Name, iface.Name)
			break
		}
	}
}

func CleanBounding(iface *OvsInterface) {
	if iface.LogicalPort == nil {
		return
	}

	log.Noticef("Port '%s' is now unbounded.", iface.LogicalPort.Name)

	iface.LogicalPort.IfaceName = ""
	iface.LogicalPort.Modified = true
	iface.LogicalPort.parent.Modified = true

	iface.LogicalPort = nil
}

// looks if there is an Interface that matches this port

func UpdateBoundingPort(ifaces *map[string]*OvsInterface, port *LogicalSwitchPort) {
	for _, iface := range *ifaces {
		if iface.ExternalIdIface != "" && iface.ExternalIdIface == port.Name {
			iface.LogicalPort = port
			port.IfaceName = iface.Name
			port.Modified = true
			if port.parent != nil {
				port.parent.Modified = true
			}

			log.Noticef("Voila! port '%s' uses inteface: '%s'", port.Name, iface.Name)
			break
		}
	}
}

func ParseOvsInterface(ovs *OvsInterface, row libovsdb.Row) {
	ovs.Name = row.Fields["name"].(string)

	if extIDs, ok := row.Fields["external_ids"]; ok {
		extIDMap := extIDs.(libovsdb.OvsMap).GoMap
		if ifaceID, ok := extIDMap["iface-id"]; ok {
			ovs.ExternalIdIface = ifaceID.(string)
		}
	}
}

func ProcessLogicalRouter(db *OvnDB, uuid string, row libovsdb.RowUpdate) {
	log.Noticef("ProcessLogicalRouter()")
	if router, ok := db.logicalRouters[uuid]; ok { // the router is already on the db
		empty := libovsdb.Row{}
		if reflect.DeepEqual(row.New, empty) {
			delete(db.logicalRouters, uuid)
			delete(db.Routers, router.Name)
		} else { // update router
			ParseLogicalRouter(router, row.New)
			UpdateRouterPorts(db, router, row.New.Fields["ports"])
			router.Modified = true
		}
	} else { // create new router
		router := new(LogicalRouter)
		router.uuid = uuid
		ParseLogicalRouter(router, row.New)
		UpdateRouterPorts(db, router, row.New.Fields["ports"])
		router.Modified = true
		db.logicalRouters[uuid] = router
		db.Routers[router.Name] = router
	}
}

func ParseLogicalRouter(r *LogicalRouter, row libovsdb.Row) {
	r.Name = row.Fields["name"].(string)
	//r.Enabled = row.Fields["enabled"].(bool)
}

func UpdateRouterPorts(db *OvnDB, r *LogicalRouter, ports interface{}) {
	// firstly convert the port list into a more handable data structure
	portsMap := make(map[string]string)
	switch ports.(type) {
	case libovsdb.UUID:
		portsMap[ports.(libovsdb.UUID).GoUUID] = ports.(libovsdb.UUID).GoUUID
	case libovsdb.OvsSet:
		for _, uuids := range ports.(libovsdb.OvsSet).GoSet {
			portsMap[uuids.(libovsdb.UUID).GoUUID] = uuids.(libovsdb.UUID).GoUUID
		}
	}

	// secondly update pointers to ports inside the router
	r.Ports = make(map[string]*LogicalRouterPort)

	for uuid, _ := range portsMap {
		port := db.logicalRouterPorts[uuid]

		if port == nil {
			log.Noticef("OMG, we have a port without a port")
		}
		r.Ports[port.Name] = port
		port.parent = r
	}
}

func ProcessLogicalRouterPort(db *OvnDB, uuid string, row libovsdb.RowUpdate) {
	log.Noticef("ProcessLogicalRouterPort()")

	if port, ok := db.logicalRouterPorts[uuid]; ok { // the port is already on the db
		empty := libovsdb.Row{}
		if reflect.DeepEqual(row.New, empty) {
			delete(db.logicalRouterPorts, uuid)
		} else { // update port
			ParseLogicalRouterPort(port, row.New)
			port.Modified = true
			if port.parent != nil {
				port.parent.Modified = true
			}
		}
	} else { // new logical router port
		port := new(LogicalRouterPort)
		port.uuid = uuid
		ParseLogicalRouterPort(port, row.New)
		port.Modified = true
		db.logicalRouterPorts[uuid] = port
	}
}

func ParseLogicalRouterPort(r *LogicalRouterPort, row libovsdb.Row) {
	r.Name = row.Fields["name"].(string)
	r.Mac = row.Fields["mac"].(string)
	r.Networks = row.Fields["networks"].(string)

	// TODO: networks
}

type MyNotifier struct {
	monitor *OVNMonitor // points to the struct that should be updated
	mutex   *sync.Mutex
}

func (n MyNotifier) Update(context interface{}, tableUpdates libovsdb.TableUpdates) {
	// It is necessary to use a mutex because it is possible that a notification
	// arrives while the last one is still being processed.
	// TODO: How does this mutex affect scalability?
	n.mutex.Lock()
	UpdateDB(&n.monitor.DB, tableUpdates)
	if n.monitor.handler != nil {
		n.monitor.handler.Update(&n.monitor.DB)
	}
	n.mutex.Unlock()
}

func (n MyNotifier) Locked([]interface{}) {
}

func (n MyNotifier) Stolen([]interface{}) {
}

func (n MyNotifier) Echo([]interface{}) {
}

func (n MyNotifier) Disconnected(client *libovsdb.OvsdbClient) {
}
