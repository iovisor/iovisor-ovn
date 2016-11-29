package l2switch

import (
	"strconv"

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

func (sw *L2SwitchModule) Deploy() (error bool) {

	if sw.deployed {
		return true
	}

	switchError, switchHover := hoverctl.ModulePOST(sw.dataplane, "bpf",
									"Switch", SwitchSecurityPolicy)
	if switchError != nil {
		log.Errorf("Error in POST Switch IOModule: %s\n", switchError)
		return false
	}

	log.Noticef("POST Switch IOModule %s\n", switchHover.Id)
	sw.ModuleId = switchHover.Id
	sw.deployed = true

	return true
}

func (sw *L2SwitchModule) Destroy() (error bool) {

	if !sw.deployed {
		return true
	}

	// TODO:
	// All interfaces must be detached before destroying the module.
	// Should it be done automatically here, or should be the application responsible for that?

	moduleDeleteError, _ := hoverctl.ModuleDELETE(sw.dataplane, sw.ModuleId)
	if moduleDeleteError != nil {
		log.Errorf("Error in destrying Switch IOModule: %s\n", moduleDeleteError)
		return false
	}

	sw.ModuleId = ""
	sw.deployed = false

	return true
}

func (sw *L2SwitchModule) AttachPort(ifaceName string) (error bool) {

	if !sw.deployed {
		log.Errorf("Trying to attach port in undeployed switch\n")
		return false
	}

	linkError, linkHover := hoverctl.LinkPOST(sw.dataplane, "i:" + ifaceName, sw.ModuleId)
	if linkError != nil {
		log.Errorf("Error in POSTing the Link: %s\n", linkError)
		return false
	}

	portNumber := sw.FindFirstFreeLogicalPort()

	if portNumber == 0 {
		log.Warningf("Switch '%s': no free ports.\n", sw.ModuleId)
		return false
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
		return false
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

	return true
}

func (sw *L2SwitchModule) DetachPort(ifaceName string) (error bool) {

	if !sw.deployed {
		log.Errorf("Trying to detach port in undeployed switch\n")
		return false
	}

	iface, ok := sw.Interfaces[ifaceName]

	if !ok {
		log.Warningf("Iface '%s' is not present in switch '%s'\n",
			ifaceName, sw.ModuleId)
		return false
	}

	linkDeleteError, _ := hoverctl.LinkDELETE(sw.dataplane, iface.LinkIdHover)

	if linkDeleteError != nil {
		//log.Debug("REMOVE Interface %s %s (1/1) LINK REMOVED\n", currentInterface.Name, currentInterface.IfaceIdExternalIds)
		log.Warningf("Problem removing iface '%s' from switch '%s'\n",
			ifaceName, sw.ModuleId)
		return false
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

	return true
}

func (sw *L2SwitchModule) FindFirstFreeLogicalPort() int {
	for i := 1; i < config.SwitchPortsNumber + 1; i++ {
		if sw.PortsArray[i] == 0 {
			return i
		}
	}
	return 0
}

// TODO: port security policies
