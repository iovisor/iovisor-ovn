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
#include <bcc/proto.h>
#include <bcc/helpers.h>

#undef BPF_TRACE
#define BPF_LOG
#undef BPF_TRACE_ICMP_ECHO_REPLY

#undef CHECK_MAC_DST

#define ROUTING_TABLE_DIM  9
#define ROUTER_PORT_N     10
#define ARP_TABLE_DIM     10

#define IP_TTL_OFFSET  8
#define IP_CSUM_OFFSET 10
#define IP_SRC_OFF 26
#define IP_DST_OFF 30
#define IP_CKSUM_OFF 24
#define ICMP_CSUM_OFFSET (sizeof(struct ethernet_t) + sizeof(struct ip_t) + offsetof(struct icmp_hdr, checksum))

#define IP_ICMP 0x01

#define ICMP_ECHO_REQUEST 0x8
#define ICMP_ECHO_REPLY 0x0

#define ETH_DST_OFFSET  0
#define ETH_SRC_OFFSET  6
#define ETH_TYPE_OFFSET 12

#define ETH_TYPE_IP 0x0800
#define ETH_TYPE_ARP 0x0806

#define MAC_BROADCAST 0xffffffffffff
#define MAC_MULTICAST_MASK 0x010000000000

#define SLOWPATH_ARP_REPLY 1
#define SLOWPATH_ARP_LOOKUP_MISS 2

/*Only for ICMP echo req, reply*/
struct icmp_hdr {
  unsigned char type;
  unsigned char code;
  unsigned short checksum;
  unsigned short id;
  unsigned short seq;
} __attribute__((packed));

/* Routing Table Entry */
struct rt_entry {
  u32 network;  //network: e.g. 192.168.1.0
  u32 netmask;  //netmask: e.g. 255.255.255.0
  u32 port;     //port of the router
  u32 nexthop;  //next hop: e.g. 192.168.1.254 (0 if local)
};

/* Router Port */
struct r_port {
  u32 ip;       //ip addr : e.g. 192.168.1.254
  u32 netmask;  //netmask : e.g. 255.255.255.0
  u64 mac;      //mac addr: e.g. a1:b2:c3:ab:cd:ef
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
  Arp Table implements a mapping between ip and mac addresses.
*/
BPF_TABLE("hash", u32, u64, arp_table, ARP_TABLE_DIM);

/*
  Check multicast bit of a mac address.
  If the address is broadcast is also multicast, so test multicast condition
  is enough.
*/
static inline bool is_multicast_or_broadcast(u64* mac) {
  u64 mask = 0;
  mask = *mac & MAC_MULTICAST_MASK;
  if (mask == 0)
    return false;
  else
    return true;
}

static int handle_rx(void *skb, struct metadata *md) {
  u8 *cursor = 0;
  struct ethernet_t *ethernet = cursor_advance(cursor, sizeof(*ethernet));

  #ifdef BPF_TRACE
    bpf_trace_printk("[router-%d]: in_ifc:%d\n", md->module_id, md->in_ifc);
    bpf_trace_printk("[router-%d]: eth_type:%x mac_scr:%lx mac_dst:%lx\n",
      md->module_id, ethernet->type, ethernet->src, ethernet->dst);
  #endif

  /*
    Check if the mac destination of the packet is multicast, broadcast, or the
    unicast address of the router port.
    If not, drop the packet.
    Multicast addresses are managed as broadcast
  */
  #ifdef CHECK_MAC_DST
  u64 ethdst = ethernet->dst;
  if (!is_multicast_or_broadcast(&ethdst)){
    struct r_port *r_port_p = 0;
    r_port_p = router_port.lookup(&md->in_ifc);
    if (r_port_p) {
      if (r_port_p->mac != ethernet->dst) {
        #ifdef BPF_LOG
          bpf_trace_printk("[router-%d]: mac destination %lx MISMATCH %lx -> DROP packet.\n",
            md->module_id, ethernet->dst, r_port_p->mac);
        #endif
        return RX_DROP;
      }
    }
  }
  #endif

  switch (ethernet->type) {
    case ETH_TYPE_IP: goto IP;   //ipv4 packet
    case ETH_TYPE_ARP: goto ARP; //arp packet
  }

  IP: ; //ipv4 packet
    struct ip_t *ip = cursor_advance(cursor, sizeof(*ip));

    #ifdef BPF_TRACE
      bpf_trace_printk("[router-%d]: ttl:%u ip_scr:%x ip_dst:%x \n", md->module_id, ip->ttl, ip->src, ip->dst);
      // bpf_trace_printk("[router-%d]: (before) ttl: %d checksum: %x\n", ip->ttl, ip->hchecksum);
    #endif

    /* ICMP Echo Responder for router ports */
    if ( ip->nextp == IP_ICMP) {
      struct __sk_buff * skb2 = (struct __sk_buff *)skb;
      void *data = (void *)(long)skb2->data;
      void *data_end = (void *)(long)skb2->data_end;
      struct icmp_hdr *icmp = data + sizeof(struct ethernet_t) + sizeof(struct ip_t);

      if (data + sizeof(struct ethernet_t) + sizeof(struct ip_t) + sizeof(*icmp) > data_end)
        return RX_DROP;

      /*Only manage ICMP Request*/
      if ( icmp->type == ICMP_ECHO_REQUEST ){
        struct r_port *r_port_p = 0;
        r_port_p = router_port.lookup(&md->in_ifc);
        if (r_port_p) {
          if (r_port_p->ip == ip->dst) {
            //Reply to ICMP Echo request

            unsigned short type = ICMP_ECHO_REPLY;
            bpf_l4_csum_replace(skb,36, icmp->type, type,sizeof(type));
            bpf_skb_store_bytes(skb, 34, &type, sizeof(type),0);

            unsigned int old_src = bpf_ntohl(ip->src);
            unsigned int old_dst = bpf_ntohl(ip->dst);
            bpf_l3_csum_replace(skb, IP_CKSUM_OFF, old_src, old_dst, sizeof(old_dst));

             bpf_skb_store_bytes(skb, IP_SRC_OFF, &old_dst, sizeof(old_dst), 0);
             bpf_l3_csum_replace(skb, IP_CKSUM_OFF, old_dst, old_src, sizeof(old_src));
             bpf_skb_store_bytes(skb, IP_DST_OFF, &old_src, sizeof(old_src), 0);

             unsigned long long old_src_mac = ethernet->src;
             unsigned long long old_dst_mac = ethernet->dst;
             ethernet->src = old_dst_mac;
             ethernet->dst = old_src_mac;

           #ifdef BPF_TRACE_ICMP_ECHO_REPLY
              bpf_trace_printk("[router-%d]: ICMP ECHO Request from 0x%x port %d . Generating Reply ...\n", md->module_id, ip->src, md->in_ifc);
           #endif
             pkt_redirect(skb, md, md->in_ifc);
             return RX_REDIRECT;
          }
        }
      }
    }

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
      // bpf_trace_printk("[router-%d]: in: %d out: -- DROP\n", md->module_id, md->in_ifc);
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
      ethernet->src = r_port_p->mac;
    }

    u32 dst_ip = 0;
    if (rt_entry_p->nexthop == 0) {
      //Next Hop is local, directly lookup in arp table for the destination ip.
      dst_ip = ip->dst;
    } else {
      //Next Hop not local, lookup in arp table for the next hop ip address.
      dst_ip = rt_entry_p->nexthop;
    }

    u64 new_dst_mac = 0xffffffffffff;
    u64 *mac_entry = arp_table.lookup(&dst_ip);
    if (mac_entry) {
      new_dst_mac = *mac_entry;
    }else{
      #ifdef BPF_LOG
        // bpf_trace_printk("[router-%d]: arp lookup failed. Send to controller",md->module_id);
      #endif

      //Set metadata and send packet to slowpath
      u32 mdata[3];
      mdata[0] = dst_ip;
      mdata[1] = out_port;
      r_port_p = router_port.lookup(&out_port);
      u32 ip = 0;
      if (r_port_p) {
        ip = r_port_p->ip;
      }
      mdata[2] = ip;
      pkt_set_metadata(skb, mdata);
      pkt_controller(skb, md, SLOWPATH_ARP_LOOKUP_MISS);
      return RX_CONTROLLER;
    }

    ethernet->dst = new_dst_mac;

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

  ARP: ; //arp packet
    struct arp_t *arp = cursor_advance(cursor, sizeof(*arp));
    if (arp->oper == 1) {	// arp request?
    #ifdef BPF_LOG
      bpf_trace_printk("[router-%d]: packet is arp request\n",md->module_id);
    #endif
      struct r_port *port = router_port.lookup(&md->in_ifc);
      if (!port)
        return RX_DROP;
      if (arp->tpa == port->ip) {
        //bpf_trace_printk("[arp]: Somebody is asking for my address\n");

        /* due to a bcc issue: https://github.com/iovisor/bcc/issues/537 it
         * is necessary to copy the data field into a temporal variable
         */
        u64 mymac = port->mac;
        u64 remotemac = arp->sha;
        u32 myip = port->ip;
        u32 remoteip = arp->spa;

        ethernet->dst = remotemac;
        ethernet->src = mymac;

        /* please note that the mac has to be copied before that the ips.  This
         * is because the temporal variable used to save the mac has 8 byes, 2
         * more than the mac itself.  Then when copying the mac into the packet
         * the two first bytes of the ip are also modified.
         */
        arp->oper = 2;
        arp->tha = remotemac;
        arp->sha = mymac;
        arp->tpa = remoteip;
        arp->spa = myip;

        /* register the requesting mac and ips */
        arp_table.update(&remoteip, &remotemac);

        pkt_redirect(skb, md, md->in_ifc);
        return RX_REDIRECT;

        //TODO make sense to send the arp packet to the slowpath
        //in order to notify some arp entries updated?
      }
    }
    if (arp->oper == 2) { //arp reply
    #ifdef BPF_LOG
      bpf_trace_printk("[router-%d]: packet is arp reply\n",md->module_id);
    #endif
      struct r_port *port = router_port.lookup(&md->in_ifc);
      if (!port){
        return RX_DROP;
      }else{
        u64 mac_ = arp->sha;
        u32 ip_  = arp->spa;
        arp_table.update(&ip_, &mac_);

        //notify the slowpath. New arp reply received.
        pkt_controller(skb, md, SLOWPATH_ARP_REPLY);
        return RX_CONTROLLER;
      }
    }
  return RX_DROP;
}
`
