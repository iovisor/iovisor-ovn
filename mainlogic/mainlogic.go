package mainlogic

import (
	//"strconv"
	//"time"
	"os"

	"github.com/netgroup-polito/iovisor-ovn/config"
	"github.com/netgroup-polito/iovisor-ovn/iomodules/l2switch"
	"github.com/netgroup-polito/iovisor-ovn/hoverctl"
	"github.com/netgroup-polito/iovisor-ovn/ovnmonitor"
	l "github.com/op/go-logging"
)

const brint = "br-int"

var log = l.MustGetLogger("mainlogic")

type L2Switch struct {
	Name string
	swIomodule 	*l2switch.L2SwitchModule
	ports map[string]*L2SwitchPort
}

type L2SwitchPort struct {
	Name string
	IfaceName 	string
}
/*
 * Contains the switches that have been created.
 * They are indexed by a name, in this case is the OVN name
 */
var switches map[string]*L2Switch

var dataplane *hoverctl.Dataplane
func MainLogic() {

	mon := ovnmonitor.CreateMonitor()
	if !mon.Connect() {
		log.Errorf("Error connecting to OVN databases\n")
		return
	}

	switches = make(map[string]*L2Switch)

	dataplane = hoverctl.NewDataplane()
	if err := dataplane.Init(config.Hover); err != nil {
		log.Errorf("unable to conect to Hover %s\n%s\n", config.Hover, err)
		os.Exit(1)
	}

	var notifier MyNotifier
	mon.Register(&notifier)
}

type MyNotifier struct {

}

func (m *MyNotifier) Update(db *ovnmonitor.OvnDB) {
	/* detect removed switches*/
	for name, sw := range switches {
		if _, ok := db.Switches[name]; !ok {
			removeSwitch(sw)
		}
	}

	/* detect new switches */
	for name, lsw := range db.Switches {
		if _, ok := switches[name]; !ok {
			addSwitch(lsw)
		}
	}

	/* process modified switches */
	for _, lsw := range db.Switches {
		if lsw.Modified {
				updateSwitch(lsw)
		}
	}
 }

func removeSwitch(sw *L2Switch) {
	/* TODO: remove ports already present on the switch*/
	sw.swIomodule.Destroy()
	delete(switches, sw.Name)
}

func addSwitch(lsw *ovnmonitor.LogicalSwitch) {
	sw := new(L2Switch)
	sw.Name = lsw.Name
	sw.ports = make(map[string]*L2SwitchPort)
	sw.swIomodule = l2switch.Create(dataplane)
	switches[sw.Name] = sw
}

func updateSwitch(lsw *ovnmonitor.LogicalSwitch) {
	sw := switches[lsw.Name]
	/* look for deleted ports*/
	for name, port := range sw.ports {
		if _, ok := lsw.Ports[name]; !ok {
			removePort(sw, port)
		}
	}

	/* look for added ports */
	for name, lport := range lsw.Ports {
		if _, ok := sw.ports[name]; !ok {
			addPort(sw, lport)
		}
	}

	/* process modified ports*/
	for _, lport := range lsw.Ports {
		if lport.Modified {
			updatePort(sw, lport)
		}
	}

	lsw.Modified = false
}

func removePort(sw *L2Switch, port *L2SwitchPort) {
	if port.IfaceName != "" {
		sw.swIomodule.DetachExternalInterface(port.IfaceName)
	}

	delete(sw.ports, port.Name)
}

func addPort(sw *L2Switch, lport *ovnmonitor.LogicalSwitchPort) {
	port := new(L2SwitchPort)
	port.Name = lport.Name
	sw.ports[port.Name] = port
}

func updatePort(sw *L2Switch, lport *ovnmonitor.LogicalSwitchPort) {
	port := sw.ports[lport.Name]

	/* is IfaceChanged? */
	if port.IfaceName != lport.IfaceName {
		/* if it was connected to an iface */
		if port.IfaceName != "" {
			sw.swIomodule.DetachExternalInterface(port.IfaceName)
		}

		port.IfaceName = lport.IfaceName

		if port.IfaceName != "" {
			if sw.swIomodule.PortsCount == 0 {
				sw.swIomodule.Deploy()
			}
			sw.swIomodule.AttachExternalInterface(port.IfaceName)
		}
	}

	lport.Modified = false;
}
