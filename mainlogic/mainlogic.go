package mainlogic

import (
	"strconv"

	"time"

	"github.com/netgroup-polito/iovisor-ovn/bpf"
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
	//log.Debugf("%+v\n%+v\n", ovsHandler, nbHandler)

	globalHandler := ovnmonitor.HandlerHandler{}
	globalHandler.Nb = nbHandler
	globalHandler.Ovs = ovsHandler
	globalHandler.Dataplane = dataplane
	//Here I have to multiplex & demultiplex (maybe it's better if i use a final var or something like that.)

	//for now I only consider ovs notifications
	if ovsHandler != nil {
		for {
			select {
			case ovsString := <-ovsHandler.MainLogicNotification:
				LogicalMappingOvs(ovsString, &globalHandler)
			}
		}

	} else {
		log.Errorf("MainLogic not started!\n")
	}
}

/*mapping events of ovs local db*/
func LogicalMappingOvs(s string, hh *ovnmonitor.HandlerHandler) {
	//log.Debugf("Ovs Event:%s\n", s)

	for ifaceName, iface := range hh.Ovs.OvsDatabase.Interface {
		if ifaceName != "br-int" {
			if iface.Up {
				//Up is a logical internal state in iovisor--ovn controller
				//it means that the corresponding iomodule is up and the interface connected
				//log.Debugf("(%s)IFACE UP -> DO NOTHING\n", ifaceName)
			} else {
				//log.Debugf("(%s)IFACE DOWN, not still mapped to an IOModule\n", ifaceName)
				//Check if interface name belongs to some logical switch
				logicalSwitchName := ovnmonitor.PortLookup(hh.Nb.NbDatabase, iface.IfaceId)
				//log.Noticef("(%s) port||external-ids:iface-id(%s)-> SWITCH NAME: %s\n", iface.Name, iface.IfaceId, logicalSwitchName)
				if logicalSwitchName != "" {
					//log.Noticef("Switch:%s\n", switchName)
					if logicalSwitch, ok := hh.Nb.NbDatabase.Logical_Switch[logicalSwitchName]; ok {
						if logicalSwitch.ModuleId == "" {
							log.Noticef("CREATE NEW SWITCH\n")

							time.Sleep(3500 * time.Millisecond)

							_, switchHover := hoverctl.ModulePOST(hh.Dataplane, "bpf", "Switch8", bpf.Switch)
							logicalSwitch.ModuleId = switchHover.Id

							_, linkHover := hoverctl.LinkPOST(hh.Dataplane, "i:"+iface.Name, switchHover.Id)
							log.Noticef("CREATE LINK from:%s to:%s id:%s\n", linkHover.From, linkHover.To, linkHover.Id)
							if linkHover.Id != "" {
								iface.Up = true
							}

							_, external_interfaces := hoverctl.ExternalInterfacesListGET(hh.Dataplane)

							portNumber := ovnmonitor.FindFirtsFreeLogicalPort(logicalSwitch)

							hoverctl.TableEntryPUT(hh.Dataplane, switchHover.Id, "ports", strconv.Itoa(portNumber), external_interfaces[iface.Name].Id)
							logicalSwitch.PortsArray[portNumber] = 1
							logicalSwitch.PortsCount++
							//Link port (in future lookup hypervisor)
						} else {
							//log.Debugf("SWITCH already present!%s\n", sw.ModuleId)
							//Only Link module

							time.Sleep(3500 * time.Millisecond)

							_, linkHover := hoverctl.LinkPOST(hh.Dataplane, "i:"+iface.Name, logicalSwitch.ModuleId)
							log.Noticef("CREATE LINK from:%s to:%s id:%s\n", linkHover.From, linkHover.To, linkHover.Id) //TODO Check if crashes
							if linkHover.Id != "" {
								iface.Up = true
							}
							_, external_interfaces := hoverctl.ExternalInterfacesListGET(hh.Dataplane)

							portNumber := ovnmonitor.FindFirtsFreeLogicalPort(logicalSwitch)

							hoverctl.TableEntryPUT(hh.Dataplane, logicalSwitch.ModuleId, "ports", strconv.Itoa(portNumber), external_interfaces[iface.Name].Id)
							logicalSwitch.PortsArray[portNumber] = 1
							logicalSwitch.PortsCount++

						}
					} else {
						log.Errorf("No Switch in Nb referring to //%s//\n", logicalSwitchName)
					}
				}
			}
		}
	}

	//Main Logic for mapping iomodules
}
