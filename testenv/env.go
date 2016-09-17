package testenv

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/netgroup-polito/iovisor-ovn/bpf"
	"github.com/netgroup-polito/iovisor-ovn/hoverctl"
)

//TestEnv Launches a defined configuration at daemon startup
func TestEnv(dataplane *hoverctl.Dataplane) {

	//testSwitch2(dataplane)

	//testSwitch2count(dataplane)

	//testSwitch3(dataplane)

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

func printLink(l hoverctl.LinkEntry) {
	fmt.Printf("id: %s\nfrom: %s\nto: %s\n", l.Id, l.From, l.To)
}

func printListModules(dataplane *hoverctl.Dataplane) {
	errore, m := hoverctl.ModuleListGET(dataplane)
	fmt.Printf("e:%s\n", errore)

	o, _ := json.Marshal(m.ListModules)
	fmt.Println(string(o))
}
