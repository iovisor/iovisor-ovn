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

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	_ "net"
	"net/http"
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

type Dataplane struct {
	client  *http.Client
	baseUrl string
	id      string
	port1   string
	port2   string
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
			"code": switchC,
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

	LinkModule(d, module.Id, d.port1)
	LinkModule(d, d.port2, module.Id)

	return nil
}

func LinkModule(d *Dataplane, link string, module string) error {
	//link port2-->module
	from := link
	to := module

	Info.Printf("Linking %s <--> %s\n", from, to)

	req2 := map[string]interface{}{
		"from": from,
		"to":   to,
	}
	var module2 moduleEntry
	err2 := d.postObject("/links/", req2, &module2)
	if err2 != nil {
		return err2
	}
	Info.Printf("Link id: %s", d.id)
	return nil
}

func (d *Dataplane) Id() string {
	return d.id
}

func (d *Dataplane) Close() error {
	return nil
}
