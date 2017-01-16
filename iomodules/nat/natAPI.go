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
package nat

import (
	"errors"
	"fmt"
	"net"
	"strconv"

	"github.com/netgroup-polito/iovisor-ovn/hoverctl"
	l "github.com/op/go-logging"
)

var log = l.MustGetLogger("iomodules-nat")

type NatModule struct {
	ModuleId   string
	PortsCount int //number of allocated ports
	Interfaces map[string]*NatModuleInterface

	deployed  bool
	dataplane *hoverctl.Dataplane // used to send commands to hover
}

type NatModuleInterface struct {
	IfaceIdRedirectHover int    //Iface id inside hover (relative to the m:1234 the interface is attached to ...) and provided my the extended hover /links/ API
	IfaceFd              int    //Interface Fd inside External_Ids (42, etc...)
	LinkIdHover          string //iomodules Link Id
	IfaceName            string
}

func Create(dp *hoverctl.Dataplane) *NatModule {

	if dp == nil {
		log.Errorf("Daplane is not valid\n")
		return nil
	}

	n := new(NatModule)
	n.Interfaces = make(map[string]*NatModuleInterface)
	n.dataplane = dp
	n.deployed = false
	return n
}

func (n *NatModule) GetModuleId() string {
	return n.ModuleId
}

func (n *NatModule) Deploy() (err error) {

	if n.deployed {
		return nil
	}

	natError, natHover := hoverctl.ModulePOST(n.dataplane, "bpf",
		"Nat", NatCode)
	if natError != nil {
		log.Errorf("Error in POST Nat IOModule: %s\n", natError)
		return natError
	}

	log.Noticef("POST Nat IOModule %s\n", natHover.Id)
	n.ModuleId = natHover.Id
	n.deployed = true

	return nil
}

func (n *NatModule) Destroy() (err error) {

	if !n.deployed {
		return nil
	}

	// TODO:
	// All interfaces must be detached before destroying the module.
	// Should it be done automatically here, or should be the application responsible for that?

	moduleDeleteError, _ := hoverctl.ModuleDELETE(n.dataplane, n.ModuleId)
	if moduleDeleteError != nil {
		log.Errorf("Error in destrying Nat IOModule: %s\n", moduleDeleteError)
		return moduleDeleteError
	}

	n.ModuleId = ""
	n.deployed = false

	return nil
}

func (n *NatModule) AttachExternalInterface(ifaceName string) (err error) {

	if !n.deployed {
		errString := "Trying to attach port in undeployed nat"
		log.Errorf(errString)
		return errors.New(errString)
	}

	if n.PortsCount == 2 {
		errString := "There are not free ports in the nat\n"
		log.Errorf(errString)
		return errors.New(errString)
	}

	linkError, linkHover := hoverctl.LinkPOST(n.dataplane, "i:"+ifaceName, n.ModuleId)
	if linkError != nil {
		log.Errorf("Error in POSTing the Link: %s\n", linkError)
		return linkError
	}

	_, external_interfaces := hoverctl.ExternalInterfacesListGET(n.dataplane)

	n.PortsCount++

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

	iface := new(NatModuleInterface)

	iface.IfaceFd, _ = strconv.Atoi(external_interfaces[ifaceName].Id)
	iface.IfaceIdRedirectHover = ifacenumber
	iface.LinkIdHover = linkHover.Id
	iface.IfaceName = ifaceName

	n.Interfaces[ifaceName] = iface

	return nil
}

func (n *NatModule) DetachExternalInterface(ifaceName string) (err error) {

	if !n.deployed {
		errString := "Trying to detach port in undeployed switch"
		log.Errorf(errString)
		return errors.New(errString)
	}

	iface, ok := n.Interfaces[ifaceName]

	if !ok {
		errString := fmt.Sprintf("Iface '%s' is not present in nat '%s'\n",
			ifaceName, n.ModuleId)
		log.Warningf(errString)
		return errors.New(errString)
	}

	linkDeleteError, _ := hoverctl.LinkDELETE(n.dataplane, iface.LinkIdHover)

	if linkDeleteError != nil {
		//log.Debug("REMOVE Interface %s %s (1/1) LINK REMOVED\n", currentInterface.Name, currentInterface.IfaceIdExternalIds)
		log.Warningf("Problem removing iface '%s' from nat '%s'\n",
			ifaceName, n.ModuleId)
		return linkDeleteError
	}

	delete(n.Interfaces, ifaceName)

	return nil
}

func (n *NatModule) AttachToIoModule(ifaceId int, ifaceName string) (err error) {
	if !n.deployed {
		errString := "Trying to attach port in undeployed nat"
		log.Errorf(errString)
		return errors.New(errString)
	}

	if n.PortsCount == 2 {
		errString := "There are not free ports in the nat"
		log.Errorf(errString)
		return errors.New(errString)
	}

	iface := new(NatModuleInterface)

	iface.IfaceFd = -1
	iface.IfaceIdRedirectHover = ifaceId
	iface.LinkIdHover = ""
	iface.IfaceName = ifaceName

	n.Interfaces[ifaceName] = iface

	return nil
}

func (n *NatModule) SetPublicIp(ip string) (err error) {
	if !n.deployed {
		errString := "undeployed nat"
		log.Errorf(errString)
		return errors.New(errString)
	}

	hoverctl.TableEntryPUT(n.dataplane, n.ModuleId, "public_ip", "0", ipToHexadecimalString(ip))

	return nil
}

func (n *NatModule) DetachFromIoModule(ifaceName string) (err error) {
	return errors.New("Not implemented")
}

func ipToHexadecimalString(ip string) string {

	trial := net.ParseIP(ip)
	if trial.To4() != nil {
		ba := []byte(trial.To4())
		// log.Debugf("0x%02x%02x%02x%02x\n", ba[0], ba[1], ba[2], ba[3])
		ipv4HexStr := fmt.Sprintf("0x%02x%02x%02x%02x", ba[0], ba[1], ba[2], ba[3])
		return ipv4HexStr
	}

	return ""
}
