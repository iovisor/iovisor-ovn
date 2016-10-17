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

	//SWITCH WITH NO SECURITY POLICIES
	if !config.SwitchSecurityPolicy {
		for ifaceName, iface := range hh.Ovs.OvsDatabase.Interface {
			//if interface is not present into the local db, add it.
			if ifaceName != "br-int" {
				if iface.Up {
					//Up is a logical internal state in iovisor--ovn controller
					//it means that the corresponding iomodule is up and the interface connected
					//log.Debugf("(%s)IFACE UP -> DO NOTHING\n", ifaceName)
				} else {
					//log.Debugf("(%s)IFACE DOWN, not still mapped to an IOModule\n", ifaceName)
					//Check if interface name belongs to some logical switch
					logicalSwitchName := ovnmonitor.PortLookup(hh.Nb.NbDatabase, iface.IfaceIdExternalIds)
					//log.Noticef("(%s) port||external-ids:iface-id(%s)-> SWITCH NAME: %s\n", iface.Name, iface.IfaceId, logicalSwitchName)
					if logicalSwitchName != "" {
						//log.Noticef("Switch:%s\n", switchName)
						if logicalSwitch, ok := hh.Nb.NbDatabase.Logical_Switch[logicalSwitchName]; ok {
							if logicalSwitch.ModuleId == "" {
								log.Noticef("CREATE NEW SWITCH\n")

								time.Sleep(config.SleepTime)

								_, switchHover := hoverctl.ModulePOST(hh.Dataplane, "bpf", "Switch8", bpf.Switch)
								logicalSwitch.ModuleId = switchHover.Id

								_, external_interfaces := hoverctl.ExternalInterfacesListGET(hh.Dataplane)

								portNumber := ovnmonitor.FindFirtsFreeLogicalPort(logicalSwitch)

								if portNumber == 0 {
									log.Warningf("Switch %s -> module %s : no free ports.\n", logicalSwitch.Name, logicalSwitch.ModuleId)
								} else {

									linkError, linkHover := hoverctl.LinkPOST(hh.Dataplane, "i:"+iface.Name, switchHover.Id)
									if linkError != nil {
										log.Errorf("Error in POSTing the Link: %s\n", linkError)
									} else {
										log.Noticef("CREATE LINK from:%s to:%s id:%s\n", linkHover.From, linkHover.To, linkHover.Id)
										if linkHover.Id != "" {
											iface.Up = true
										}

										iface.IfaceNumber = portNumber
										iface.IfaceFd, _ = strconv.Atoi(external_interfaces[iface.Name].Id)

										hoverctl.TableEntryPUT(hh.Dataplane, switchHover.Id, "ports", strconv.Itoa(portNumber), external_interfaces[iface.Name].Id)
										logicalSwitch.PortsArray[portNumber] = iface.IfaceFd
										logicalSwitch.PortsCount++

										iface.LinkId = linkHover.Id
										log.Debugf("link-id:%s\n", iface.LinkId)
										iface.ToRemove = false
										//Link port (in future lookup hypervisor)
									}
								}
							} else {
								//log.Debugf("SWITCH already present!%s\n", sw.ModuleId)
								//Only Link module

								time.Sleep(config.SleepTime)

								_, external_interfaces := hoverctl.ExternalInterfacesListGET(hh.Dataplane)

								portNumber := ovnmonitor.FindFirtsFreeLogicalPort(logicalSwitch)

								if portNumber == 0 {
									log.Warningf("Switch %s -> module %s : no free ports.\n", logicalSwitch.Name, logicalSwitch.ModuleId)
								} else {
									linkError, linkHover := hoverctl.LinkPOST(hh.Dataplane, "i:"+iface.Name, logicalSwitch.ModuleId)
									if linkError != nil {
										log.Errorf("Error in POSTing the Link: %s\n", linkError)
									} else {
										log.Noticef("CREATE LINK from:%s to:%s id:%s\n", linkHover.From, linkHover.To, linkHover.Id) //TODO Check if crashes
										if linkHover.Id != "" {
											iface.Up = true
										}

										iface.IfaceNumber = portNumber
										iface.IfaceFd, _ = strconv.Atoi(external_interfaces[iface.Name].Id)

										hoverctl.TableEntryPUT(hh.Dataplane, logicalSwitch.ModuleId, "ports", strconv.Itoa(portNumber), external_interfaces[iface.Name].Id)
										logicalSwitch.PortsArray[portNumber] = iface.IfaceFd
										logicalSwitch.PortsCount++

										iface.LinkId = linkHover.Id
										log.Debugf("link-id:%s\n", iface.LinkId)
										iface.ToRemove = false
									}
								}
							}
						} else {
							log.Errorf("No Switch in Nb referring to //%s//\n", logicalSwitchName)
						}
					}
				}
			}
		}
	} else {
		//Switch with security policies
		for ifaceName, iface := range hh.Ovs.OvsDatabase.Interface {
			//if interface is not present into the local db, add it.
			if ifaceName != "br-int" {
				if iface.Up {
					//Up is a logical internal state in iovisor--ovn controller
					//it means that the corresponding iomodule is up and the interface connected
					//log.Debugf("(%s)IFACE UP -> DO NOTHING\n", ifaceName)
				} else {
					//log.Debugf("(%s)IFACE DOWN, not still mapped to an IOModule\n", ifaceName)
					//Check if interface name belongs to some logical switch
					logicalSwitchName := ovnmonitor.PortLookup(hh.Nb.NbDatabase, iface.IfaceIdExternalIds)
					//log.Noticef("(%s) port||external-ids:iface-id(%s)-> SWITCH NAME: %s\n", iface.Name, iface.IfaceId, logicalSwitchName)
					if logicalSwitchName != "" {
						//log.Noticef("Switch:%s\n", switchName)
						if logicalSwitch, ok := hh.Nb.NbDatabase.Logical_Switch[logicalSwitchName]; ok {
							if logicalSwitch.ModuleId == "" {
								log.Noticef("CREATE NEW SWITCH\n")

								time.Sleep(config.SleepTime)

								_, switchHover := hoverctl.ModulePOST(hh.Dataplane, "bpf", "Switch8Security", bpf.SwitchSecurityPolicy)
								logicalSwitch.ModuleId = switchHover.Id

								_, external_interfaces := hoverctl.ExternalInterfacesListGET(hh.Dataplane)

								portNumber := ovnmonitor.FindFirtsFreeLogicalPort(logicalSwitch)

								if portNumber == 0 {
									log.Warningf("Switch %s -> module %s : no free ports.\n", logicalSwitch.Name, logicalSwitch.ModuleId)
								} else {

									linkError, linkHover := hoverctl.LinkPOST(hh.Dataplane, "i:"+iface.Name, switchHover.Id)
									if linkError != nil {
										log.Errorf("Error in POSTing the Link: %s\n", linkError)
									} else {
										log.Noticef("CREATE LINK from:%s to:%s id:%s\n", linkHover.From, linkHover.To, linkHover.Id)
										if linkHover.Id != "" {
											iface.Up = true
										}

										//Push security Policy?
										//check if present security policy
										//(if not) remove if not already removed
										lsp, ok := hh.Nb.NbDatabase.Logical_Switch_Port[iface.IfaceIdExternalIds]
										if ok {
											//logical switch port found!

											if lsp.SecurityMacStr != "" {
												//Security Policy Set

												if lsp.SecurityMacStr != iface.SecurityMacString {
													//Security Mac Policy not yet injected to the module

													ifacenumber := -1
													if linkHover.From[0:2] == "m:" {
														ifacenumber = linkHover.FromId
													}
													if linkHover.To[0:2] == "m:" {
														ifacenumber = linkHover.ToId
													}

													iface.IfaceNumberHover = ifacenumber

													errorTablePost, _ := hoverctl.TableEntryPOST(hh.Dataplane, switchHover.Id, "securitymac", strconv.Itoa(ifacenumber), lsp.SecurityMacStr)

													if errorTablePost == nil {
														iface.SecurityMacString = lsp.SecurityMacStr
													}
												}
											} else {
												//No Security Policy

												if iface.SecurityMacString != "" {
													//No security policy injected
													errorTableDelete, _ := hoverctl.TableEntryDELETE(hh.Dataplane, switchHover.Id, "securitymac", strconv.Itoa(iface.IfaceNumberHover))

													if errorTableDelete == nil {
														iface.SecurityMacString = ""
													}
												}
											}

										}

										iface.IfaceNumber = portNumber
										iface.IfaceFd, _ = strconv.Atoi(external_interfaces[iface.Name].Id)

										hoverctl.TableEntryPUT(hh.Dataplane, switchHover.Id, "ports", strconv.Itoa(portNumber), external_interfaces[iface.Name].Id)
										logicalSwitch.PortsArray[portNumber] = iface.IfaceFd
										logicalSwitch.PortsCount++

										iface.LinkId = linkHover.Id
										log.Debugf("link-id:%s\n", iface.LinkId)
										iface.ToRemove = false
										//Link port (in future lookup hypervisor)
									}
								}
							} else {
								//log.Debugf("SWITCH already present!%s\n", sw.ModuleId)
								//Only Link module

								time.Sleep(config.SleepTime)

								_, external_interfaces := hoverctl.ExternalInterfacesListGET(hh.Dataplane)

								portNumber := ovnmonitor.FindFirtsFreeLogicalPort(logicalSwitch)

								if portNumber == 0 {
									log.Warningf("Switch %s -> module %s : no free ports.\n", logicalSwitch.Name, logicalSwitch.ModuleId)
								} else {
									linkError, linkHover := hoverctl.LinkPOST(hh.Dataplane, "i:"+iface.Name, logicalSwitch.ModuleId)
									if linkError != nil {
										log.Errorf("Error in POSTing the Link: %s\n", linkError)
									} else {
										log.Noticef("CREATE LINK from:%s to:%s id:%s\n", linkHover.From, linkHover.To, linkHover.Id) //TODO Check if crashes
										if linkHover.Id != "" {
											iface.Up = true
										}

										//Push security Policy?
										//check if present security policy
										//(if not) remove if not already removed
										lsp, ok := hh.Nb.NbDatabase.Logical_Switch_Port[iface.IfaceIdExternalIds]
										if ok {
											//logical switch port found!

											if lsp.SecurityMacStr != "" {
												//Security Policy Set

												if lsp.SecurityMacStr != iface.SecurityMacString {
													//Security Mac Policy not yet injected to the module

													ifacenumber := -1
													if linkHover.From[0:2] == "m:" {
														ifacenumber = linkHover.FromId
													}
													if linkHover.To[0:2] == "m:" {
														ifacenumber = linkHover.ToId
													}

													iface.IfaceNumberHover = ifacenumber

													errorTablePost, _ := hoverctl.TableEntryPOST(hh.Dataplane, logicalSwitch.ModuleId, "securitymac", strconv.Itoa(ifacenumber), lsp.SecurityMacStr)

													if errorTablePost == nil {
														iface.SecurityMacString = lsp.SecurityMacStr
													}
												}
											} else {
												//No Security Policy

												if iface.SecurityMacString != "" {
													//No security policy injected
													errorTableDelete, _ := hoverctl.TableEntryDELETE(hh.Dataplane, logicalSwitch.ModuleId, "securitymac", strconv.Itoa(iface.IfaceNumberHover))

													if errorTableDelete == nil {
														iface.SecurityMacString = ""
													}
												}
											}

										}

										iface.IfaceNumber = portNumber
										iface.IfaceFd, _ = strconv.Atoi(external_interfaces[iface.Name].Id)

										hoverctl.TableEntryPUT(hh.Dataplane, logicalSwitch.ModuleId, "ports", strconv.Itoa(portNumber), external_interfaces[iface.Name].Id)
										logicalSwitch.PortsArray[portNumber] = iface.IfaceFd
										logicalSwitch.PortsCount++

										iface.LinkId = linkHover.Id
										log.Debugf("link-id:%s\n", iface.LinkId)
										iface.ToRemove = false
									}
								}
							}
						} else {
							log.Errorf("No Switch in Nb referring to //%s//\n", logicalSwitchName)
						}
					}
				}
			}
		}
	}
	//if interface is present into the local db and not in ovs cache, delete it. (or mark it as deleted.)

	//little "hack". In future centralize this logic.
	var cache = *hh.Ovs.Cache

	table, _ := cache["Interface"]

	for ifaceName, iface := range hh.Ovs.OvsDatabase.Interface {
		//If iface.Toremove .. someone has taken in chargethe remove of the iface
		if !iface.ToRemove {
			found := false
			for _, row := range table {
				name := row.Fields["name"].(string)
				//log.Debugf("[ovs] %s -> [localdb] %s\n", name, ifaceName)
				if name == ifaceName {
					found = true
					break
				}
			}
			if !found {
				//I have removed an interface from Interfaces in local ovs bridge
				//mark this interface to remove (I take in charge the remove operation)
				iface.ToRemove = true
				log.Noticef("Interface removed: %s\n", ifaceName)

				//The interface correspond to a link in hover. So DELETE the link
				if iface.LinkId != "" {
					log.Debugf("link-id:%s\n", iface.LinkId)
					linkDeleteError, _ := hoverctl.LinkDELETE(hh.Dataplane, iface.LinkId)

					if linkDeleteError != nil {
						//log.Warningf("Failed to remove link. ToRemove = false to re-try delete\n")
						iface.ToRemove = false
						break
					}
				}

				//lookup in NbDatabase if there is a corresponding port in the bridge.
				//TODO if I update NbDb -> mark item as deleted and not remove the correspondence to IOModule & Links
				logicalSwitchName := ovnmonitor.PortLookup(hh.Nb.NbDatabase, iface.IfaceIdExternalIds)
				if logicalSwitch, ok := hh.Nb.NbDatabase.Logical_Switch[logicalSwitchName]; ok {
					hoverctl.TableEntryPUT(hh.Dataplane, logicalSwitch.ModuleId, "ports", strconv.Itoa(iface.IfaceNumber), "0")
					logicalSwitch.PortsArray[iface.IfaceNumber] = 0
					logicalSwitch.PortsCount--
				}
				delete(hh.Ovs.OvsDatabase.Interface, ifaceName)
			}
		}
	}
}
