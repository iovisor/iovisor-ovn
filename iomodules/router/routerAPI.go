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
package router

import (
	"errors"
	"fmt"
	"net"
	"sort"
	"strconv"

	"github.com/mvbpolito/gosexy/to"

	"github.com/iovisor/iovisor-ovn/iomodules"

	"github.com/iovisor/iovisor-ovn/hover"
	l "github.com/op/go-logging"
)

var log = l.MustGetLogger("iomodules-router")

type RouterModule struct {
	ModuleId          string
	RoutingTable      []RoutingTableEntry
	routingTableCount int // number of elements in the routing table
	Interfaces        map[string]*RouterModuleInterface
	OutputBuffer      map[uint32]*BufferQueue
	PktCounter        int

	deployed bool
	hc       *hover.Client // used to send commands to hover
}

type RouterModuleInterface struct {
	IfaceIdRedirectHover int    // Iface id inside hover
	LinkIdHover          string // iomodules Link Id
	IfaceName            string
	IP                   net.IP
	Netmask              net.IPMask
	MAC                  net.HardwareAddr
}

type RoutingTableEntry struct {
	network     net.IPNet
	outputIface *RouterModuleInterface
	nexthop     net.IP
}

func Create(hc *hover.Client) *RouterModule {

	if hc == nil {
		log.Errorf("HoverClient is not valid")
		return nil
	}

	r := new(RouterModule)
	r.Interfaces = make(map[string]*RouterModuleInterface)
	r.OutputBuffer = make(map[uint32]*BufferQueue)
	r.PktCounter = 0
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

	if len(r.Interfaces) == 32 {
		errString := "There are not free ports in the router\n"
		log.Errorf(errString)
		return errors.New(errString)
	}

	linkError, linkHover := r.hc.LinkPOST("i:"+ifaceName, r.ModuleId)
	if linkError != nil {
		log.Errorf("Error in POSTing the Link: %s\n", linkError)
		return linkError
	}

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

	if len(r.Interfaces) == 32 {
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
func (r *RouterModule) ConfigureInterface(ifaceName string, ip net.IP,
	netmask net.IPMask, mac net.HardwareAddr) (err error) {
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

	// configure port entry
	ifaceIdString := strconv.Itoa(iface.IfaceIdRedirectHover)
	ipString := iomodules.IpToHexBigEndian(ip)
	netmaskString := iomodules.NetmaskToHexBigEndian(netmask)
	macString := iomodules.MacToHexadecimalStringBigEndian(mac)

	toSend := ipString + " " + netmaskString + " " + macString

	r.hc.TableEntryPOST(r.ModuleId, "router_port",
		ifaceIdString, toSend)

	network := ip.Mask(netmask)

	net := net.IPNet{network, netmask}

	// add route for that port
	if r.AddRoutingTableEntryLocal(net, ifaceName) != nil {
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
func (r *RouterModule) AddRoutingTableEntryLocal(network net.IPNet,
	outputIface string) (err error) {
	return r.AddRoutingTableEntry(network, outputIface, net.ParseIP("0.0.0.0"))
}

// Routes in the routing table have to be ordered according to the netmask length.
// This is because of a limitation in the eBPF datapath (no way to perform LPM there)
// next hop is a string indicating the ip address of the nexthop, 0.0.0.0 indicates that
// the network is directly attached to the router.
func (r *RouterModule) AddRoutingTableEntry(network net.IPNet,
	outputIface string, nexthop net.IP) (err error) {

	log.Infof("add routing table entry: '%s' -> '%d'", network.String(), outputIface)

	if r.routingTableCount == 6 {
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
	neti := s[i].network.Mask
	netj := s[j].network.Mask

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
		if len(i.network.IP) == 0 {
			break
		}

		stringIndex := strconv.Itoa(index)

		toSend := iomodules.IpToHexBigEndian(i.network.IP) + " " +
			iomodules.NetmaskToHexBigEndian(i.network.Mask) + " " +
			"0x" + strconv.FormatUint(uint64(i.outputIface.IfaceIdRedirectHover), 16) + " " +
			iomodules.IpToHexBigEndian(i.nexthop)

		r.hc.TableEntryPOST(r.ModuleId, "routing_table", stringIndex, toSend)

		index++
	}

	return nil
}

// TODO: Implement this function
//func (r *RouterModule) RemoveRoutingTableEntry(network string) {
//
//}

func (r *RouterModule) AddArpEntry(ip net.IP, mac net.HardwareAddr) (err error) {
	if !r.deployed {
		errString := "Trying to add arp entry in undeployed router"
		log.Errorf(errString)
		return errors.New(errString)
	}

	r.hc.TableEntryPOST(r.ModuleId, "arp_table",
		iomodules.IpToHexBigEndian(ip),
		iomodules.MacToHexadecimalStringBigEndian(mac))

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
			ip_, ok2 := entryMap["ip"]
			netmask_, ok3 := entryMap["netmask"]
			mac_, ok4 := entryMap["mac"]

			if !ok1 || !ok2 || !ok3 || !ok4 {
				log.Errorf("Skipping non valid interface")
				continue
			}

			ip := net.ParseIP(ip_.(string))
			netmask := iomodules.ParseIPv4Mask(netmask_.(string))
			mac, err := net.ParseMAC(mac_.(string))
			if err != nil {
				return errors.New("'%s' is a not valid mac")
			}

			log.Infof("Configuring Interface '%s', '%s', '%s', '%s'",
				name.(string), ip.String(), netmask.String(), mac.String())

			err = r.ConfigureInterface(name.(string), ip, netmask, mac)
			if err != nil {
				return err
			}
		}
	}

	// configure static routes
	if static_routes, ok := confMap["static_routes"]; ok {
		for _, entry := range to.List(static_routes) {
			entryMap := to.Map(entry)

			network_, ok1 := entryMap["network"]
			netmask_, ok2 := entryMap["netmask"]
			interface_, ok3 := entryMap["interface"]
			next_hop_, ok4 := entryMap["next_hop"]

			if !ok1 || !ok2 || !ok3 {
				log.Errorf("Skipping non valid static route")
				continue
			}

			if !ok4 {
				next_hop_ = "0.0.0.0"
			}

			network := net.ParseIP(network_.(string))
			netmask := iomodules.ParseIPv4Mask(netmask_.(string))
			next_hop := net.ParseIP(next_hop_.(string))

			net := net.IPNet{network, netmask}

			log.Infof("Adding Static Route: '%s', '%s', '%s', '%s'",
				net.String(), interface_.(string), next_hop.String())

			err := r.AddRoutingTableEntry(net, interface_.(string), next_hop)
			if err != nil {
				return err
			}
		}
	}

	// configure arp entries
	if arp_entries, ok := confMap["arp_entries"]; ok {
		for _, entry := range to.List(arp_entries) {
			entryMap := to.Map(entry)

			ip_, ok1 := entryMap["ip"]
			mac_, ok2 := entryMap["mac"]

			if !ok1 || !ok2 {
				log.Errorf("Skipping non valid arp entry")
				continue
			}

			ip := net.ParseIP(ip_.(string))
			mac, err := net.ParseMAC(mac_.(string))
			if err != nil {
				return errors.New("no valid mac")
			}

			log.Infof("Adding arp entry Route: '%s' -> '%s'",
				ip.String(), mac.String())

			err = r.AddArpEntry(ip, mac)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
