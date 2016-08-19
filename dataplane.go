// Copyright 2016 PLUMgrid
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

package politoctrl

/*
import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	_ "net"
	"net/http"
	//	"github.com/mbertrone/politoctrl/helper"
)

var switchC = `
struct mac_key {
	u64 mac;
};

struct host_info {
	u32 ifindex;
	u64 rx_pkts;
	u64 tx_pkts;
};

BPF_TABLE("hash", struct mac_key, struct host_info, mac2host, 10240);

static int handle_rx(void *skb, struct metadata *md) {
	//LEARNING PHASE: mapping src_mac with src_interface

	//Lets assume that the packet is at 0th location of the memory.
	u8 *cursor = 0;

	struct mac_key src_key = {};
	struct host_info src_info = {};

	//Extract ethernet header from the memory and point cursor to payload of ethernet header.
	struct ethernet_t *ethernet = cursor_advance(cursor, sizeof(*ethernet));

	//set src_mac as key
	src_key.mac = ethernet->src;

	//set in_ifc, and 0 counters as leaf
	src_info.ifindex = md->in_ifc;
	src_info.rx_pkts = 0;
	src_info.tx_pkts = 0;

	//lookup in mac2host; if no key present -> initialize with src_info
	struct host_info *src_host = mac2host.lookup_or_init(&src_key, &src_info);

	//if the same mac has changed interface, update it
	if(src_host->ifindex!=md->in_ifc)
		src_host->ifindex = md->in_ifc;

	//update rx_pkts counter
	lock_xadd(&src_host->rx_pkts, 1);



	//FORWARDING PHASE: select interface(s) to send the packet

	//set dst_mac as key
	struct mac_key dst_key = {ethernet->dst};

	//lookup in forwarding table mac2host
	struct host_info *dst_host = mac2host.lookup(&dst_key);

	if (dst_host) {
		//hit in forwarding table

		//increment tx_pkts counter
		lock_xadd(&dst_host->tx_pkts, 1);

		//redirect packet to correspondent interface
		pkt_redirect(skb, md, dst_host->ifindex);
		return RX_REDIRECT;

	} else {
		//miss in forwarding table

		//increment tx_pkts counter
		//lock_xadd(&dst_host->tx_pkts, 1);

		//TODO implement real broadcast!
		if(md->in_ifc==1){
			pkt_redirect(skb,md,2);
			return RX_REDIRECT;
		}
		if(md->in_ifc==2){
			pkt_redirect(skb,md,1);
			return RX_REDIRECT;
		}
		return RX_DROP;
	}
}
`

var switchC2 = `
struct mac_key {
	u64 mac;
};

struct host_info {
	u32 ifindex;
	u64 rx_pkts;
	u64 tx_pkts;
};

BPF_TABLE("hash", struct mac_key, struct host_info, mac2host, 10240);

static int handle_rx(void *skb, struct metadata *md) {
	//debug init for broadcast_count
	u32 key = 0, value = 0, *cnt;
	//LEARNING PHASE: mapping src_mac with src_interface

	//Lets assume that the packet is at 0th location of the memory.
	u8 *cursor = 0;

	struct mac_key src_key = {};
	struct host_info src_info = {};

	//Extract ethernet header from the memory and point cursor to payload of ethernet header.
	struct ethernet_t *ethernet = cursor_advance(cursor, sizeof(*ethernet));

	//set src_mac as key
	src_key.mac = ethernet->src;

	//set in_ifc, and 0 counters as leaf
	src_info.ifindex = md->in_ifc;
		src_info.rx_pkts = 0;
		src_info.tx_pkts = 0;

	//lookup in mac2host; if no key present -> initialize with src_info
	struct host_info *src_host = mac2host.lookup_or_init(&src_key, &src_info);

	//if the same mac has changed interface, update it
	if(src_host->ifindex!=md->in_ifc)
		src_host->ifindex = md->in_ifc;

	//update rx_pkts counter
	lock_xadd(&src_host->rx_pkts, 1);


	//FORWARDING PHASE: select interface(s) to send the packet
	//set dst_mac as key
	struct mac_key dst_key = {ethernet->dst};

	//lookup in forwarding table mac2host
	struct host_info *dst_host = mac2host.lookup(&dst_key);

	if (dst_host) {
		//increment tx_pkts counter
		lock_xadd(&dst_host->tx_pkts, 1);

		//redirect packet to correspondent interface
		pkt_redirect(skb, md, dst_host->ifindex);

		return RX_REDIRECT;

	} else {
		//miss in forwarding table
		//broadcast test implementation
			if(md->in_ifc==1){
				bpf_clone_redirect(skb, 2, 0);
				bpf_clone_redirect(skb, 3, 0);
				return RX_OK;
			}
			if(md->in_ifc==2){
				bpf_clone_redirect(skb, 1, 0);
				bpf_clone_redirect(skb, 3, 0);
				return RX_OK;
			}
			if(md->in_ifc==3){
				bpf_clone_redirect(skb, 1, 0);
				bpf_clone_redirect(skb, 2, 0);
				return RX_OK;
			}
		return RX_DROP;
	}
}
`
*/

/*
var switchC3 = `
//dummy indicates dummy broadcast implementation
//#define DUMMY

struct mac_key {
	u64 mac;
};

struct host_info {
	u32 ifindex;
	u64 rx_pkts;
	u64 tx_pkts;
};

BPF_TABLE("hash", struct mac_key, struct host_info, mac2host, 10240);

//Debug array to test how many time a certain portion of code is executed!
BPF_TABLE("array", u32, u32, broadcast_count, 20);

static int handle_rx(void *skb, struct metadata *md) {
	//debug init for broadcast_count
	u32 key = 0, value = 0, *cnt;
	//LEARNING PHASE: mapping src_mac with src_interface

	//0-Handle Rx & learning phase (should be processed for all packets)
	key = 0;cnt = broadcast_count.lookup_or_init(&key,&value);(*cnt)++;

	//Lets assume that the packet is at 0th location of the memory.
	u8 *cursor = 0;

	struct mac_key src_key = {};
	struct host_info src_info = {};

	//Extract ethernet header from the memory and point cursor to payload of ethernet header.
	struct ethernet_t *ethernet = cursor_advance(cursor, sizeof(*ethernet));

	//set src_mac as key
	src_key.mac = ethernet->src;

	//set in_ifc, and 0 counters as leaf
	src_info.ifindex = md->in_ifc;
		src_info.rx_pkts = 0;
		src_info.tx_pkts = 0;

	//lookup in mac2host; if no key present -> initialize with src_info
	struct host_info *src_host = mac2host.lookup_or_init(&src_key, &src_info);

	//if the same mac has changed interface, update it
	if(src_host->ifindex!=md->in_ifc)
		src_host->ifindex = md->in_ifc;

	//update rx_pkts counter
	lock_xadd(&src_host->rx_pkts, 1);



	//FORWARDING PHASE: select interface(s) to send the packet

	//1-fowarding phase
	key = 1;cnt = broadcast_count.lookup_or_init(&key,&value);(*cnt)++;

	//set dst_mac as key
	struct mac_key dst_key = {ethernet->dst};

	//lookup in forwarding table mac2host
	struct host_info *dst_host = mac2host.lookup(&dst_key);

	if (dst_host) {
		//2-hit in forwarding table
		key = 2;cnt = broadcast_count.lookup_or_init(&key,&value);(*cnt)++;

		//increment tx_pkts counter
		lock_xadd(&dst_host->tx_pkts, 1);

		//redirect packet to correspondent interface
		pkt_redirect(skb, md, dst_host->ifindex);

		return RX_REDIRECT;

	} else {

		//miss in forwarding table
		//3-miss in forwarding table
		key = 3;cnt = broadcast_count.lookup_or_init(&key,&value);(*cnt)++;

#ifndef DUMMY
			//real implementation of broadcast seems not to work!
		//	key = 9;cnt = broadcast_count.lookup_or_init(&key,&value);(*cnt)++;
/*
			if(md->in_ifc==1)
				bpf_clone_redirect(skb, 2, 0);
			if(md->in_ifc==2)
				bpf_clone_redirect(skb, 1, 0);
			return TC_ACT_SHOT;

			pkt_mirror(skb,md,2);
			key = 7;cnt = broadcast_count.lookup_or_init(&key,&value);(*cnt)++;
			pkt_mirror(skb,md,1);
			key = 8;cnt = broadcast_count.lookup_or_init(&key,&value);(*cnt)++;
			//return RX_OK;
*/
/*
	if(md->in_ifc==1){
		key = 4;cnt = broadcast_count.lookup_or_init(&key,&value);(*cnt)++;
		bpf_clone_redirect(skb, 2, 0);
		bpf_clone_redirect(skb, 3, 0);

	}
	if(md->in_ifc==2){
		key = 5;cnt = broadcast_count.lookup_or_init(&key,&value);(*cnt)++;
		bpf_clone_redirect(skb, 1, 0);
		bpf_clone_redirect(skb, 3, 0);
	}
	if(md->in_ifc==3){
		key = 6;cnt = broadcast_count.lookup_or_init(&key,&value);(*cnt)++;
		bpf_clone_redirect(skb, 1, 0);
		bpf_clone_redirect(skb, 2, 0);
	}
*/

//1-->3
//2-->3
//3-->2

//reality
//1-->3
//2-->3
//3-->1
//3-->2

/*
if(md->in_ifc==1){
	key = 4;cnt = broadcast_count.lookup_or_init(&key,&value);(*cnt)++;
	//bpf_redirect(2, 0);
	pkt_mirror(skb, md, 2);
	key = 7;cnt = broadcast_count.lookup_or_init(&key,&value);(*cnt)++;
	pkt_mirror(skb, md, 3);
	key = 8;cnt = broadcast_count.lookup_or_init(&key,&value);(*cnt)++;
	return RX_OK;
}
if(md->in_ifc==2){
	key = 5;cnt = broadcast_count.lookup_or_init(&key,&value);(*cnt)++;
	//bpf_clone_redirect(skb, 1, 0);
	//bpf_clone_redirect(skb, 3, 0);
	pkt_mirror(skb,md,1);
	pkt_mirror(skb,md,3);
	return RX_OK;
}
if(md->in_ifc==3){
	key = 6;cnt = broadcast_count.lookup_or_init(&key,&value);(*cnt)++;
	pkt_mirror(skb,md,1);
	pkt_mirror(skb,md,2);
	return RX_OK;
}

return RX_OK;

#endif

#ifdef DUMMY
		//Dummy broadcast implementation
		if(md->in_ifc==1){
			pkt_redirect(skb,md,2);
			key = 4;cnt = broadcast_count.lookup_or_init(&key,&value);(*cnt)++;
			return RX_REDIRECT;
		}
		if(md->in_ifc==2){
			pkt_redirect(skb,md,1);
			key = 5;cnt = broadcast_count.lookup_or_init(&key,&value);(*cnt)++;
			return RX_REDIRECT;
		}
#endif

		//key = 6;cnt = broadcast_count.lookup_or_init(&key,&value);(*cnt)++;
		//return RX_DROP;
	}
}
`

type Dataplane struct {
	client  *http.Client
	baseUrl string
	id      string
}

func NewDataplane() *Dataplane {
	client := &http.Client{}
	d := &Dataplane{
		client: client,
	}

	return d
}

func (d *Dataplane) postObject(url string, requestObj interface{}, responseObj interface{}) (err error) {
	b, err := json.Marshal(requestObj)
	if err != nil {
		return
	}
	resp, err := d.client.Post(d.baseUrl+url, "application/json", bytes.NewReader(b))
	if err != nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		var body []byte
		if body, err = ioutil.ReadAll(resp.Body); err != nil {
			Error.Print(string(body))
		}
		return fmt.Errorf("module server returned %s", resp.Status)
	}
	if responseObj != nil {
		err = json.NewDecoder(resp.Body).Decode(responseObj)
	}
	return
}

type moduleEntry struct {
	Id          string                 `json:"id"`
	ModuleType  string                 `json:"module_type"`
	DisplayName string                 `json:"display_name"`
	Perm        string                 `json:"permissions"`
	Config      map[string]interface{} `json:"config"`
}

func (d *Dataplane) Init(baseUrl string) error {
	d.baseUrl = baseUrl

	req := map[string]interface{}{
		"module_type":  "bpf",
		"display_name": "dummyBridge",
		"config": map[string]interface{}{
			"code": switchC2,
		},
	}
	var module moduleEntry
	err := d.postObject("/modules/", req, &module)
	if err != nil {
		return err
	}
	d.id = module.Id
	Info.Printf("dummyBridge module id: %s\n", d.id)
	Info.Printf("base url: %s\n", d.baseUrl)

	//LinkModule(d, module.Id, "i:veth1_")
	//LinkModule(d, "i:veth2_", module.Id)

	//	helper.LinkModule(d, module.Id, d.port1)
	//	helper.LinkModule(d, module.Id, d.port2)
	//added
	//	helper.LinkModule(d, module.Id, "i:veth3_")

	return nil
}

func (d *Dataplane) Id() string {
	return d.id
}

func (d *Dataplane) Close() error {
	return nil
}
*/
