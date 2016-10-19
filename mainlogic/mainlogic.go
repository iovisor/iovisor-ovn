package mainlogic

import (
	"strconv"
	"time"

	"github.com/netgroup-polito/iovisor-ovn/bpf"
	"github.com/netgroup-polito/iovisor-ovn/config"
	"github.com/netgroup-polito/iovisor-ovn/hoverctl"
	"github.com/netgroup-polito/iovisor-ovn/ovnmonitor"
	l "github.com/op/go-logging"
)

const brint = "br-int"

var log = l.MustGetLogger("politoctrl")

func MainLogic(globalHandler *ovnmonitor.HandlerHandler) {
	//Start monitoring ovn/s databases
	nbHandler := ovnmonitor.MonitorOvnNb()
	ovsHandler := ovnmonitor.MonitorOvsDb()
	go ovnmonitor.MonitorOvnSb()
	//log.Debugf("%+v\n%+v\n", ovsHandler, nbHandler)

	globalHandler.Nb = nbHandler
	globalHandler.Ovs = ovsHandler
	//global.Hh = globalHandler
	//Here I have to multiplex & demultiplex (maybe it's better if i use a final var or something like that.)

	//Init databases
	globalHandler.Ovs.OvsDatabase = &ovnmonitor.Ovs_Database{}
	globalHandler.Ovs.OvsDatabase.Clear()

	//for now I only consider ovs notifications
	if ovsHandler != nil {
		for {
			select {
			case ovsString := <-ovsHandler.MainLogicNotification:
				LogicalMappingOvs(ovsString, globalHandler)
			}
		}
	} else {
		log.Errorf("MainLogic not started!\n")
	}
}

/*mapping events of ovs local db*/
func LogicalMappingOvs(s string, hh *ovnmonitor.HandlerHandler) {
	//log.Debugf("Ovs Event:%s\n", s)

	//NEW INTERFACES?
	//Iterate on OvsNewDatabase (There are New Interfaces?)
	for newIfaceName, newIface := range hh.Ovs.OvsNewDatabase.Interface {
		if newIfaceName != brint {
			if iface, ok := hh.Ovs.OvsDatabase.Interface[newIfaceName]; ok {
				//Interface already present
				//Check if there are changes

				//Check differences
				//.IfaceIdExternalIds
				if iface.IfaceIdExternalIds != newIface.IfaceIdExternalIds {
					iface.IfaceIdExternalIds = newIface.IfaceIdExternalIds
				}
				//log.Debugf("Interface MODIFIED: %+v\n", iface)
			} else {
				//Interface NOT present; ADD it
				ifc := ovnmonitor.Interface_Item{}
				ifc.Init()
				ifc.Name = newIface.Name
				ifc.IfaceIdExternalIds = newIface.IfaceIdExternalIds
				hh.Ovs.OvsDatabase.Interface[newIfaceName] = &ifc
				//log.Debugf("Interface ADDED: %+v\n", ifc)
			}
		}
	}

	//DELETED INTERFACES?
	//Iterate on OvsDatabase (There are Removed Interfaces?)
	for ifaceName, iface := range hh.Ovs.OvsDatabase.Interface {
		if ifaceName != brint {
			if _, ok := hh.Ovs.OvsNewDatabase.Interface[ifaceName]; ok {
				//nothing to do
			} else {
				//iface removed.
				//mark Interface as Removed (if not yet marked!).
				if !iface.ToRemove {
					iface.ToRemove = true
					//log.Debugf("Interface REMOVED (mark as ToRemove..): %+v\n", iface)
				}
			}
		}
	}

	//PROCESS CHANGES...
	//Process changes
	for currentInterfaceName, currentInterface := range hh.Ovs.OvsDatabase.Interface {
		//If Toremove -> REMOVE INTERFACE
		if currentInterface.ToRemove {
			//try to remove interface
			if currentInterface.LinkIdHover != "" {
				//log.Debug("Deleting Link: %s Interface: %s ExternalIds-IfaceId: %s ...\n", currentInterface.LinkIdHover, currentInterface.Name, currentInterface.IfaceIdExternalIds)
				linkDeleteError, _ := hoverctl.LinkDELETE(hh.Dataplane, currentInterface.LinkIdHover)

				if linkDeleteError == nil {
					//Complete the link deletion..
					currentInterface.LinkIdHover = ""
				}
				//else
				//this function will be automatically re-called at the next notification.
				//maybe sould be delayed with a fixed number of re-try and a Sleep between them.
			}
			if currentInterface.LinkIdHover == "" {
				//deletion of link successfully completed!

				//lookup in NbDatabase if there is a corresponding port in the bridge.
				//TODO if I update NbDb -> mark item as deleted and not remove the correspondence to IOModule & Links
				logicalSwitchName := ovnmonitor.PortLookup(hh.Nb.NbDatabase, currentInterface.IfaceIdExternalIds)
				if logicalSwitch, ok := hh.Nb.NbDatabase.Logical_Switch[logicalSwitchName]; ok {
					//manual broadcast stuff... cleanup after interface disconnect!
					if logicalSwitch.PortsArray[currentInterface.IfaceIdArrayBroadcast] != 0 {
						hoverctl.TableEntryPUT(hh.Dataplane, logicalSwitch.ModuleId, "ports", strconv.Itoa(currentInterface.IfaceIdArrayBroadcast), "0")
						//TODO if not successful retry
						logicalSwitch.PortsArray[currentInterface.IfaceIdArrayBroadcast] = 0
						logicalSwitch.PortsCount--
					}
					if currentInterface.SecurityMacString != "" {
						//TODO cleanup Security Policy
					}
					//TODO same cleanup for IP
				}
				delete(hh.Ovs.OvsDatabase.Interface, currentInterfaceName)
				//log.Debug("Deleting Link: %s Interface: %s ExternalIds-IfaceId: %s OK\n", currentInterface.LinkIdHover, currentInterface.Name, currentInterface.IfaceIdExternalIds)
			}
		} else {
			//If this Interface is NOT Toremove...
			//CONNECT INTERFACE TO IOMODULE
			if currentInterface.IfaceIdExternalIds != "" {
				//IF NOT already connected
				if !currentInterface.Up {
					//Proceed connect Interface to IOModule
					logicalSwitchName := ovnmonitor.PortLookup(hh.Nb.NbDatabase, currentInterface.IfaceIdExternalIds)
					if logicalSwitchName != "" {
						if logicalSwitch, ok := hh.Nb.NbDatabase.Logical_Switch[logicalSwitchName]; ok {
							//NO SWITCH POSTED
							if logicalSwitch.ModuleId == "" {
								log.Noticef("POST Switch IOModule ...\n")
								time.Sleep(config.SleepTime)

								switchError, switchHover := hoverctl.ModulePOST(hh.Dataplane, "bpf", "Switch8Security", bpf.SwitchSecurityPolicy)
								if switchError != nil {
									log.Errorf("Error in POST Switch IOModule: %s\n", switchError)
								} else {
									logicalSwitch.ModuleId = switchHover.Id
								}
							}
							if logicalSwitch.ModuleId != "" {
								//logicalSwitch.ModuleId !!!!

								time.Sleep(config.SleepTime)

								_, external_interfaces := hoverctl.ExternalInterfacesListGET(hh.Dataplane)
								portNumber := ovnmonitor.FindFirtsFreeLogicalPort(logicalSwitch)

								if portNumber == 0 {
									log.Warningf("Switch %s -> module %s : no free ports.\n", logicalSwitch.Name, logicalSwitch.ModuleId)
								} else {

									linkError, linkHover := hoverctl.LinkPOST(hh.Dataplane, "i:"+currentInterface.Name, logicalSwitch.ModuleId)
									if linkError != nil {
										log.Errorf("Error in POSTing the Link: %s\n", linkError)
									} else {
										log.Noticef("CREATE LINK from:%s to:%s id:%s\n", linkHover.From, linkHover.To, linkHover.Id)
										if linkHover.Id != "" {
											currentInterface.Up = true
										}

										//We are assuming that this process is made only once... If fails it could be a problem.

										//Configuring broadcast on the switch module
										currentInterface.IfaceIdArrayBroadcast = portNumber
										currentInterface.IfaceFd, _ = strconv.Atoi(external_interfaces[currentInterface.Name].Id)

										tablePutError, _ := hoverctl.TableEntryPUT(hh.Dataplane, logicalSwitch.ModuleId, "ports", strconv.Itoa(portNumber), external_interfaces[currentInterface.Name].Id)
										if tablePutError != nil {
											log.Warningf("Error in PUT entry into ports table... Probably problems with broadcast in the module. Error: %s\n", tablePutError)
										}
										logicalSwitch.PortsArray[portNumber] = currentInterface.IfaceFd
										logicalSwitch.PortsCount++

										//Saving IfaceIdRedirectHover for this port. The number will be used by security policies
										ifacenumber := -1
										if linkHover.From[0:2] == "m:" {
											ifacenumber = linkHover.FromId
										}
										if linkHover.To[0:2] == "m:" {
											ifacenumber = linkHover.ToId
										}
										if ifacenumber == -1 {
											log.Warningf("IfaceIdRedirectHover == -1 something wrong happend...\n")
										}
										currentInterface.IfaceIdRedirectHover = ifacenumber

										currentInterface.LinkIdHover = linkHover.Id
										//log.Debugf("link-id:%s\n", currentInterface.LinkIdHover)
										currentInterface.ToRemove = false
									}
								}
							}
						}
					}
				} //end iface NOT UP
				if currentInterface.Up {
					if config.SwitchSecurityPolicy {
						logicalSwitchName := ovnmonitor.PortLookup(hh.Nb.NbDatabase, currentInterface.IfaceIdExternalIds)
						if logicalSwitchName != "" {
							if logicalSwitch, ok := hh.Nb.NbDatabase.Logical_Switch[logicalSwitchName]; ok {
								lsp, ok := hh.Nb.NbDatabase.Logical_Switch_Port[currentInterface.IfaceIdExternalIds]
								if ok {
									//push or modify security policy?
									if lsp.SecurityMacStr != "" {
										//Security Policy Set
										if lsp.SecurityMacStr != currentInterface.SecurityMacString {
											errorTablePost, _ := hoverctl.TableEntryPOST(hh.Dataplane, logicalSwitch.ModuleId, "securitymac", strconv.Itoa(currentInterface.IfaceIdRedirectHover), lsp.SecurityMacStr)
											if errorTablePost == nil {
												currentInterface.SecurityMacString = lsp.SecurityMacStr
											}
										}
									} else {
										if currentInterface.SecurityMacString != "" {
											//delete security policy?
											errorDelete, _ := hoverctl.TableEntryDELETE(hh.Dataplane, logicalSwitch.ModuleId, "securitymac", strconv.Itoa(currentInterface.IfaceIdRedirectHover))
											if errorDelete == nil {
												currentInterface.SecurityMacString = ""
											}
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}
}
