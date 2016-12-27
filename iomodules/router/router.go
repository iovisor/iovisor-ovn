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

//Sanity Check Packet -> minimum length and correct checksum
//decrement TTL and recompute packet checksum (l3 recompute checksum)

//lookup in the longest prefix matching table:
//destination ip address of the packet.

//LONGEST PREFIX MATCHING trivialimplementation

var RouterCode = `
#include <linux/ip.h>
#include <linux/bpf.h>
#include <linux/kernel.h>

// #define BPF_TRACE
#undef BPF_TRACE

#define BPF_LOG
// #undef BPF_LOG

#define ROUTING_TABLE_DIM 10
#define ROUTER_PORT_N     10
#define ARP_TABLE_DIM     10

#define IP_TTL_OFFSET  8
#define IP_CSUM_OFFSET 10

#define ETH_DST_OFFSET  0
#define ETH_SRC_OFFSET  6
#define ETH_TYPE_OFFSET 12

/*Routing Table Entry*/
struct rt_entry{
  u32 network;  //network: e.g. 192.168.1.0
  u32 netmask;  //netmask: e.g. 255.255.255.0
  u32 port;     //port of the router
};

/*Router Port*/
struct r_port{
  u32 ip;       //ip addr : e.g. 192.168.1.254
  u32 netmask;  //netmask : e.g. 255.255.255.0
  u64 mac;      //mac addr: e.g. a1:b2:c3:ab:cd:ef
};

/*Arp Table Key*/
struct arp_table_key{
  u32 ip;       //ip addr : e.g. 192.168.1.2
  u32 port;     //port    : e.g. 1
};

/*
  The Routing table is implemented as an array of struct rt_entry (Routing Table Entry)
  the longest prefix matching algorithm (at least a simplified version)
  is implemented performing a bounded loop over the entries of the routing table.
  We assume that the control plane puts entry ordered from the longest netmask
  to the shortest one.
*/
BPF_TABLE("array", u32, struct rt_entry, routing_table, ROUTING_TABLE_DIM);

/*
  Router Port table provides a way to simulate the physical interface of the router
  The ip address is used to answer to the arp request (TO IMPLEMENT)
  The mac address is used as mac_scr for the outcoming packet on that interface,
  and as mac address contained in the arp reply
*/
BPF_TABLE("hash", u32, struct r_port, router_port, ROUTER_PORT_N);

/*
  We shold have an arp table for each port of the router?
  For now we assume to send packet exiting the router interfaces in broadcast
  (mac dst = ff:ff:ff:ff:ff:ff)

  How can we implement multiple arp tables?
  One possible implementation using one single map is the following
  key{ ip + port number } -> value {mac_address}
*/
BPF_TABLE("hash", u32, u64, arp_table, ARP_TABLE_DIM);

static int handle_rx(void *skb, struct metadata *md) {
  u8 *cursor = 0;
  struct ethernet_t *ethernet = cursor_advance(cursor, sizeof(*ethernet));

  #ifdef BPF_TRACE
    bpf_trace_printk("[router-%d]: in_ifc:%d\n", md->module_id, md->in_ifc);
    bpf_trace_printk("[router-%d]: eth_type:%x mac_scr:%lx mac_dst:%lx\n",
      md->module_id, ethernet->type, ethernet->src, ethernet->dst);
  #endif

  //TODO
  //sanity check of the packet.
  //if something wrong -> DROP the packet

  // is it an ipv4 packet?
  if (ethernet->type == 0x0800) {
    struct ip_t *ip = cursor_advance(cursor, sizeof(*ip));

    #ifdef BPF_TRACE
      bpf_trace_printk("[router-%d]: ttl:%u ip_scr:%x ip_dst:%x \n", md->module_id, ip->ttl, ip->src, ip->dst);
      // bpf_trace_printk("[router-%d]: (before) ttl: %d checksum: %x\n", ip->ttl, ip->hchecksum);
    #endif

    /*
      decrement TTL and recompute packet checksum (l3 recompute checksum).
      if ttl <= 1 DROP the packet.
      eventually send ICMP message for the packet dropped.
      (maybe to avoid for security reasons)
    */

    __u8 old_ttl = ip->ttl;
    __u8 new_ttl;

    if (old_ttl <= 1) {
      #ifdef BPF_TRACE
        bpf_trace_printk("[router-%d]: packet DROP (ttl <= 1)\n", md->module_id);
      #endif
      return RX_DROP;
    }

    new_ttl = old_ttl - 1;
    bpf_l3_csum_replace(skb, sizeof(*ethernet) + IP_CSUM_OFFSET , old_ttl, new_ttl, sizeof(__u16));
    bpf_skb_store_bytes(skb, sizeof(*ethernet) + IP_TTL_OFFSET , &new_ttl, sizeof(old_ttl), 0);

    #ifdef BPF_TRACE
      // bpf_trace_printk("[router-%d]: (after ) ttl: %d checksum: %x\n",ip->ttl,ip->hchecksum);
    #endif

    /*
      ROUTING ALGORITHM (simplified)

      for each item in the routing table (upbounded loop)
      apply the netmask on dst_ip_address
      (possible optimization, not recompute if at next iteration the netmask is the same)
      if masked address == network in the routing table
        1- change src mac to otuput port mac
        2- change dst mac to lookup arp table (or send to fffffffffffff)
        3- forward the packet to dst port
    */

    int i = 0;
    struct rt_entry *rt_entry_p = 0;

    u64 new_src_mac = 0;
    u64 new_dst_mac = 0;
    u32 out_port = 0;
    struct r_port *r_port_p = 0;

    #pragma unroll
    for (i = 0; i < ROUTING_TABLE_DIM; i++) {
      u32 t = i;
      rt_entry_p = routing_table.lookup(&t);
       if (rt_entry_p) {
        if ((ip->dst & rt_entry_p->netmask) == rt_entry_p->network) {
          goto FORWARD;
        }
      }
    }

  DROP:
    #ifdef BPF_LOG
      bpf_trace_printk("[router-%d]: in: %d out: -- DROP\n", md->module_id, md->in_ifc);
    #endif
    return RX_DROP;

  FORWARD:
    //Select out interface
    out_port = rt_entry_p->port;
    if (out_port <= 0)
      goto DROP;

    #ifdef BPF_LOG
      bpf_trace_printk("[router-%d]: routing table match (#%d) network: %x\n",
        md->module_id, i, rt_entry_p->network);
    #endif

    //change src mac
    r_port_p = router_port.lookup(&out_port);
    if (r_port_p) {
      new_src_mac = cpu_to_be64(r_port_p->mac<<16);
      bpf_skb_store_bytes(skb,ETH_SRC_OFFSET, &new_src_mac, 6, 0);
    }

    //change dst mac to ff:ff:ff:ff:ff:ff (TODO arp table)
    new_dst_mac = 0xffffffffffff;
    bpf_skb_store_bytes(skb, ETH_DST_OFFSET, &new_dst_mac, 6, 0);

    #ifdef BPF_TRACE
      bpf_trace_printk("[router-%d]: eth_type:%x mac_scr:%lx mac_dst:%lx\n",
        md->module_id, ethernet->type, ethernet->src, ethernet->dst);
      bpf_trace_printk("[router-%d]: out_ifc: %d\n", out_port);
    #endif

    #ifdef BPF_LOG
      bpf_trace_printk("[router-%d]: in: %d out: %d REDIRECT\n", md->module_id, md->in_ifc, out_port);
    #endif

    pkt_redirect(skb,md,out_port);
    return RX_REDIRECT;
  }
  else if(ethernet->type == 0x0806) { // is it ARP?
    struct arp_t *arp = cursor_advance(cursor, sizeof(*arp));
    if (arp->oper == 1) {	// arp request?
      //bpf_trace_printk("[arp]: packet is arp request\n");

      struct r_port *port = router_port.lookup(&md->in_ifc);
      if (!port)
        return RX_DROP;
      if (arp->tpa == port->ip) {
        //bpf_trace_printk("[arp]: Somebody is asking for my address\n");

        /* answer arp request */

        u16 two = cpu_to_be16(0x0002);
        u64 mymac = cpu_to_be64(port->mac<<16);
        u64 remotemac = arp->sha;
        remotemac = cpu_to_be64(remotemac<<16);
        u32 myip = cpu_to_be32(port->ip);
        u32 remoteip = arp->spa;
        remoteip = cpu_to_be32(remoteip);

        bpf_skb_store_bytes(skb, 0, &remotemac, 6, 0); // dst_mac
        bpf_skb_store_bytes(skb, 6, &mymac, 6, 0); // src_mac
        bpf_skb_store_bytes(skb, sizeof(*ethernet)+6, &two, 2, 0); // operation
        bpf_skb_store_bytes(skb, sizeof(*ethernet)+8, &mymac, 6, 0);// sha
        bpf_skb_store_bytes(skb, sizeof(*ethernet)+14, &myip, 4, 1);// spa
        bpf_skb_store_bytes(skb, sizeof(*ethernet)+18, &remotemac, 6, 0);// tha
        bpf_skb_store_bytes(skb, sizeof(*ethernet)+24, &remoteip, 4, 0);// tpa

        /* register the requesting mac and ips */
        arp_table.update(&remoteip, &remotemac);

        pkt_redirect(skb, md, md->in_ifc);

        return RX_REDIRECT;

      }
    }
    else if (arp->oper == 2) { //arp reply
      bpf_trace_printk("[router-%d]: packet is arp reply\n", md->module_id);

      struct r_port *port = router_port.lookup(&md->in_ifc);
      if (!port)
        return RX_DROP;
      if (arp->sha == port->mac && arp->spa == port->ip) {
        u64 mac_ = port->mac;
        u32 ip_ = port->ip;
        arp_table.update(&ip_, &mac_);
        return RX_DROP;
      }
    }
  }

  return RX_DROP;
}
`
