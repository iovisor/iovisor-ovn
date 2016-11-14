package mainlogic

import (
	"strconv"
	"time"

	"github.com/netgroup-polito/iovisor-ovn/config"
	"github.com/netgroup-polito/iovisor-ovn/hoverctl"
	"github.com/netgroup-polito/iovisor-ovn/iomodules/l2switch"
	"github.com/netgroup-polito/iovisor-ovn/ovnmonitor"
	l "github.com/op/go-logging"
)

const brint = "br-int"

var log = l.MustGetLogger("iovisor-ovn-daemon")

func MainLogic(globalHandler *ovnmonitor.HandlerHandler) {
	//Start monitoring ovn/s databases
	nbHandler := ovnmonitor.MonitorOvnNb()
	ovsHandler := ovnmonitor.MonitorOvsDb()

	go ovnmonitor.MonitorOvnSb()
	//log.Debugf("%+v\n%+v\n", ovsHandler, nbHandler)

	globalHandler.Nb = nbHandler
	globalHandler.Ovs = ovsHandler

	//Init databases
	globalHandler.Nb.NbDatabase = &ovnmonitor.Nb_Database{}
	globalHandler.Nb.NbDatabase.Clear()

	globalHandler.Ovs.OvsDatabase = &ovnmonitor.Ovs_Database{}
	globalHandler.Ovs.OvsDatabase.Clear()

	go FlushCache(globalHandler)

	//for now I only consider ovs notifications
	if ovsHandler != nil {
		for {
			select {
			case ovsString := <-ovsHandler.MainLogicNotification:
				LogicalMappingOvs(ovsString, globalHandler)
			case nbString := <-nbHandler.MainLogicNotification:
				LogicalMappingNb(nbString, globalHandler)
			}
		}
	} else {
		log.Errorf("MainLogic not started!\n")
	}
}

//TODO Check. I don't know if something could be wrong...
//MAYBE notifcation to the parse logic looks better...
func FlushCache(globalHandler *ovnmonitor.HandlerHandler) {
	if config.FlushEnabled {
		for {
			time.Sleep(config.FlushTime)
			log.Debugf("FLUSHING Northbound......\n")
			globalHandler.Nb.MainLogicNotification <- "FlushNb"
			log.Debugf("FLUSHING Ovs Database......\n")
			globalHandler.Ovs.MainLogicNotification <- "FlushOvs"
		}
	}
}

func LogicalMappingNb(s string, hh *ovnmonitor.HandlerHandler) {
	//log.Debugf("Nb Event:%s\n", s)

	//Read Lock on Northbound Db
	hh.Nb.RWMutex.RLock()
	defer hh.Nb.RWMutex.RUnlock()

	//NEW MODULE?

	for newLogicalSwitchName, newLogicalSwitch := range hh.Nb.NbNewDatabase.Logical_Switch {
		if logicalSwitch, ok := hh.Nb.NbDatabase.Logical_Switch[newLogicalSwitchName]; ok {
			//logicalSwitch already present

			logicalSwitch.ToRemove = false

			//if new ports are added, add them
			for newPortUUID, _ := range newLogicalSwitch.PortsUUID {
				logicalSwitch.PortsUUID[newPortUUID] = newPortUUID
			}

			//TODO Check compatibility with PortLookup!!!
			//if ports are deleted. Delete them
			for portUUID, _ := range logicalSwitch.PortsUUID {
				if _, ok := newLogicalSwitch.PortsUUID[portUUID]; ok {
					//it's ok
				} else {
					//TODO NOT WORKING
					// log.Debugf("Deleting %s\n", portUUID)
					delete(logicalSwitch.PortsUUID, portUUID)
				}
			}
		} else {
			//logical switch not present
			logicalSwitch := ovnmonitor.Logical_Switch_Item{}
			logicalSwitch.Init()
			logicalSwitch.Name = newLogicalSwitch.Name
			//if new ports are added, add them
			for newPortUUID, _ := range newLogicalSwitch.PortsUUID {
				logicalSwitch.PortsUUID[newPortUUID] = newPortUUID
			}
			hh.Nb.NbDatabase.Logical_Switch[newLogicalSwitchName] = &logicalSwitch
		}
	}

	//DELETE MODULE?
	for _, logicalSwitch := range hh.Nb.NbDatabase.Logical_Switch {
		if _, ok := hh.Nb.NbNewDatabase.Logical_Switch[logicalSwitch.Name]; ok {
			//it's ok
		} else {
			// log.Debugf("MARK ToRemove: logicalSwitchName %s \n", logicalSwitch.Name)
			//delete the current switch!
			logicalSwitch.ToRemove = true

			//remove all the ports.
			//HACK but must work!
			for logicalSwitchPortName, _ := range logicalSwitch.PortsUUID {
				// log.Debugf("Deleting %s\n", logicalSwitchPortName)
				delete(logicalSwitch.PortsUUID, logicalSwitchPortName)
			}
		}
	}

	//NEW PORT?
	for newLogicalPortName, newLogicalSwitchPort := range hh.Nb.NbNewDatabase.Logical_Switch_Port {
		if logicalSwitchPort, ok := hh.Nb.NbDatabase.Logical_Switch_Port[newLogicalPortName]; ok {
			//check modified fields
			// log.Debugf("B logicalSwitchPort %s LogicalSwitchName %s\n", logicalSwitchPort.Name, logicalSwitchPort.LogicalSwitchName)
			if logicalSwitchPort.LogicalSwitchName == "" {
				logicalSwitchPort.LogicalSwitchName = ovnmonitor.PortLookupNoCached(hh.Nb.NbNewDatabase, logicalSwitchPort.Name)
				// log.Debugf("A logicalSwitchPort %s LogicalSwitchName %s\n", logicalSwitchPort.Name, logicalSwitchPort.LogicalSwitchName)
				//log.Debugf("RETRY PortLookupNoCached %s -> %s", logicalSwitchPort.Name, logicalSwitchPort.LogicalSwitchName)
				if logicalSwitchPort.LogicalSwitchName == "" {
					log.Warningf("Logical switch Lookup not found for port %s UUID %s\n", logicalSwitchPort.Name, logicalSwitchPort.UUID)
				}
			}

			if logicalSwitchPort.Addresses != newLogicalSwitchPort.Addresses {
				logicalSwitchPort.Addresses = newLogicalSwitchPort.Addresses
			}
			if logicalSwitchPort.PortSecutiry != newLogicalSwitchPort.PortSecutiry {
				logicalSwitchPort.PortSecutiry = newLogicalSwitchPort.PortSecutiry
				//compute SecurityMacStr
				if logicalSwitchPort.PortSecutiry != "" {
					logicalSwitchPort.SecurityMacStr = ovnmonitor.FromPortSecurityStrToMacStr(logicalSwitchPort.PortSecutiry)
					logicalSwitchPort.SecurityIpStr = ovnmonitor.FromPortSecurityStrToIpStr(logicalSwitchPort.PortSecutiry)
					hh.Ovs.MainLogicNotification <- "TestNotificationForSecurityPoliciesUPDATE"
					// log.Noticef("MAC:%s\n", logicalSwitchPort.SecurityMacStr)
				}
				//TODO compute SecurityIpStr
			}

		} else {
			//add a new Logical Port at all
			logicalSwitchPort := ovnmonitor.Logical_Switch_Port_Item{}
			logicalSwitchPort.Init()
			logicalSwitchPort.UUID = newLogicalSwitchPort.UUID
			logicalSwitchPort.Name = newLogicalSwitchPort.Name
			logicalSwitchPort.Addresses = newLogicalSwitchPort.Addresses
			logicalSwitchPort.PortSecutiry = newLogicalSwitchPort.PortSecutiry
			logicalSwitchPort.LogicalSwitchName = ovnmonitor.PortLookupNoCached(hh.Nb.NbNewDatabase, logicalSwitchPort.Name)
			//log.Debugf("PortLookupNoCached %s -> %s", logicalSwitchPort.Name, logicalSwitchPort.LogicalSwitchName)
			if logicalSwitchPort.LogicalSwitchName == "" {
				//log.Infof("Logical switch Lookup not found for port %s UUID %s\n", logicalSwitchPort.Name, logicalSwitchPort.UUID)
			}

			//compute SecurityMacStr
			if logicalSwitchPort.PortSecutiry != "" {
				logicalSwitchPort.SecurityMacStr = ovnmonitor.FromPortSecurityStrToMacStr(logicalSwitchPort.PortSecutiry)
				logicalSwitchPort.SecurityIpStr = ovnmonitor.FromPortSecurityStrToIpStr(logicalSwitchPort.PortSecutiry)
				hh.Ovs.MainLogicNotification <- "TestNotificationForSecurityPoliciesUPDATE"
			}

			hh.Nb.NbDatabase.Logical_Switch_Port[logicalSwitchPort.Name] = &logicalSwitchPort
		}
	}

	//DELETE PORT?
	for logicalPortName, logicalPort := range hh.Nb.NbDatabase.Logical_Switch_Port {
		if _, ok := hh.Nb.NbNewDatabase.Logical_Switch_Port[logicalPortName]; ok {
			//it's ok
		} else {
			//I deleted the logical switch portStr
			logicalPort.ToRemove = true
			//log.Debugf("MARK To Remove: logicalPort %s\n", logicalPort.Name)
		}
	}

	//PROCESS CHANGES.. what changes??
	for _, logicalSwitch := range hh.Nb.NbDatabase.Logical_Switch {
		if logicalSwitch.ToRemove {
			if logicalSwitch.PortsCount == 0 {
				if logicalSwitch.ModuleId != "" {
					moduleDeleteError, _ := hoverctl.ModuleDELETE(hh.Dataplane, logicalSwitch.ModuleId)
					if moduleDeleteError == nil {
						logicalSwitch.ModuleId = ""
						delete(hh.Nb.NbDatabase.Logical_Switch, logicalSwitch.Name)
					}
				}
			} else {
				log.Debugf("trying to remove logical switch %s module: %s with %d ports still active. Fail.\n", logicalSwitch.Name, logicalSwitch.ModuleId, logicalSwitch.PortsCount)
			}
		}
	}

	for logicalSwitchPortName, logicalSwitchPort := range hh.Nb.NbDatabase.Logical_Switch_Port {
		if logicalSwitchPort.ToRemove {
			//TODO Manage this fiels into ovs main logic!!!
			if logicalSwitchPort.InterfaceReference == "" {
				delete(hh.Nb.NbDatabase.Logical_Switch_Port, logicalSwitchPortName)
			}
		}
	}

}

/*mapping events of ovs local db*/
func LogicalMappingOvs(s string, hh *ovnmonitor.HandlerHandler) {
	//log.Debugf("Ovs Event:%s\n", s)

	//Read Lock on Northbound Db
	hh.Ovs.RWMutex.RLock()
	defer hh.Ovs.RWMutex.RUnlock()

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
					//TODO
					ovnmonitor.DeleteInterfaceReference(hh.Nb.NbDatabase, iface.IfaceIdExternalIds)
					iface.IfaceIdExternalIds = newIface.IfaceIdExternalIds
					ovnmonitor.SetInterfaceReference(hh.Nb.NbDatabase, iface.IfaceIdExternalIds, iface.Name)
				}
				//log.Debugf("Interface MODIFIED: %+v\n", iface)
			} else {
				//Interface NOT present; ADD it
				ifc := ovnmonitor.Interface_Item{}
				ifc.Init()
				ifc.Name = newIface.Name
				ifc.IfaceIdExternalIds = newIface.IfaceIdExternalIds
				if ifc.IfaceIdExternalIds != "" {
					//TODO
					ovnmonitor.SetInterfaceReference(hh.Nb.NbDatabase, ifc.IfaceIdExternalIds, ifc.Name)
				}
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
					ovnmonitor.DeleteInterfaceReference(hh.Nb.NbDatabase, iface.IfaceIdExternalIds)
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
					//log.Debug("REMOVE Interface %s %s (1/1) LINK REMOVED\n", currentInterface.Name, currentInterface.IfaceIdExternalIds)
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
						//log.Debugf("REMOVE Interface %s %s (2/2) NB IfaceIdArrayBroadcast\n", currentInterface.Name, currentInterface.IfaceIdExternalIds)

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
								// log.Noticef("POST Switch IOModule\n")
								time.Sleep(config.SleepTime)

								switchError, switchHover := hoverctl.ModulePOST(hh.Dataplane, "bpf", "Switch", l2switch.SwitchSecurityPolicy)
								if switchError != nil {
									log.Errorf("Error in POST Switch IOModule: %s\n", switchError)
								} else {
									log.Noticef("POST Switch IOModule %s\n", switchHover.Id)
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
									//Mac security
									if lsp.SecurityMacStr != "" {
										//Security Policy Mac Set
										if lsp.SecurityMacStr != currentInterface.SecurityMacString {
											errorTablePost, _ := hoverctl.TableEntryPOST(hh.Dataplane, logicalSwitch.ModuleId, "securitymac", strconv.Itoa(currentInterface.IfaceIdRedirectHover), lsp.SecurityMacStr)
											if errorTablePost == nil {
												currentInterface.SecurityMacString = lsp.SecurityMacStr
											}
										}
									} else {
										if currentInterface.SecurityMacString != "" {
											//delete security policy?
											errorDelete, _ := hoverctl.TableEntryDELETE(hh.Dataplane, logicalSwitch.ModuleId, "securitymac", "{"+strconv.Itoa(currentInterface.IfaceIdRedirectHover)+"}")
											if errorDelete == nil {
												currentInterface.SecurityMacString = ""
											}
										}
									}
									//Ip security
									if lsp.SecurityIpStr != "" {
										//Security Policy Ip Set
										if lsp.SecurityIpStr != currentInterface.SecurityIpString {
											errorTablePost, _ := hoverctl.TableEntryPOST(hh.Dataplane, logicalSwitch.ModuleId, "securityip", strconv.Itoa(currentInterface.IfaceIdRedirectHover), lsp.SecurityIpStr)
											if errorTablePost == nil {
												currentInterface.SecurityIpString = lsp.SecurityIpStr
											}
										}
									} else {
										if currentInterface.SecurityIpString != "" {
											//delete security policy?
											errorDelete, _ := hoverctl.TableEntryDELETE(hh.Dataplane, logicalSwitch.ModuleId, "securityip", "{"+strconv.Itoa(currentInterface.IfaceIdRedirectHover)+"}")
											if errorDelete == nil {
												currentInterface.SecurityIpString = ""
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
