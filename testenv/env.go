// TestEnv launches a set of predefined configurations.
// It is useful to launch some fixed environments
// for example a switch witch ports connected to veth1_ veth2_ etc..

package testenv

import (
	"fmt"
	"time"

	"github.com/netgroup-polito/iovisor-ovn/bpf"
	"github.com/netgroup-polito/iovisor-ovn/hoverctl"

	l "github.com/op/go-logging"
)

var log = l.MustGetLogger("politoctrl")

//TestEnv Launches a defined configuration at daemon startup
func TestEnv(dataplane *hoverctl.Dataplane) {

	//testSwitch2(dataplane)

	//testSwitch2count(dataplane)

	//TestSwitch3(dataplane)

	TestSwitch5(dataplane)

}

//2 ports switch test implementation
//veth1_ <-> DummySwitch2 <-> veth2_
func TestLinkPostDelete(dataplane *hoverctl.Dataplane) {
	_, sw := hoverctl.ModulePOST(dataplane, "bpf", "DummySwitch2", bpf.Switch2Redirect)
	_, l1 := hoverctl.LinkPOST(dataplane, "i:veth1_", sw.Id)
	_, l2 := hoverctl.LinkPOST(dataplane, "i:veth2_", sw.Id)
	time.Sleep(time.Millisecond * 2000)
	hoverctl.LinkDELETE(dataplane, l1.Id)
	hoverctl.LinkDELETE(dataplane, l2.Id)
}

//2 ports switch test implementation
//veth1_ <-> DummySwitch2 <-> veth2_
func TestSwitch2(dataplane *hoverctl.Dataplane) {
	_, sw := hoverctl.ModulePOST(dataplane, "bpf", "DummySwitch2", bpf.DummySwitch2)
	/*	_, l1 := */ hoverctl.LinkPOST(dataplane, "i:veth1_", sw.Id)
	/*	_, l2 := */ hoverctl.LinkPOST(dataplane, "i:veth2_", sw.Id)
}

//2 ports switch test implementation
//veth1_ <-> DummySwitch2 <-> veth2_
func TestSwitch2ifc(dataplane *hoverctl.Dataplane, ifc1 string, ifc2 string) {
	_, sw := hoverctl.ModulePOST(dataplane, "bpf", "DummySwitch2", bpf.DummySwitch2)
	/*	_, l1 := */ hoverctl.LinkPOST(dataplane, ifc1, sw.Id)
	/*	_, l2 := */ hoverctl.LinkPOST(dataplane, ifc2, sw.Id)
}

//3 ports switch test implementation
//veth123_ <-> DummySwitch3
func TestSwitch3(dataplane *hoverctl.Dataplane) {
	_, sw := hoverctl.ModulePOST(dataplane, "bpf", "DummySwitch3", bpf.DummySwitch3)
	/*	_, l1 := */ hoverctl.LinkPOST(dataplane, "i:veth1_", sw.Id)
	/*	_, l2 := */ hoverctl.LinkPOST(dataplane, "i:veth2_", sw.Id)
	/*	_, l3 := */ hoverctl.LinkPOST(dataplane, "i:veth3_", sw.Id)
}

//8 ports switch implementation with 5 ports attached
//veth12345_ <-> Switch
func TestSwitch5(dataplane *hoverctl.Dataplane) {
	log.Noticef("TestSwitch5: create a Switch with 8 ports, and connect vethx_ (1,2,3,4,5) to the switch\n")

	_, sw := hoverctl.ModulePOST(dataplane, "bpf", "Switch", bpf.Switch)
	/*	_, l1 := */ hoverctl.LinkPOST(dataplane, "i:veth1_", sw.Id)
	/*	_, l2 := */ hoverctl.LinkPOST(dataplane, "i:veth2_", sw.Id)
	/*	_, l3 := */ hoverctl.LinkPOST(dataplane, "i:veth3_", sw.Id)
	/*  _, l4 := */ hoverctl.LinkPOST(dataplane, "i:veth4_", sw.Id)
	/*	_, l3 := */ hoverctl.LinkPOST(dataplane, "i:veth5_", sw.Id)

	_, external_interfaces := hoverctl.ExternalInterfacesListGET(dataplane)
	log.Debugf("%+v\n", external_interfaces)

	hoverctl.TableEntryPUT(dataplane, sw.Id, "ports", "0x1", external_interfaces["veth1_"].Id)
	hoverctl.TableEntryPUT(dataplane, sw.Id, "ports", "0x2", external_interfaces["veth2_"].Id)
	hoverctl.TableEntryPUT(dataplane, sw.Id, "ports", "0x3", external_interfaces["veth3_"].Id)
	hoverctl.TableEntryPUT(dataplane, sw.Id, "ports", "0x4", external_interfaces["veth4_"].Id)
	hoverctl.TableEntryPUT(dataplane, sw.Id, "ports", "0x5", external_interfaces["veth5_"].Id)

	hoverctl.TableEntryGET(dataplane, sw.Id, "ports", "0x1")
	hoverctl.TableEntryGET(dataplane, sw.Id, "ports", "0x2")
	hoverctl.TableEntryGET(dataplane, sw.Id, "ports", "0x3")
	hoverctl.TableEntryGET(dataplane, sw.Id, "ports", "0x4")
	hoverctl.TableEntryGET(dataplane, sw.Id, "ports", "0x5")

	hoverctl.ModuleListGET(dataplane)
}

//2 ports switch test implementation & pkt count
//veth1_ <-> DummySwitch2 <-> veth2_
//with counters
func TestSwitch2count(dataplane *hoverctl.Dataplane) {

	_, sw := hoverctl.ModulePOST(dataplane, "bpf", "DummySwitch2count", bpf.DummySwitch2count)
	/*	_, l1 := */ hoverctl.LinkPOST(dataplane, "i:veth1_", sw.Id)
	/*	_, l2 := */ hoverctl.LinkPOST(dataplane, "i:veth2_", sw.Id)

	//get Count before packet traffic
	hoverctl.TableEntryGET(dataplane, sw.Id, "count", "0x1")

	//fmt.Printf("key: %s value: %s\n", kv.Key, kv.Value)

	//wait...
	time.Sleep(time.Second * 8)

	//get Count after packet traffic
	hoverctl.TableEntryGET(dataplane, sw.Id, "count", "0x1")

	//force value to counter
	hoverctl.TableEntryPUT(dataplane, sw.Id, "count", "0x1", "0x9")

	//fmt.Printf("key: %s value: %s\n", kv.Key, kv.Value)

	//read value should be the previous.
	hoverctl.TableEntryGET(dataplane, sw.Id, "count", "0x1")

	//fmt.Printf("key: %s value: %s\n", kv.Key, kv.Value)

	//delete entry (TODO test why error)
	//probably because of array map type
	hoverctl.TableEntryDELETE(dataplane, sw.Id, "count", "0x1")

	//fmt.Printf("key: %s value: %s\n", kv.Key, kv.Value)

	//get after delete (probably returns error)
	hoverctl.TableEntryGET(dataplane, sw.Id, "count", "0x1")
}

func printLink(l hoverctl.Link) {
	fmt.Printf("id: %s\nfrom: %s\nto: %s\n", l.Id, l.From, l.To)
}

func printListModules(dataplane *hoverctl.Dataplane) {
	// errore, m := hoverctl.ModuleListGET(dataplane)
	// if errore!=nil {log.Errorf("e:%s\n", errore)}
}
