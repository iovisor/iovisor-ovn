package testenv

import (
	"github.com/mbertrone/politoctrl/bpf"
	"github.com/mbertrone/politoctrl/helper"
)

func TestEnv(dataplane *helper.Dataplane) {

	_, sw := helper.ModulePOST(dataplane, "bpf", "DummySwitch", bpf.DummySwitch2count)
	/*	_, l1 := */ helper.LinkPOST(dataplane, "i:veth1_", sw.Id)
	/*	_, l2 := */ helper.LinkPOST(dataplane, "i:veth2_", sw.Id)
	/*	_, l3 := */ helper.LinkPOST(dataplane, "i:veth3_", sw.Id)

	//	time.Sleep(time.Second * 8)

	helper.TableEntryGET(dataplane, sw.Id, "count", "0x1")

	//fmt.Printf("key: %s value: %s\n", kv.Key, kv.Value)

	helper.TableEntryPUT(dataplane, sw.Id, "count", "0x1", "0x9")

	//	fmt.Printf("key: %s value: %s\n", kv.Key, kv.Value)

	helper.TableEntryGET(dataplane, sw.Id, "count", "0x1")

	//fmt.Printf("key: %s value: %s\n", kv.Key, kv.Value)

	helper.TableEntryDELETE(dataplane, sw.Id, "count", "0x1")

	//fmt.Printf("key: %s value: %s\n", kv.Key, kv.Value)
	helper.TableEntryGET(dataplane, sw.Id, "count", "0x1")

	//fmt.Printf("key: %s value: %s\n", kv.Key, kv.Value)

	/*
		fmt.Printf("id: %s\nfrom: %s\nto: %s\n", l.Id, l.From, l.To)
		errore, m := helper.ModuleListGET(dataplane)
		fmt.Printf("e:%s\n", errore)

		o, _ := json.Marshal(m.ListModules)
		fmt.Println(string(o))
	*/

	/*
		err, l := helper.LinkGET(dataplane, l1.Id)
		if err != nil {
			fmt.Printf("error: %s\n", err)
		}
		out, _ := json.Marshal(l)
		fmt.Println(string(out))
	*/

	/*
		helper.LinkDELETE(dataplane, l1.Id)
		time.Sleep(time.Second * 2)
		helper.LinkDELETE(dataplane, l2.Id)
		time.Sleep(time.Second * 2)
		helper.LinkDELETE(dataplane, l3.Id)
		time.Sleep(time.Second * 2)

		helper.ModuleDELETE(dataplane, sw.Id)
	*/

}
