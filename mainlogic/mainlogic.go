package mainlogic

import (
	"github.com/mbertrone/politoctrl/bpf"
	"github.com/netgroup-polito/iovisor-ovn/hoverctl"
	"github.com/netgroup-polito/iovisor-ovn/ovnmonitor"
	l "github.com/op/go-logging"
)

var log = l.MustGetLogger("politoctrl")

func MainLogic(dataplane *hoverctl.Dataplane) {
	//Start monitoring ovn/s databases
	nbHandler := ovnmonitor.MonitorOvnNb()

	ovsHandler := ovnmonitor.MonitorOvsDb()

	go ovnmonitor.MonitorOvnSb()
	log.Debugf("%+v\n%+v\n", ovsHandler, nbHandler)

	globalHandler := ovnmonitor.HandlerHandler{}
	globalHandler.Nb = nbHandler
	globalHandler.Ovs = ovsHandler
	globalHandler.Dataplane = dataplane
	//Here I have to multiplex & demultiplex (maybe it's better if i use a final var or something like that.)

	//for now I only consider ovs notifications

	for {
		select {
		case ovsString := <-ovsHandler.MainLogicNotification:
			LogicalMapping(ovsString, &globalHandler)
		}
	}

}

func LogicalMapping(s string, hh *ovnmonitor.HandlerHandler) {
	log.Debugf("%s\n", s)

	for ifacename, iface := range hh.Ovs.OvsDatabase.Interface {
		if iface.Up {
			log.Debugf("(%s)IFACE UP -> DO NOTHING\n", ifacename)

		} else {
			log.Debugf("(%s)IFACE DOWN\n", ifacename)
			//Check if interface name belongs to some logical switch
			switchName := ovnmonitor.PortLookup(hh.Nb.NbDatabase, iface.IfaceId)
			log.Noticef("(%s) port||external-ids:iface-id(%s)-> SWITCH NAME: %s\n", iface.Name, iface.IfaceId, switchName)
			if switchName != "" {
				//log.Noticef("Switch:%s\n", switchName)
				if sw, ok := hh.Nb.NbDatabase.Logical_Switch[switchName]; ok {
					if sw.ModuleId == "" {
						log.Noticef("CREATE NEW SWITCH\n")

						_, swi := hoverctl.ModulePOST(hh.Dataplane, "bpf", "DummySwitch2", bpf.DummySwitch2)
						sw.ModuleId = swi.Id

						_, l1 := hoverctl.LinkPOST(hh.Dataplane, "i:"+iface.Name, swi.Id)
						log.Noticef("CREATE LINK from:%s to:%s id:%s\n", l1.From, l1.To, l1.Id)
						if l1.Id != "" {
							iface.Up = true
						}
						//Link port (in future lookup hypervisor)
					} else {
						//log.Debugf("SWITCH already present!%s\n", sw.ModuleId)
						//Only Link module
						_, l1 := hoverctl.LinkPOST(hh.Dataplane, "i:"+iface.Name, sw.ModuleId)
						log.Noticef("CREATE LINK from:%s to:%s id:%s\n", l1.From, l1.To, l1.Id) //TODO Check if crashes
						if l1.Id != "" {
							iface.Up = true
						}
					}
				} else {
					log.Errorf("No Switch in Nb referring to //%s//\n", switchName)
				}
			}
		}
	}

	//Main Logic for mapping iomodules
}
