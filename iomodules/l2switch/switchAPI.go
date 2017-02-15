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
package l2switch

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"

	"github.com/mvbpolito/gosexy/to"

	"github.com/iovisor/iovisor-ovn/config"
	"github.com/iovisor/iovisor-ovn/hover"
	l "github.com/op/go-logging"
)

var log = l.MustGetLogger("iomodules-switch")

type L2SwitchModule struct {
	ModuleId   string
	PortsArray [config.SwitchPortsNumber + 1]int // Saves the network interfaces file ids used to implement the broadcast
	PortsCount int                               // number of allocated ports

	Interfaces map[string]*L2SwitchModuleInterface

	deployed  bool
	hc *hover.Client // used to send commands to hover
}

type L2SwitchModuleInterface struct {
	IfaceIdRedirectHover  int    //Iface id inside hover (relative to the m:1234 the interface is attached to ...) and provided my the extended hover /links/ API
	IfaceIdArrayBroadcast int    //Interface Id in the array for broadcast (id->fd for broadcast)
	IfaceFd               int    //Interface Fd inside External_Ids (42, etc...)
	LinkIdHover           string //iomodules Link Id
	IfaceName             string
}

func Create(hc *hover.Client) *L2SwitchModule {

	if hc == nil {
		log.Errorf("HoverClient is not valid")
		return nil
	}

	x := new(L2SwitchModule)
	x.Interfaces = make(map[string]*L2SwitchModuleInterface)
	x.hc = hc
	x.deployed = false
	return x
}

func (sw *L2SwitchModule) GetModuleId() string {
	return sw.ModuleId
}

func (sw *L2SwitchModule) Deploy() (err error) {

	if sw.deployed {
		return nil
	}

	switchError, switchHover := sw.hc.ModulePOST("bpf",
		"Switch", SwitchSecurityPolicy)
	if switchError != nil {
		log.Errorf("Error in POST Switch IOModule: %s\n", switchError)
		return switchError
	}

	log.Noticef("POST Switch IOModule %s\n", switchHover.Id)
	sw.ModuleId = switchHover.Id
	sw.deployed = true

	id, _ := strconv.Atoi(sw.ModuleId[2:])
	sw.hc.GetController().RegisterCallBack(uint16(id), sw.ProcessPacket)

	return nil
}

func (sw *L2SwitchModule) Destroy() (err error) {

	if !sw.deployed {
		return nil
	}

	// TODO:
	// All interfaces must be detached before destroying the module.
	// Should it be done automatically here, or should be the application responsible for that?

	moduleDeleteError, _ := sw.hc.ModuleDELETE(sw.ModuleId)
	if moduleDeleteError != nil {
		log.Errorf("Error in destrying Switch IOModule: %s\n", moduleDeleteError)
		return moduleDeleteError
	}

	sw.ModuleId = ""
	sw.deployed = false

	return nil
}

func (sw *L2SwitchModule) AttachExternalInterface(ifaceName string) (err error) {

	if !sw.deployed {
		errString := "Trying to attach port in undeployed switch"
		log.Errorf(errString)
		return errors.New(errString)
	}

	linkError, linkHover := sw.hc.LinkPOST("i:"+ifaceName, sw.ModuleId)
	if linkError != nil {
		log.Errorf("Error in POSTing the Link: %s\n", linkError)
		return linkError
	}

	portNumber, err := sw.FindFirstFreeLogicalPort()
	if err != nil {
		log.Errorf("Error in finding free port: %s\n", err)
		return err
	}

	_, external_interfaces := sw.hc.ExternalInterfacesListGET()

	// We are assuming that this process is made only once... If fails it could be a problem.

	iface := new(L2SwitchModuleInterface)

	// Configuring broadcast on the switch module
	iface.IfaceIdArrayBroadcast = portNumber
	iface.IfaceFd, _ = strconv.Atoi(external_interfaces[ifaceName].Id)

	tablePutError, _ := sw.hc.TableEntryPUT(sw.ModuleId, "ports",
		strconv.Itoa(portNumber), external_interfaces[ifaceName].Id)
	if tablePutError != nil {
		log.Warningf("Error in PUT entry into ports table... ",
			"Probably problems with broadcast in the module. Error: %s\n", tablePutError)
		return tablePutError
	}

	sw.PortsArray[portNumber] = iface.IfaceFd
	sw.PortsCount++

	// Saving IfaceIdRedirectHover for this port. The number will be used by security policies
	ifacenumber := -1
	if linkHover.From[0:2] == "m:" {
		ifacenumber = linkHover.FromId
	}
	if linkHover.To[0:2] == "m:" {
		ifacenumber = linkHover.ToId
	}
	if ifacenumber == -1 {
		log.Warningf("IfaceIdRedirectHover == -1 something wrong happened...\n")
	}
	iface.IfaceIdRedirectHover = ifacenumber

	iface.LinkIdHover = linkHover.Id

	iface.IfaceName = ifaceName

	sw.Interfaces[ifaceName] = iface

	// TODO: security policies

	return nil
}

func (sw *L2SwitchModule) DetachExternalInterface(ifaceName string) (err error) {

	if !sw.deployed {
		errString := "Trying to detach port in undeployed switch"
		log.Errorf(errString)
		return errors.New(errString)
	}

	iface, ok := sw.Interfaces[ifaceName]

	if !ok {
		errString := fmt.Sprintf("Iface '%s' is not present in switch '%s'\n",
			ifaceName, sw.ModuleId)
		log.Warningf(errString)
		return errors.New(errString)
	}

	linkDeleteError, _ := sw.hc.LinkDELETE(iface.LinkIdHover)

	if linkDeleteError != nil {
		//log.Debug("REMOVE Interface %s %s (1/1) LINK REMOVED\n", currentInterface.Name, currentInterface.IfaceIdExternalIds)
		log.Warningf("Problem removing iface '%s' from switch '%s'\n",
			ifaceName, sw.ModuleId)
		return linkDeleteError
	}

	// Complete the link deletion...
	iface.LinkIdHover = ""

	// cleanup broadcast tables
	if sw.PortsArray[iface.IfaceIdArrayBroadcast] != 0 {
		sw.hc.TableEntryPUT(sw.ModuleId, "ports", strconv.Itoa(iface.IfaceIdArrayBroadcast), "0")
		// TODO: if not successful retry

		sw.PortsArray[iface.IfaceIdArrayBroadcast] = 0
		sw.PortsCount--
	}

	// TODO: clean up port security tables

	delete(sw.Interfaces, ifaceName)

	return nil
}

// This function is still experimental
// Adds an interface that is connected to another IOModule, the connection must
// be already been created by an external component.
// TODO: How to handle broadcast in this case?
func (sw *L2SwitchModule) AttachToIoModule(ifaceId int, ifaceName string) (err error) {
	if !sw.deployed {
		log.Errorf("Trying to attach port in undeployed switch\n")
		return errors.New("Trying to attach port in undeployed switch")
	}

	iface := new(L2SwitchModuleInterface)

	sw.PortsCount++

	portNumber := config.SwitchPortsNumber - 1

	iface.IfaceIdRedirectHover = ifaceId
	iface.IfaceName = ifaceName
	iface.IfaceIdArrayBroadcast = portNumber
	iface.IfaceFd = ifaceId

	sw.Interfaces[ifaceName] = iface

	// TODO: security policies

	tablePutError, _ := sw.hc.TableEntryPUT(sw.ModuleId, "ports",
		strconv.Itoa(portNumber), strconv.Itoa(ifaceId))
	if tablePutError != nil {
		log.Warningf("Error in PUT entry into ports table... ",
			"Probably problems with broadcast in the module. Error: %s\n", tablePutError)
		return tablePutError
	}

	sw.PortsArray[portNumber] = ifaceId

	return nil
}

// This is also experimental, same considerations as previous function should
// be considered
func (sw *L2SwitchModule) DetachFromIoModule(ifaceName string) (err error) {
	if !sw.deployed {
		log.Errorf("Trying to detach port in undeployed switch\n")
		return errors.New("Trying to detach port in undeployed switch")
	}

	_, ok := sw.Interfaces[ifaceName]

	if !ok {
		errString := fmt.Sprintf("Iface '%s' is not present in switch '%s'\n",
			ifaceName, sw.ModuleId)
		log.Warningf(errString)
		return errors.New(errString)
	}

	sw.PortsCount--

	// TODO: clean up port security tables

	delete(sw.Interfaces, ifaceName)
	return nil
}

func (sw *L2SwitchModule) FindFirstFreeLogicalPort() (portNumber int, err error) {
	for i := 0; i < config.SwitchPortsNumber; i++ {
		if sw.PortsArray[i] == 0 {
			return i, nil
		}
	}
	return -1, errors.New("No free port found")
}

// adds a entry in the forwarding table of the switch
// mac MUST be in the format xx:xx:xx:xx:xx:xx
func (sw *L2SwitchModule) AddForwardingTableEntry(mac string, ifaceName string) (err error) {

	swIface, ok := sw.Interfaces[ifaceName]
	if !ok {
		errString := fmt.Sprintf("Iface '%s' is not present in switch '%s'\n",
			ifaceName, sw.ModuleId)
		log.Warningf(errString)
		return errors.New(errString)
	}

	macString := "{" + macToHexadecimalString(mac) + "}"

	sw.hc.TableEntryPOST(sw.ModuleId, "fwdtable", macString,
		strconv.Itoa(swIface.IfaceIdRedirectHover))

	return nil
}

func (sw *L2SwitchModule) AddPortSecurityMac(mac string, ifaceName string) (err error) {

	swIface, ok := sw.Interfaces[ifaceName]
	if !ok {
		errString := fmt.Sprintf("Iface '%s' is not present in switch '%s'\n",
			ifaceName, sw.ModuleId)
		log.Warningf(errString)
		return errors.New(errString)
	}

	macString := macToHexadecimalString(mac)

	sw.hc.TableEntryPOST(sw.ModuleId, "securitymac",
		"{0x"+strconv.Itoa(swIface.IfaceIdRedirectHover)+"}", macString)
	return nil
}

func (sw *L2SwitchModule) Configure(conf interface{}) (err error) {
	// The interface is a map with the following elements:
	// forwarding_table: a list of maps, each one has:
	//		port: the port where mac can be reached
	//		mac: the mac itself
	// TODO: support for port security policies
	log.Infof("Configuring Switch")
	confMap := to.Map(conf)
	if fwd_table, ok := confMap["forwarding_table"]; ok {
		for _, entry := range to.List(fwd_table) {
			entryMap := to.Map(entry)

			port, ok1 := entryMap["port"]
			mac, ok2 := entryMap["mac"]
			if !ok1 || !ok2 {
				log.Errorf("Skipping non valid forwarding table entry")
				continue
			}

			log.Infof("Adding forwardig table entry '%s' -> '%s'",
				mac.(string), port.(string))

			err := sw.AddForwardingTableEntry(mac.(string), port.(string))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (sw *L2SwitchModule) ProcessPacket(p *hover.Packet) (err error) {
	_ = p

	log.Infof("Switch: '%s': Packet arrived from dataplane", sw.ModuleId)
	return nil
}

// TODO: this function should be smarter
func macToHexadecimalString(s string) string {
	var buffer bytes.Buffer

	buffer.WriteString("0x")
	buffer.WriteString(s[0:2])
	buffer.WriteString(s[3:5])
	buffer.WriteString(s[6:8])
	buffer.WriteString(s[9:11])
	buffer.WriteString(s[12:14])
	buffer.WriteString(s[15:17])

	return buffer.String()
}

// TODO: port security policies
