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
package onetoonenat

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"strconv"

	"github.com/mvbpolito/gosexy/to"

	"github.com/iovisor/iovisor-ovn/hover"
	l "github.com/op/go-logging"
)

var log = l.MustGetLogger("iomodules-nat")

type NatModule struct {
	ModuleId   string
	PortsCount int //number of allocated ports
	Interfaces map[string]*NatModuleInterface

	deployed  bool
	hc *hover.Client // used to send commands to hover
}

type NatModuleInterface struct {
	IfaceIdRedirectHover int    //Iface id inside hover (relative to the m:1234 the interface is attached to ...) and provided my the extended hover /links/ API
	IfaceFd              int    //Interface Fd inside External_Ids (42, etc...)
	LinkIdHover          string //iomodules Link Id
	IfaceName            string
}

func Create(hc *hover.Client) *NatModule {

	if hc == nil {
		log.Errorf("Dataplane is not valid")
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

	id, _ := strconv.Atoi(n.ModuleId[2:])
	n.hc.GetController().RegisterCallBack(uint16(id), n.ProcessPacket)

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

func (n *NatModule) SetAddressAssociation(internal_ip string, external_ip string) (err error) {
	if !n.deployed {
		errString := "undeployed nat"
		log.Errorf(errString)
		return errors.New(errString)
	}

	n.hc.TableEntryPOST(n.ModuleId, "egress_nat_table",
		"{"+ipToHexadecimalString(internal_ip)+"}", ipToHexadecimalString(external_ip))
	n.hc.TableEntryPOST(n.ModuleId, "reverse_nat_table",
		"{"+ipToHexadecimalString(external_ip)+"}", ipToHexadecimalString(internal_ip))

	return nil
}

func (n *NatModule) DetachFromIoModule(ifaceName string) (err error) {
	return errors.New("Not implemented")
}

func (n *NatModule) Configure(conf interface{}) (err error) {
	// conf is a map that contains:
	//		public_ip: ip that is put as src address on the ongoing packets

	log.Infof("Configuring NAT module")
	confMap := to.Map(conf)

	//configure nat_entries
	if nat_entries, ok := confMap["nat_entries"]; ok {
		for _, entry := range to.List(nat_entries) {
			entryMap := to.Map(entry)

			internal_ip, ok1 := entryMap["internal_ip"]
			if !ok1 {
				return errors.New("Missing internal_ip")
			}

			external_ip, ok1 := entryMap["external_ip"]
			if !ok1 {
				return errors.New("Missing external_ip")
			}

			err = n.SetAddressAssociation(internal_ip.(string), external_ip.(string))
			if err != nil {
				return
			}

		}
	}

	return nil
}

func (n *NatModule) ProcessPacket(p *hover.Packet) (err error) {
	_ = p

	log.Infof("OneToOneNat: '%s': Packet arrived from dataplane", n.ModuleId)
	return nil
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
