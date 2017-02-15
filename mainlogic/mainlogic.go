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
package mainlogic

import (
	"encoding/hex"
	"fmt"
	"net"
	"os"

	"github.com/iovisor/iovisor-ovn/config"
	"github.com/iovisor/iovisor-ovn/hover"
	"github.com/iovisor/iovisor-ovn/iomodules"
	"github.com/iovisor/iovisor-ovn/iomodules/l2switch"
	"github.com/iovisor/iovisor-ovn/iomodules/router"
	"github.com/iovisor/iovisor-ovn/ovnmonitor"
	l "github.com/op/go-logging"
)

const brint = "br-int"

var log = l.MustGetLogger("mainlogic")

type L2Switch struct {
	Name       string
	swIomodule *l2switch.L2SwitchModule
	ports      map[string]*L2SwitchPort
}

type L2SwitchPort struct {
	Name      string
	IfaceName string
}

type Router struct {
	Name      string
	rIoModule *router.RouterModule
	ports     map[string]*RouterPort
}

type RouterPort struct {
	Name string
	IP   string // TODO: change this to a better data structure
	Mask string
	Mac  string
}

/*
 * Contains the switches that have been created.
 * They are indexed by a name, in this case is the OVN name
 */
var switches map[string]*L2Switch

// Contains the routers that haven been created.  Indexed by named
var routers map[string]*Router

var hc *hover.Client
var Mon *ovnmonitor.OVNMonitor

func GetHoverClient() *hover.Client {
	return hc
}

func MainLogic() {

	Mon = ovnmonitor.CreateMonitor()
	db, err := Mon.Connect()

	if err == false { // it is a quite odd that false means error
		log.Errorf("Error connecting to OVN databases\n")
		return
	}

	switches = make(map[string]*L2Switch)
	routers = make(map[string]*Router)

	hc = hover.NewClient()
	if err := hc.Init(config.Hover); err != nil {
		log.Errorf("unable to conect to Hover %s\n%s\n", config.Hover, err)
		os.Exit(1)
	}

	var notifier MyNotifier
	notifier.Update(db) // I think that there is an instant of time where the info could be lost
	Mon.Register(&notifier)

}

type MyNotifier struct {
}

func (m *MyNotifier) Update(db *ovnmonitor.OvnDB) {

	log.Noticef("Mainlogic update() init")

	// detect removed routers
	for name, r := range routers {
		if _, ok := db.Routers[name]; !ok {
			removeRouter(r)
		}
	}

	// detect new routers
	for name, lr := range db.Routers {
		if _, ok := routers[name]; !ok {
			addRouter(lr)
		}
	}

	// process modified routers
	for _, lr := range db.Routers {
		if lr.Modified {
			updateRouter(lr)
		}
	}
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

	log.Noticef("Mainlogic update() finished")
}

func removeRouter(r *Router) {
	r.rIoModule.Destroy()
	delete(routers, r.Name)
}

func addRouter(lr *ovnmonitor.LogicalRouter) {
	r := new(Router)
	r.Name = lr.Name
	r.ports = make(map[string]*RouterPort)
	r.rIoModule = router.Create(hc)
	routers[r.Name] = r

	r.rIoModule.Deploy()
}

func updateRouter(lr *ovnmonitor.LogicalRouter) {
	r := routers[lr.Name]
	// look for deleted ports
	for name, port := range r.ports {
		if _, ok := lr.Ports[name]; !ok {
			removeRouterPort(r, port)
		}
	}

	// look for added ports
	for name, lrp := range lr.Ports {
		if _, ok := r.ports[name]; !ok {
			addRouterPort(r, lrp)
		}
	}

	// look for modified ports
	for _, lrp := range lr.Ports {
		if lrp.Modified {
			updateRouterPort(r, lrp)
		}
	}

	lr.Modified = false
}

func removeRouterPort(r *Router, port *RouterPort) {
	// TODO: what else should be done here?
	delete(r.ports, port.Name)
}

func addRouterPort(r *Router, lrp *ovnmonitor.LogicalRouterPort) {
	port := new(RouterPort)
	port.Name = lrp.Name
	ip, ipnet, _ := net.ParseCIDR(lrp.Networks)
	if ip.To4() != nil {
		port.IP = ip.String()
		a, _ := hex.DecodeString(ipnet.Mask.String())
		port.Mask = fmt.Sprintf("%v.%v.%v.%v", a[0], a[1], a[2], a[3])
	}
	port.Mac = lrp.Mac
	r.ports[port.Name] = port
}

func updateRouterPort(r *Router, lrp *ovnmonitor.LogicalRouterPort) {
	//port := r.ports[lrp.Name]

	// TODO: Another thing to be done :(
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
	sw.swIomodule = l2switch.Create(hc)
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

	if lport.Type == "" {
		/* is IfaceChanged? */
		if port.IfaceName != lport.IfaceName {
			/* if it was connected to an iface */
			if port.IfaceName != "" {
				sw.swIomodule.DetachExternalInterface(port.IfaceName)
			}

			port.IfaceName = lport.IfaceName

			// should this port be bound to a VIF interface?
			if port.IfaceName != "" {
				if sw.swIomodule.PortsCount == 0 {
					sw.swIomodule.Deploy()
				}
				sw.swIomodule.AttachExternalInterface(port.IfaceName)
			}
		}
	} else if lport.Type == "router" { // is that port connected to a switch?

		r := findRouterByPortName(lport.RouterPort)
		if r == nil {
			log.Errorf("Unable to find router...")
		} else {
			lrp := r.ports[lport.RouterPort]

			log.Noticef("lsp '%s' is connected to lrp '%s'", port.Name, lrp.Name)

			if sw.swIomodule.PortsCount == 0 {
				sw.swIomodule.Deploy()
			}

			// attach both iomodules
			if err := iomodules.AttachIoModules(hc,
				sw.swIomodule, port.Name, r.rIoModule, lrp.Name); err != nil {

				log.Errorf("Unable attach router to switch")
			}

			if lrp.IP != "" {
				// configure router
				err := r.rIoModule.ConfigureInterface(lrp.Name, lrp.IP, lrp.Mask, lrp.Mac)
				if err != nil {
					log.Errorf("Error configuring router")
				}

				// the mac table of the switch is setted statically in order to
				// avoid problems with the broadcast.
				// (this issue will be solved soon)
				// Mac address of the router is present through this interface
				sw.swIomodule.AddForwardingTableEntry(lrp.Mac, port.Name)
			}
		}
	}

	lport.Modified = false
}

func findRouterByPortName(lrp string) *Router {
	for _, router := range routers {
		if _, ok := router.ports[lrp]; ok {
			return router
		}
	}

	return nil
}
