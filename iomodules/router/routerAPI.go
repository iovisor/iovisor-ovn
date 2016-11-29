package router

import (
	"strconv"
	"bytes"
	"net"
	"fmt"

	//"github.com/netgroup-polito/iovisor-ovn/config"
	"github.com/netgroup-polito/iovisor-ovn/hoverctl"
	l "github.com/op/go-logging"
)

var log = l.MustGetLogger("iomodules-switch")

type RouterModule struct {
	ModuleId	string
	PortsCount	int                               //number of allocated ports
	PortsArray	[10]bool // This is used to keep track of available index in
						 // the router_port map of the switch
	RoutingTable [10]RoutingTableEntry
	Interfaces map[string]*RouterModuleInterface

	deployed	bool
	dataplane	*hoverctl.Dataplane	// used to send commands to hover
}

type RouterModuleInterface struct {
	IfaceIdRedirectHover  int    //Iface id inside hover (relative to the m:1234 the interface is attached to ...) and provided my the extended hover /links/ API
	IfaceFd               int    //Interface Fd inside External_Ids (42, etc...)
	LinkIdHover           string //iomodules Link Id
	IfaceName             string
	IP                    string
	Netmask               string
	MAC                   string
	portIndex             int
}

type RoutingTableEntry struct {
	network string
	netmask    string
	port    int
}

func Create(dp *hoverctl.Dataplane) *RouterModule {

	if dp == nil {
		log.Errorf("Daplane is not valid\n")
		return nil
	}

	r := new(RouterModule)
	r.Interfaces = make(map[string]*RouterModuleInterface)
	//r.PortsArray = make(map[int]bool)
	r.dataplane = dp
	r.deployed = false
	return r
}

func (r *RouterModule) Deploy() (error bool) {

	if r.deployed {
		return true
	}

	routerError, routerHover := hoverctl.ModulePOST(r.dataplane, "bpf",
									"Router", RouterCode)
	if routerError != nil {
		log.Errorf("Error in POST Router IOModule: %s\n", routerError)
		return false
	}

	log.Noticef("POST Router IOModule %s\n", routerHover.Id)
	r.ModuleId = routerHover.Id
	r.deployed = true

	return true
}

func (r *RouterModule) Destroy() (error bool) {

	if !r.deployed {
		return true
	}

	// TODO:
	// All interfaces must be detached before destroying the module.
	// Should it be done automatically here, or should be the application responsible for that?

	moduleDeleteError, _ := hoverctl.ModuleDELETE(r.dataplane, r.ModuleId)
	if moduleDeleteError != nil {
		log.Errorf("Error in destrying Router IOModule: %s\n", moduleDeleteError)
		return false
	}

	r.ModuleId = ""
	r.deployed = false

	return true
}

func (r *RouterModule) AttachPort(ifaceName string, ip string, netmask string, mac string) (error bool) {

	if !r.deployed {
		log.Errorf("Trying to attach port in undeployed router\n")
		return false
	}

	portIndex := -1

	for i := 0; i < 10; i++ {
		if !r.PortsArray[i] {
			portIndex = i
			break
		}
	}

	if portIndex == -1 {
		log.Errorf("There are not free ports in the router\n")
		return false
	}

	linkError, linkHover := hoverctl.LinkPOST(r.dataplane, "i:" + ifaceName, r.ModuleId)
	if linkError != nil {
		log.Errorf("Error in POSTing the Link: %s\n", linkError)
		return false
	}

	portIndexString := strconv.Itoa(portIndex)
	ipString := ipToHexadecimalString(ip)
	netmaskString := ipToHexadecimalString(netmask)
	macString := macToHexadecimalString(mac)

	toSend := ipString + " " + netmaskString + " " + macString

	hoverctl.TableEntryPOST(r.dataplane, r.ModuleId, "router_port",
		portIndexString, toSend)

	_, external_interfaces := hoverctl.ExternalInterfacesListGET(r.dataplane)

	r.PortsArray[portIndex] = true
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
		log.Warningf("IfaceIdRedirectHover == -1 something wrong happend...\n")
	}

	ip_ := net.ParseIP(ip)
	netmask_ := net.ParseIP(netmask)
	network_ := ip_.Mask(net.IPMask{netmask_[0], netmask_[1], netmask_[2], netmask_[3]})

	if network_.To4() == nil {
		log.Warningf("Something when wrong. (I know you are hating me for writing this non descrite error log)\n")
		return false
	}

	if !r.AddRoutingTableEntry(network_.String(), netmask, ifacenumber) {
		log.Warningf("Error adding static route for port '%s' in router '%s'\n",
			ifaceName, r.ModuleId)
		return false
	}

	iface := new(RouterModuleInterface)

	iface.IfaceFd, _ = strconv.Atoi(external_interfaces[ifaceName].Id)
	iface.IfaceIdRedirectHover = ifacenumber
	iface.LinkIdHover = linkHover.Id
	iface.IfaceName = ifaceName
	iface.IP = ip
	iface.Netmask = netmask
	iface.MAC = mac
	iface.portIndex = portIndex

	r.Interfaces[ifaceName] = iface

	return true
}

func (r *RouterModule) DetachPort(ifaceName string) (error bool) {

	if !r.deployed {
		log.Errorf("Trying to detach port in undeployed switch\n")
		return false
	}

	iface, ok := r.Interfaces[ifaceName]

	if !ok {
		log.Warningf("Iface '%s' is not present in router '%s'\n",
			ifaceName, r.ModuleId)
		return false
	}

	linkDeleteError, _ := hoverctl.LinkDELETE(r.dataplane, iface.LinkIdHover)

	if linkDeleteError != nil {
		//log.Debug("REMOVE Interface %s %s (1/1) LINK REMOVED\n", currentInterface.Name, currentInterface.IfaceIdExternalIds)
		log.Warningf("Problem removing iface '%s' from router '%s'\n",
			ifaceName, r.ModuleId)
		return false
	}

	// remove port from list of ports
	portIndexString := strconv.Itoa(iface.portIndex)
	hoverctl.TableEntryDELETE(r.dataplane, r.ModuleId, "router_port", portIndexString)

	delete(r.Interfaces, ifaceName)

	return true
}

// With the current implementation of the eBPF router this functions is a kind
// of complicated.  This is because it has to add the routes in a sorte way in the
// table, routes with the longest prefix should appear first.
// However current implementation, it is the very first one, only adds the routes
// one after the other, without performing this sorting
func (r *RouterModule) AddRoutingTableEntry(network string, netmask string, port int) (error bool) {

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

	stringIndex := strconv.Itoa(index)
	toSend := "{" + ipToHexadecimalString(network) + " " + ipToHexadecimalString(netmask) + " " + strconv.Itoa(port) + "}"

	hoverctl.TableEntryPUT(r.dataplane, r.ModuleId, "routing_table",
		stringIndex, toSend)

	r.RoutingTable[index].network = network
	r.RoutingTable[index].netmask = netmask
	r.RoutingTable[index].port = port

	return true
}

// TODO: Implement this function
//func (r *RouterModule) RemoveRoutingTableEntry(network string) {
//
//}

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
