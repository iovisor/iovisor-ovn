// TestEnv launches a set of predefined configurations.
// It is useful to launch some fixed environments
// for example a switch witch ports connected to veth1_ veth2_ etc..

package tests

import (
	"bufio"
	"fmt"
	"os"
	"time"

	"github.com/netgroup-polito/iovisor-ovn/hoverctl"
	"github.com/netgroup-polito/iovisor-ovn/iomodules/developing"
	"github.com/netgroup-polito/iovisor-ovn/iomodules/l2switch"
	"github.com/netgroup-polito/iovisor-ovn/iomodules/router"

	l "github.com/op/go-logging"
)

var log = l.MustGetLogger("iovisor-ovn-daemon")

//tests Launches a defined configuration at daemon startup
func TestEnv(dataplane *hoverctl.Dataplane) {

	//testSwitch2(dataplane)

	//testSwitch2count(dataplane)

	//TestSwitch3(dataplane)

	//TestSwitch5(dataplane)

	//TestChainModule(dataplane)

	TestRouter(dataplane)
}

//veth1_<>m1<>m2<>m3<>veth2_
// func TestChainModule(dataplane *hoverctl.Dataplane) {
// 	_, m1 := hoverctl.ModulePOST(dataplane, "bpf", "Module1", l2switch.Module1)
// 	_, m2 := hoverctl.ModulePOST(dataplane, "bpf", "Module2", l2switch.Module2)
// 	_, m3 := hoverctl.ModulePOST(dataplane, "bpf", "Module3", l2switch.Module3)
//
// 	hoverctl.LinkPOST(dataplane, "i:veth1_", m1.Id)
// 	hoverctl.LinkPOST(dataplane, m1.Id, m2.Id)
// 	hoverctl.LinkPOST(dataplane, m2.Id, m3.Id)
// 	hoverctl.LinkPOST(dataplane, "i:veth2_", m3.Id)
//
// }

func TestRouter(dataplane *hoverctl.Dataplane) {
	_, router := hoverctl.ModulePOST(dataplane, "bpf", "RouterDev", router.Router)
	log.Debug(router.Id)

	_, l1 := hoverctl.LinkPOST(dataplane, "i:veth1_", router.Id)
	_, l2 := hoverctl.LinkPOST(dataplane, "i:veth2_", router.Id)
	_, l3 := hoverctl.LinkPOST(dataplane, "i:veth3_", router.Id)

	// hoverctl.LinkPOST(dataplane, "i:veth3_", sw.Id)

	hoverctl.TableEntryPUT(dataplane, router.Id, "routing_table", "0x0", "{ 0x01010100  0xffffff00  0x1}")
	hoverctl.TableEntryPUT(dataplane, router.Id, "routing_table", "0x1", "{ 0x02020200  0xffffff00  0x2}")
	hoverctl.TableEntryPUT(dataplane, router.Id, "routing_table", "0x2", "{ 0x03030300  0xffffff00  0x3}")

	hoverctl.TableEntryPOST(dataplane, router.Id, "router_port", "1", "0x010101fe 0xffffff00 0x987654321abc")
	hoverctl.TableEntryPOST(dataplane, router.Id, "router_port", "2", "0x020202fe 0xffffff00 0x123456789abc")
	hoverctl.TableEntryPOST(dataplane, router.Id, "router_port", "3", "0x030303fe 0xffffff00 0xaa34bb78ccdd")

	fmt.Print("Press 'Enter' to continue...")
	bufio.NewReader(os.Stdin).ReadBytes('\n')

	hoverctl.LinkDELETE(dataplane, l1.Id)
	hoverctl.LinkDELETE(dataplane, l2.Id)
	hoverctl.LinkDELETE(dataplane, l3.Id)

	hoverctl.ModuleDELETE(dataplane, router.Id)
}

func TestModule(dataplane *hoverctl.Dataplane) {
	_, sw := hoverctl.ModulePOST(dataplane, "bpf", "Switch6SecurityMacIp", l2switch.SwitchSecurityPolicy)
	hoverctl.LinkPOST(dataplane, "i:veth1_", sw.Id)
	hoverctl.LinkPOST(dataplane, "i:veth2_", sw.Id)

	_, external_interfaces := hoverctl.ExternalInterfacesListGET(dataplane)

	hoverctl.TableEntryPUT(dataplane, sw.Id, "ports", "0x1", external_interfaces["veth1_"].Id)
	hoverctl.TableEntryPUT(dataplane, sw.Id, "ports", "0x2", external_interfaces["veth2_"].Id)
}

//2 ports switch test implementation
//veth1_ <-> DummySwitch2 <-> veth2_
func TestLinkPostDelete(dataplane *hoverctl.Dataplane) {

	_, sw := hoverctl.ModulePOST(dataplane, "bpf", "Switch", developing.Switch)
	_, l1 := hoverctl.LinkPOST(dataplane, "i:veth1_", sw.Id)
	_, l2 := hoverctl.LinkPOST(dataplane, "i:veth2_", sw.Id)

	_, external_interfaces := hoverctl.ExternalInterfacesListGET(dataplane)

	hoverctl.TableEntryPUT(dataplane, sw.Id, "ports", "0x1", external_interfaces["veth1_"].Id)
	hoverctl.TableEntryPUT(dataplane, sw.Id, "ports", "0x2", external_interfaces["veth2_"].Id)

	time.Sleep(time.Millisecond * 10000)
	hoverctl.LinkDELETE(dataplane, l1.Id)
	hoverctl.LinkDELETE(dataplane, l2.Id)

	time.Sleep(time.Millisecond * 10000)

	_, l1 = hoverctl.LinkPOST(dataplane, "i:veth1_", sw.Id)
	_, l2 = hoverctl.LinkPOST(dataplane, "i:veth2_", sw.Id)

	_, external_interfaces = hoverctl.ExternalInterfacesListGET(dataplane)

	hoverctl.TableEntryPUT(dataplane, sw.Id, "ports", "0x1", external_interfaces["veth1_"].Id)
	hoverctl.TableEntryPUT(dataplane, sw.Id, "ports", "0x2", external_interfaces["veth2_"].Id)

	time.Sleep(time.Millisecond * 10000)
	hoverctl.LinkDELETE(dataplane, l1.Id)
	hoverctl.LinkDELETE(dataplane, l2.Id)

	// hoverctl.TableEntryPUT(dataplane, sw.Id, "ports", "0x3", external_interfaces["veth3_"].Id)
	// hoverctl.TableEntryPUT(dataplane, sw.Id, "ports", "0x4", external_interfaces["veth4_"].Id)
	// hoverctl.TableEntryPUT(dataplane, sw.Id, "ports", "0x5", external_interfaces["veth5_"].Id)
	//
	// hoverctl.TableEntryGET(dataplane, sw.Id, "ports", "0x1")
	// hoverctl.TableEntryGET(dataplane, sw.Id, "ports", "0x2")
	// hoverctl.TableEntryGET(dataplane, sw.Id, "ports", "0x3")
	// hoverctl.TableEntryGET(dataplane, sw.Id, "ports", "0x4")
	// hoverctl.TableEntryGET(dataplane, sw.Id, "ports", "0x5")
	//
	// hoverctl.ModuleListGET(dataplane)

	// _, sw := hoverctl.ModulePOST(dataplane, "bpf", "DummySwitch2", developing.Switch2Redirect)
	// _, l1 := hoverctl.LinkPOST(dataplane, "i:veth1_", sw.Id)
	// _, l2 := hoverctl.LinkPOST(dataplane, "i:veth2_", sw.Id)
	//
	// time.Sleep(time.Millisecond * 2000)
	// hoverctl.LinkDELETE(dataplane, l1.Id)
	// hoverctl.LinkDELETE(dataplane, l2.Id)
	//
	// time.Sleep(time.Millisecond * 2000)
	// /*_, l1 :=*/ hoverctl.LinkPOST(dataplane, "i:veth1_", sw.Id)
	// time.Sleep(time.Millisecond * 2000)
	// /*_, l2 :=*/ hoverctl.LinkPOST(dataplane, "i:veth2_", sw.Id)

}

//2 ports switch test implementation
//veth1_ <-> DummySwitch2 <-> veth2_
func TestSwitch2(dataplane *hoverctl.Dataplane) {
	_, sw := hoverctl.ModulePOST(dataplane, "bpf", "DummySwitch2", developing.DummySwitch2)
	/*	_, l1 := */ hoverctl.LinkPOST(dataplane, "i:veth1_", sw.Id)
	/*	_, l2 := */ hoverctl.LinkPOST(dataplane, "i:veth2_", sw.Id)
}

//2 ports switch test implementation
//veth1_ <-> DummySwitch2 <-> veth2_
func TestSwitch2ifc(dataplane *hoverctl.Dataplane, ifc1 string, ifc2 string) {
	_, sw := hoverctl.ModulePOST(dataplane, "bpf", "DummySwitch2", developing.DummySwitch2)
	/*	_, l1 := */ hoverctl.LinkPOST(dataplane, ifc1, sw.Id)
	/*	_, l2 := */ hoverctl.LinkPOST(dataplane, ifc2, sw.Id)
}

//3 ports switch test implementation
//veth123_ <-> DummySwitch3
// func TestSwitch3(dataplane *hoverctl.Dataplane) {
// 	_, sw := hoverctl.ModulePOST(dataplane, "bpf", "DummySwitch3", l2switch.DummySwitch3)
// 	/*	_, l1 := */ hoverctl.LinkPOST(dataplane, "i:veth1_", sw.Id)
// 	/*	_, l2 := */ hoverctl.LinkPOST(dataplane, "i:veth2_", sw.Id)
// 	/*	_, l3 := */ hoverctl.LinkPOST(dataplane, "i:veth3_", sw.Id)
// }

//8 ports switch implementation with 5 ports attached
//veth12345_ <-> Switch
func TestSwitch5(dataplane *hoverctl.Dataplane) {
	log.Noticef("TestSwitch5: create a Switch with 8 ports, and connect vethx_ (1,2,3,4,5) to the switch\n")

	_, sw := hoverctl.ModulePOST(dataplane, "bpf", "Switch", developing.Switch)
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
// func TestSwitch2count(dataplane *hoverctl.Dataplane) {
//
// 	_, sw := hoverctl.ModulePOST(dataplane, "bpf", "DummySwitch2count", l2switch.DummySwitch2count)
// 	/*	_, l1 := */ hoverctl.LinkPOST(dataplane, "i:veth1_", sw.Id)
// 	/*	_, l2 := */ hoverctl.LinkPOST(dataplane, "i:veth2_", sw.Id)
//
// 	//get Count before packet traffic
// 	hoverctl.TableEntryGET(dataplane, sw.Id, "count", "0x1")
//
// 	//fmt.Printf("key: %s value: %s\n", kv.Key, kv.Value)
//
// 	//wait...
// 	time.Sleep(time.Second * 8)
//
// 	//get Count after packet traffic
// 	hoverctl.TableEntryGET(dataplane, sw.Id, "count", "0x1")
//
// 	//force value to counter
// 	hoverctl.TableEntryPUT(dataplane, sw.Id, "count", "0x1", "0x9")
//
// 	//fmt.Printf("key: %s value: %s\n", kv.Key, kv.Value)
//
// 	//read value should be the previous.
// 	hoverctl.TableEntryGET(dataplane, sw.Id, "count", "0x1")
//
// 	//fmt.Printf("key: %s value: %s\n", kv.Key, kv.Value)
//
// 	//delete entry (TODO test why error)
// 	//probably because of array map type
// 	hoverctl.TableEntryDELETE(dataplane, sw.Id, "count", "0x1")
//
// 	//fmt.Printf("key: %s value: %s\n", kv.Key, kv.Value)
//
// 	//get after delete (probably returns error)
// 	hoverctl.TableEntryGET(dataplane, sw.Id, "count", "0x1")
// }

func printLink(l hoverctl.Link) {
	fmt.Printf("id: %s\nfrom: %s\nto: %s\n", l.Id, l.From, l.To)
}

func printListModules(dataplane *hoverctl.Dataplane) {
	// errore, m := hoverctl.ModuleListGET(dataplane)
	// if errore!=nil {log.Errorf("e:%s\n", errore)}
}
