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
package router

import (
	"bytes"
	"errors"
	"fmt"
	"net"
	"strconv"

	"github.com/mvbpolito/gosexy/to"

	//"github.com/netgroup-polito/iovisor-ovn/config"
	"github.com/netgroup-polito/iovisor-ovn/hoverctl"
	l "github.com/op/go-logging"
)

var log = l.MustGetLogger("iomodules-router")

type RouterModule struct {
	ModuleId     string
	PortsCount   int //number of allocated ports
	RoutingTable [10]RoutingTableEntry
	Interfaces   map[string]*RouterModuleInterface

	deployed  bool
	dataplane *hoverctl.Dataplane // used to send commands to hover
}

type RouterModuleInterface struct {
	IfaceIdRedirectHover int    //Iface id inside hover (relative to the m:1234 the interface is attached to ...) and provided my the extended hover /links/ API
	IfaceFd              int    //Interface Fd inside External_Ids (42, etc...)
	LinkIdHover          string //iomodules Link Id
	IfaceName            string
	IP                   string
	Netmask              string
	MAC                  string
}

type RoutingTableEntry struct {
	network string
	netmask string
	outputIface string
	nexthop string
}

func Create(dp *hoverctl.Dataplane) *RouterModule {

	if dp == nil {
		log.Errorf("Daplane is not valid\n")
		return nil
	}

	r := new(RouterModule)
	r.Interfaces = make(map[string]*RouterModuleInterface)
	r.dataplane = dp
	r.deployed = false
	return r
}

func (r *RouterModule) GetModuleId() string {
	return r.ModuleId
}

func (r *RouterModule) Deploy() (err error) {

	if r.deployed {
		return nil
	}

	routerError, routerHover := hoverctl.ModulePOST(r.dataplane, "bpf",
		"Router", RouterCode)
	if routerError != nil {
		log.Errorf("Error in POST Router IOModule: %s\n", routerError)
		return routerError
	}

	log.Noticef("POST Router IOModule %s\n", routerHover.Id)
	r.ModuleId = routerHover.Id
	r.deployed = true

	return nil
}

func (r *RouterModule) Destroy() (err error) {

	if !r.deployed {
		return nil
	}

	// TODO:
	// All interfaces must be detached before destroying the module.
	// Should it be done automatically here, or should be the application responsible for that?

	moduleDeleteError, _ := hoverctl.ModuleDELETE(r.dataplane, r.ModuleId)
	if moduleDeleteError != nil {
		log.Errorf("Error in destrying Router IOModule: %s\n", moduleDeleteError)
		return moduleDeleteError
	}

	r.ModuleId = ""
	r.deployed = false

	return nil
}

func (r *RouterModule) AttachExternalInterface(ifaceName string) (err error) {

	if !r.deployed {
		errString := "Trying to attach port in undeployed router"
		log.Errorf(errString)
		return errors.New(errString)
	}

	if r.PortsCount == 10 {
		errString := "There are not free ports in the router\n"
		log.Errorf(errString)
		return errors.New(errString)
	}

	linkError, linkHover := hoverctl.LinkPOST(r.dataplane, "i:"+ifaceName, r.ModuleId)
	if linkError != nil {
		log.Errorf("Error in POSTing the Link: %s\n", linkError)
		return linkError
	}

	_, external_interfaces := hoverctl.ExternalInterfacesListGET(r.dataplane)

	r.PortsCount++

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

	iface := new(RouterModuleInterface)

	iface.IfaceFd, _ = strconv.Atoi(external_interfaces[ifaceName].Id)
	iface.IfaceIdRedirectHover = ifacenumber
	iface.LinkIdHover = linkHover.Id
	iface.IfaceName = ifaceName

	r.Interfaces[ifaceName] = iface

	return nil
}

func (r *RouterModule) DetachExternalInterface(ifaceName string) (err error) {

	if !r.deployed {
		errString := "Trying to detach port in undeployed router"
		log.Errorf(errString)
		return errors.New(errString)
	}

	iface, ok := r.Interfaces[ifaceName]

	if !ok {
		errString := fmt.Sprintf("Iface '%s' is not present in router '%s'\n",
			ifaceName, r.ModuleId)
		log.Warningf(errString)
		return errors.New(errString)
	}

	linkDeleteError, _ := hoverctl.LinkDELETE(r.dataplane, iface.LinkIdHover)

	if linkDeleteError != nil {
		//log.Debug("REMOVE Interface %s %s (1/1) LINK REMOVED\n", currentInterface.Name, currentInterface.IfaceIdExternalIds)
		log.Warningf("Problem removing iface '%s' from router '%s'\n",
			ifaceName, r.ModuleId)
		return linkDeleteError
	}

	// remove port from list of ports
	ifaceIdString := strconv.Itoa(iface.IfaceIdRedirectHover)
	hoverctl.TableEntryDELETE(r.dataplane, r.ModuleId, "router_port", ifaceIdString)

	// TODO: remove static route entry
	delete(r.Interfaces, ifaceName)

	return nil
}

func (r *RouterModule) AttachToIoModule(ifaceId int, ifaceName string) (err error) {
	if !r.deployed {
		errString := "Trying to attach port in undeployed router"
		log.Errorf(errString)
		return errors.New(errString)
	}

	if r.PortsCount == 10 {
		errString := "There are not free ports in the router"
		log.Errorf(errString)
		return errors.New(errString)
	}

	iface := new(RouterModuleInterface)

	iface.IfaceFd = -1
	iface.IfaceIdRedirectHover = ifaceId
	iface.LinkIdHover = ""
	iface.IfaceName = ifaceName

	r.Interfaces[ifaceName] = iface

	return nil
}

func (r *RouterModule) DetachFromIoModule(ifaceName string) (err error) {
	return errors.New("Not implemented")
}

// After a interface has been added, it is necessary to configure it before
// it can be used to route packets
//TODO I think we have to add next hop parameter here!
func (r *RouterModule) ConfigureInterface(ifaceName string, ip string, netmask string, mac string) (err error) {
	if !r.deployed {
		errString := "Trying to configure an interface in undeployed router"
		log.Errorf(errString)
		return errors.New(errString)
	}

	iface, ok := r.Interfaces[ifaceName]

	if !ok {
		errString := fmt.Sprintf("Iface '%s' is not present in router '%s'\n",
			ifaceName, r.ModuleId)
		log.Warningf(errString)
		return errors.New(errString)
	}

	// TODO: check ip, netmask and mac

	iface.IP = ip
	iface.Netmask = netmask
	iface.MAC = mac

	// configure port entry
	ifaceIdString := strconv.Itoa(iface.IfaceIdRedirectHover)
	ipString := ipToHexadecimalString(ip)
	netmaskString := ipToHexadecimalString(netmask)
	macString := macToHexadecimalString(mac)

	toSend := ipString + " " + netmaskString + " " + macString

	hoverctl.TableEntryPOST(r.dataplane, r.ModuleId, "router_port",
		ifaceIdString, toSend)

	ip_ := net.ParseIP(ip)
	netmask_ := ParseIPv4Mask(netmask)
	network_ := ip_.Mask(netmask_)

	// add route for that port
	if !r.AddRoutingTableEntryLocal(network_.String(), netmask, ifaceName) {
		errString := fmt.Sprintf("Error adding static route for port '%s' in router '%s'\n",
			ifaceName, r.ModuleId)
		log.Warningf(errString)
		return errors.New(errString)
	}

	return nil
}

// A local entry of the routing table indicates that the network interface is
// directly attached, so there is no need of the next hop address.
// This function force the routing table entry to be local, pushing 0 as nexthop
func (r *RouterModule) AddRoutingTableEntryLocal(network string, netmask string, outputIface string) (error bool) {
	return r.AddRoutingTableEntry(network, netmask, outputIface, "0")
}

// With the current implementation of the eBPF router this functions is a kind
// of complicated.  This is because it has to add the routes in a sorte way in the
// table, routes with the longest prefix should appear first.
// However current implementation, it is the very first one, only adds the routes
// one after the other, without performing this sorting
// next hop is a string indicating the ip address of the nexthop, 0 if local iface
func (r *RouterModule) AddRoutingTableEntry(network string, netmask string, outputIface string, nexthop string) (error bool) {

	// look for a free entry in the routing table
	index := -1
	for i := 0; i < 10; i++ {
		if r.RoutingTable[i].network == "" {
			index = i
			break
		}
	}

	if index == -1 {
		log.Errorf("Routing table is full\n")
		return false
	}

	iface, ok := r.Interfaces[outputIface]

	if !ok {
		errString := fmt.Sprintf("Iface '%s' is not present in router '%s'\n",
			outputIface, r.ModuleId)
		log.Warningf(errString)
		return false
	}

	stringIndex := strconv.Itoa(index)
	toSend := "{" + ipToHexadecimalString(network) + " " +
		ipToHexadecimalString(netmask) + " " +
		strconv.Itoa(iface.IfaceIdRedirectHover) + " " + nexthop + "}"

	hoverctl.TableEntryPUT(r.dataplane, r.ModuleId, "routing_table",
		stringIndex, toSend)

	r.RoutingTable[index].network = network
	r.RoutingTable[index].netmask = netmask
	r.RoutingTable[index].outputIface = outputIface
	r.RoutingTable[index].nexthop = nexthop

	return true
}

// TODO: Implement this function
//func (r *RouterModule) RemoveRoutingTableEntry(network string) {
//
//}

func (r *RouterModule) Configure(conf interface{}) (err error) {
	// The interface is a map that contains:
	// interfaces: A list of maps for the interfaces present on the router, each
	// of this has to have:
	//		name, ip, netmask, mac
	// static_router: A lis of map containing the static routes to be configured
	// 		network: CIDR notation of the network
	//		next_hop: how to reach that network
	// FIXME: static routes interface is probably wrong
	log.Infof("Configuring Router")
	confMap := to.Map(conf)

	// configure interfaces
	if interfaces, ok := confMap["interfaces"]; ok {
		for _, entry := range to.List(interfaces) {
			entryMap := to.Map(entry)

			name, ok1 := entryMap["name"]
			ip, ok2 := entryMap["ip"]
			netmask, ok3 := entryMap["netmask"]
			mac, ok4 := entryMap["mac"]

			if !ok1 || !ok2 || !ok3 || !ok4 {
				log.Errorf("Skipping non valid interface")
				continue
			}

			log.Infof("Configuring Interface '%s', '%s', '%s', '%s'",
				name.(string), ip.(string), netmask.(string), mac.(string))

			err := r.ConfigureInterface(name.(string), ip.(string),
				netmask.(string), mac.(string))
			if err != nil {
				return err
			}
		}
	}

	// TODO: configure static routes

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

func ParseIPv4Mask(s string) net.IPMask {
	mask := net.ParseIP(s)
	if mask == nil {
		return nil
	}
	return net.IPv4Mask(mask[12], mask[13], mask[14], mask[15])
}
