// Copyright 2017 Politecnico di Torino
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

	"github.com/mvbpolito/gosexy/to"

	"github.com/iovisor/iovisor-ovn/hover"
	"github.com/iovisor/iovisor-ovn/iomodules"
	l "github.com/op/go-logging"
)

var log = l.MustGetLogger("iomodules-nat")

type NatModule struct {
	ModuleId   string
	PortsCount int //number of allocated ports
	Interfaces map[string]*NatModuleInterface

	deployed bool
	hc       *hover.Client // used to send commands to hover
}

type NatModuleInterface struct {
	IfaceIdRedirectHover int    //Iface id inside hover (relative to the m:1234 the interface is attached to ...) and provided my the extended hover /links/ API
	IfaceFd              int    //Interface Fd inside External_Ids (42, etc...)
	LinkIdHover          string //iomodules Link Id
	IfaceName            string
}

func Create(hc *hover.Client) *NatModule {

	if hc == nil {
		log.Errorf("HoverClient is not valid")
		return nil
	}

	n := new(NatModule)
	n.Interfaces = make(map[string]*NatModuleInterface)
	n.hc = hc
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

	natError, natHover := n.hc.ModulePOST("bpf", "Nat", NatCode)
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

	moduleDeleteError, _ := n.hc.ModuleDELETE(n.ModuleId)
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

	linkError, linkHover := n.hc.LinkPOST("i:"+ifaceName, n.ModuleId)
	if linkError != nil {
		log.Errorf("Error in POSTing the Link: %s\n", linkError)
		return linkError
	}

	_, external_interfaces := n.hc.ExternalInterfacesListGET()

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
		errString := "Trying to detach port in undeployed nat"
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

	linkDeleteError, _ := n.hc.LinkDELETE(iface.LinkIdHover)

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

func (n *NatModule) SetPublicIp(ip net.IP) (err error) {
	if !n.deployed {
		errString := "undeployed nat"
		log.Errorf(errString)
		return errors.New(errString)
	}

	n.hc.TableEntryPUT(n.ModuleId, "public_ip", "0", iomodules.IpToHexBigEndian(ip.To4()))

	return nil
}

func (n *NatModule) DetachFromIoModule(ifaceName string) (err error) {
	return errors.New("Not implemented")
}

func (n *NatModule) Configure(conf interface{}) (err error) {

	log.Infof("Configuring NAT module")
	confMap := to.Map(conf)

	public_ip, ok1 := confMap["public_ip"]
	if !ok1 {
		return errors.New("Missing public_ip")
	}

	err = n.SetPublicIp(net.ParseIP(public_ip.(string)))
	if err != nil {
		return
	}

	return nil
}
