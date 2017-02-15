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
	"sort"
	"strconv"

	"github.com/mvbpolito/gosexy/to"

	"github.com/iovisor/iovisor-ovn/hover"
	l "github.com/op/go-logging"
)

var log = l.MustGetLogger("iomodules-router")

type RouterModule struct {
	ModuleId          string
	PortsCount        int //number of allocated ports
	RoutingTable      []RoutingTableEntry
	routingTableCount int // number of elements in the routing table
	Interfaces        map[string]*RouterModuleInterface

	deployed  bool
	hc *hover.Client // used to send commands to hover
}

type RouterModuleInterface struct {
	IfaceIdRedirectHover int    // Iface id inside hover
	LinkIdHover          string // iomodules Link Id
	IfaceName            string
	IP                   string
	Netmask              string
	MAC                  string
}

type RoutingTableEntry struct {
	network     string
	netmask     string
	outputIface *RouterModuleInterface
	nexthop     string
}

func Create(hc *hover.Client) *RouterModule {

	if hc == nil {
		log.Errorf("HoverClient is not valid")
		return nil
	}

	r := new(RouterModule)
	r.Interfaces = make(map[string]*RouterModuleInterface)
	r.RoutingTable = make([]RoutingTableEntry, 10)
	r.hc = hc
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

	routerError, routerHover := r.hc.ModulePOST("bpf", "Router", RouterCode)
	if routerError != nil {
		log.Errorf("Error in POST Router IOModule: %s\n", routerError)
		return routerError
	}

	log.Noticef("POST Router IOModule %s\n", routerHover.Id)
	r.ModuleId = routerHover.Id
	r.deployed = true

	id, _ := strconv.Atoi(r.ModuleId[2:])
	r.hc.GetController().RegisterCallBack(uint16(id), r.ProcessPacket)

	return nil
}

func (r *RouterModule) Destroy() (err error) {

	if !r.deployed {
		return nil
	}

	// TODO:
	// All interfaces must be detached before destroying the module.
	// Should it be done automatically here, or should be the application responsible for that?

	moduleDeleteError, _ := r.hc.ModuleDELETE(r.ModuleId)
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

	linkError, linkHover := r.hc.LinkPOST("i:"+ifaceName, r.ModuleId)
	if linkError != nil {
		log.Errorf("Error in POSTing the Link: %s\n", linkError)
		return linkError
	}

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

	linkDeleteError, _ := r.hc.LinkDELETE(iface.LinkIdHover)

	if linkDeleteError != nil {
		log.Warningf("Problem removing iface '%s' from router '%s'\n",
			ifaceName, r.ModuleId)
		return linkDeleteError
	}

	// remove port from list of ports
	ifaceIdString := strconv.Itoa(iface.IfaceIdRedirectHover)
	r.hc.TableEntryDELETE(r.ModuleId, "router_port", ifaceIdString)

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
func (r *RouterModule) ConfigureInterface(ifaceName string, ip string,
	netmask string, mac string) (err error) {
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

	r.hc.TableEntryPOST(r.ModuleId, "router_port",
		ifaceIdString, toSend)

	ip_ := net.ParseIP(ip)
	netmask_ := ParseIPv4Mask(netmask)
	network_ := ip_.Mask(netmask_)

	// add route for that port
	if r.AddRoutingTableEntryLocal(network_.String(), netmask, ifaceName) != nil {
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
func (r *RouterModule) AddRoutingTableEntryLocal(network string, netmask string,
	outputIface string) (err error) {
	return r.AddRoutingTableEntry(network, netmask, outputIface, "0.0.0.0")
}

// Routes in the routing table have to be ordered according to the netmask length.
// This is because of a limitation in the eBPF datapath (no way to perform LPM there)
// next hop is a string indicating the ip address of the nexthop, 0.0.0.0 indicates that
// the network is directly attached to the router.
func (r *RouterModule) AddRoutingTableEntry(network string, netmask string,
	outputIface string, nexthop string) (err error) {

	if r.routingTableCount == 10 {
		return errors.New("Routing table is full")
	}

	index := r.routingTableCount

	iface, ok := r.Interfaces[outputIface]

	if !ok {
		errString := fmt.Sprintf("Iface '%s' is not present in router '%s'\n",
			outputIface, r.ModuleId)
		log.Warningf(errString)
		return errors.New(errString)
	}

	r.RoutingTable[index].network = network
	r.RoutingTable[index].netmask = netmask
	r.RoutingTable[index].outputIface = iface
	r.RoutingTable[index].nexthop = nexthop

	r.routingTableCount++

	r.sortRoutingTable()
	r.sendRoutingTable()

	return nil
}

// implement sort.Interface in order to use the sort function

type ByMaskLen []RoutingTableEntry

func (s ByMaskLen) Len() int {
	return len(s)
}

func (s ByMaskLen) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s ByMaskLen) Less(i, j int) bool {
	neti := ParseIPv4Mask(s[i].netmask)
	netj := ParseIPv4Mask(s[j].netmask)

	si, _ := neti.Size()
	sj, _ := netj.Size()

	return si > sj
}

// sort routing table entries according to the length of the netmas, the longest
// ones are first
func (r *RouterModule) sortRoutingTable() {
	sort.Sort(ByMaskLen(r.RoutingTable))
}

func (r *RouterModule) sendRoutingTable() (err error) {

	index := 0
	for _, i := range r.RoutingTable {
		if i.network == "" {
			break
		}

		stringIndex := strconv.Itoa(index)

		toSend := "{" + ipToHexadecimalString(i.network) + " " +
			ipToHexadecimalString(i.netmask) + " " +
			strconv.Itoa(i.outputIface.IfaceIdRedirectHover) + " " +
			ipToHexadecimalString(i.nexthop) + "}"

		r.hc.TableEntryPUT(r.ModuleId, "routing_table",
			stringIndex, toSend)

		index++
	}

	return nil
}

// TODO: Implement this function
//func (r *RouterModule) RemoveRoutingTableEntry(network string) {
//
//}

func (r *RouterModule) AddArpEntry(ip string, mac string) (err error) {
	if !r.deployed {
		errString := "Trying to add arp entry in undeployed router"
		log.Errorf(errString)
		return errors.New(errString)
	}

	r.hc.TableEntryPUT(r.ModuleId, "arp_table",
		ipToHexadecimalString(ip), macToHexadecimalString(mac))

	return nil
}

func (r *RouterModule) Configure(conf interface{}) (err error) {
	// The interface is a map that contains:
	// interfaces: A list of maps for the interfaces present on the router, each
	// of this has to have:
	//		name, ip, netmask, mac
	// static_routes: A list of map containing the static routes to be configured
	// 		network: network address
	//		netmask: netmask for network
	//		interface: output interface where destination can be reached
	//		next_hop: [optional] ip address of next hop
	// arp_entries: A list of static arp entries to be configured
	//		ip:
	//		mac:
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

	// configure static routes
	if static_routes, ok := confMap["static_routes"]; ok {
		for _, entry := range to.List(static_routes) {
			entryMap := to.Map(entry)

			network, ok1 := entryMap["network"]
			netmask, ok2 := entryMap["netmask"]
			interface_, ok3 := entryMap["interface"]
			next_hop, ok4 := entryMap["next_hop"]

			if !ok1 || !ok2 || !ok3 {
				log.Errorf("Skipping non valid static route")
				continue
			}

			if !ok4 {
				next_hop = "0.0.0.0"
			}

			log.Infof("Adding Static Route: '%s', '%s', '%s', '%s'",
				network.(string), netmask.(string), interface_.(string), next_hop.(string))

			err := r.AddRoutingTableEntry(network.(string), netmask.(string),
				interface_.(string), next_hop.(string))
			if err != nil {
				return err
			}
		}
	}

	// configure arp entries
	if arp_entries, ok := confMap["arp_entries"]; ok {
		for _, entry := range to.List(arp_entries) {
			entryMap := to.Map(entry)

			ip, ok1 := entryMap["ip"]
			mac, ok2 := entryMap["mac"]

			if !ok1 || !ok2 {
				log.Errorf("Skipping non valid arp entry")
				continue
			}

			log.Infof("Adding arp entry Route: '%s' -> '%s'",
				ip.(string), mac.(string))

			err := r.AddArpEntry(ip.(string), mac.(string))
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (r *RouterModule) ProcessPacket(p *hover.Packet) (err error) {
	_ = p

	log.Infof("Router: '%s': Packet arrived from dataplane", r.ModuleId)
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
