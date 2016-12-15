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
	"strconv"
	"bytes"
	"errors"
	"fmt"

	"github.com/netgroup-polito/iovisor-ovn/config"
	"github.com/netgroup-polito/iovisor-ovn/hoverctl"
	l "github.com/op/go-logging"
)

var log = l.MustGetLogger("iomodules-switch")

type L2SwitchModule struct {
	ModuleId	string
	PortsArray	[config.SwitchPortsNumber + 1]int //[0]=empty [1..8]=contains the port allocation(with fd) for broadcast tricky implemented inside hover
	PortsCount	int                               //number of allocated ports

	Interfaces map[string]*L2SwitchModuleInterface

	deployed	bool
	dataplane	*hoverctl.Dataplane	// used to send commands to hover
}

type L2SwitchModuleInterface struct {
	IfaceIdRedirectHover  int    //Iface id inside hover (relative to the m:1234 the interface is attached to ...) and provided my the extended hover /links/ API
	IfaceIdArrayBroadcast int    //Interface Id in the array for broadcast (id->fd for broadcast)
	IfaceFd               int    //Interface Fd inside External_Ids (42, etc...)
	LinkIdHover           string //iomodules Link Id
	IfaceName             string
}

func Create(dp *hoverctl.Dataplane) *L2SwitchModule {

	if dp == nil {
		log.Errorf("Daplane is not valid\n")
		return nil
	}

	x := new(L2SwitchModule)
	x.Interfaces = make(map[string]*L2SwitchModuleInterface)
	x.dataplane = dp
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

	switchError, switchHover := hoverctl.ModulePOST(sw.dataplane, "bpf",
									"Switch", SwitchSecurityPolicy)
	if switchError != nil {
		log.Errorf("Error in POST Switch IOModule: %s\n", switchError)
		return switchError
	}

	log.Noticef("POST Switch IOModule %s\n", switchHover.Id)
	sw.ModuleId = switchHover.Id
	sw.deployed = true

	return nil
}

func (sw *L2SwitchModule) Destroy() (err error) {

	if !sw.deployed {
		return nil
	}

	// TODO:
	// All interfaces must be detached before destroying the module.
	// Should it be done automatically here, or should be the application responsible for that?

	moduleDeleteError, _ := hoverctl.ModuleDELETE(sw.dataplane, sw.ModuleId)
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

	linkError, linkHover := hoverctl.LinkPOST(sw.dataplane, "i:" + ifaceName, sw.ModuleId)
	if linkError != nil {
		log.Errorf("Error in POSTing the Link: %s\n", linkError)
		return linkError
	}

	portNumber := sw.FindFirstFreeLogicalPort()

	if portNumber == 0 {
		errString := fmt.Sprintf("Switch '%s': no free ports.\n", sw.ModuleId)
		log.Warningf(errString)
		return errors.New(errString)
	}

	_, external_interfaces := hoverctl.ExternalInterfacesListGET(sw.dataplane)

	// We are assuming that this process is made only once... If fails it could be a problem.

	iface := new(L2SwitchModuleInterface)

	// Configuring broadcast on the switch module
	iface.IfaceIdArrayBroadcast = portNumber
	iface.IfaceFd, _ = strconv.Atoi(external_interfaces[ifaceName].Id)

	tablePutError, _ := hoverctl.TableEntryPUT(sw.dataplane, sw.ModuleId, "ports",
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
		log.Warningf("IfaceIdRedirectHover == -1 something wrong happend...\n")
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

	linkDeleteError, _ := hoverctl.LinkDELETE(sw.dataplane, iface.LinkIdHover)

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
		hoverctl.TableEntryPUT(sw.dataplane, sw.ModuleId, "ports", strconv.Itoa(iface.IfaceIdArrayBroadcast), "0")
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
	iface.IfaceIdRedirectHover = ifaceId
	iface.IfaceName = ifaceName

	sw.Interfaces[ifaceName] = iface

	// TODO: security policies

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

func (sw *L2SwitchModule) FindFirstFreeLogicalPort() int {
	for i := 1; i < config.SwitchPortsNumber + 1; i++ {
		if sw.PortsArray[i] == 0 {
			return i
		}
	}
	return 0
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

	hoverctl.TableEntryPOST(sw.dataplane, sw.ModuleId, "fwdtable", macString,
		strconv.Itoa(swIface.IfaceIdRedirectHover))

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