// Copyright 2017 Politecnico di Torino
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
package onetoonenat

var NatCode = `
#include <uapi/linux/bpf.h>
#include <uapi/linux/if_ether.h>
#include <uapi/linux/if_packet.h>
#include <uapi/linux/ip.h>
#include <uapi/linux/in.h>
#include <uapi/linux/filter.h>
#include <uapi/linux/pkt_cls.h>

#include <bcc/proto.h>

#define BPF_TRACE_INGRESS
#undef BPF_TRACE_EGRESS_UDP
#undef BPF_TRACE_REVERSE_UDP
#undef BPF_TRACE_EGRESS_TCP
#undef BPF_TRACE_REVERSE_TCP
#undef BPF_TRACE_DROP
#undef BPF_TRACE_LOOKUP_FAILED

#define EGRESS_NAT_TABLE_DIM  1024
#define REVERSE_NAT_TABLE_DIM 1024

#define IN_IFC  1
#define OUT_IFC 2

#define ETH_TYPE_IP  0x800
#define ETH_TYPE_ARP 0x806

#define IP_TCP  0x06
#define IP_UDP  0x11
#define IP_ICMP 0x01

#define UDP_CSUM_OFFSET (sizeof(struct ethernet_t) + sizeof(struct ip_t) + offsetof(struct udp_t, crc))
#define TCP_CSUM_OFFSET (sizeof(struct ethernet_t) + sizeof(struct ip_t) + offsetof(struct tcp_t, cksum))
#define IP_CSUM_OFFSET (sizeof(struct ethernet_t) + offsetof(struct ip_t, hchecksum))

#define IS_PSEUDO 0x10

#define FORWARD_NO_NAT_TRAFFIC //If defined, forward also the no natted traffic.

/*Egress Nat key*/
struct egress_nat_key{
  u32 ip_src;
};

/*Egress Nat Value*/
struct egress_nat_value{
  u32 ip_src_new;
};

/*Reverse Nat key*/
struct reverse_nat_key{
  u32 ip_dst;
};

/*Reverse Nat Value*/
struct reverse_nat_value{
  u32 ip_dst_new;
};

/*
  Egress Nat Table.
*/
BPF_TABLE("hash", struct egress_nat_key, struct egress_nat_value, egress_nat_table, EGRESS_NAT_TABLE_DIM);

/*
  Reverse Nat Table.
*/
BPF_TABLE("hash", struct reverse_nat_key, struct reverse_nat_value, reverse_nat_table, REVERSE_NAT_TABLE_DIM);

/*given ip_src returns the correspondent association in
  the nat table.
*/
static inline struct egress_nat_value * get_egress_value(u32 ip_src){
  //Lookup in the egress_nat_table
  struct egress_nat_key egress_key = {};
  egress_key.ip_src = ip_src;
  struct egress_nat_value *egress_value_p = 0;
  egress_value_p = egress_nat_table.lookup(&egress_key);
  return egress_value_p;
}

static int handle_rx(void *skb, struct metadata *md) {
  u8 *cursor = 0;
  struct ethernet_t *ethernet = cursor_advance(cursor, sizeof(*ethernet));

#ifdef BPF_TRACE_INGRESS
  bpf_trace_printk("[nat-0]: eth_type:%x mac_src:%lx mac_dst:%lx\n",
  ethernet->type, ethernet->src, ethernet->dst);
#endif

  switch (ethernet->type){
    case ETH_TYPE_IP: goto IP;
    case ETH_TYPE_ARP: goto arp;
    default: goto EOP;
  }

  IP: ;
  struct ip_t *ip = cursor_advance(cursor, sizeof(*ip));

  switch (ip->nextp){
    case IP_UDP: goto udp;
    case IP_TCP: goto tcp;
    default: goto EOP;
  }

  /*
    IN_IFC (1) --> NAT --> (2) OUT_IFC
    Only ARP requests from OUT_IFC are considered.
    IN_IFC Should be attached to a router iface,
    for this reason packet should be forwarded to BCAST_MAC
  */
  arp: {
    if (md->in_ifc == OUT_IFC){
      pkt_redirect(skb, md, IN_IFC);
          return RX_REDIRECT;
      } else if (md->in_ifc == IN_IFC) {
      pkt_redirect(skb, md, OUT_IFC);
          return RX_REDIRECT;
    }

    return RX_DROP;
  }

  udp: {
      struct udp_t *udp = cursor_advance(cursor, sizeof(*udp));
      /*BEGIN UDP*/
      switch (md->in_ifc){
        case IN_IFC: goto EGRESS_UDP;
        case OUT_IFC: goto REVERSE_UDP;
      }
      goto DROP;

      EGRESS_UDP: ;
      //Packet exiting the nat, apply nat translation
    #ifdef BPF_TRACE_EGRESS_UDP
      bpf_trace_printk("[nat-0]: EGRESS NAT UDP nextp: 0x%x ip_src: %x ip_dst: %x\n",ip->nextp, ip->src, ip->dst);
    #endif

      struct egress_nat_value *egress_value_p = 0;
      egress_value_p = get_egress_value(ip->src);
      if(egress_value_p){
      #ifdef BPF_TRACE_EGRESS_UDP
        bpf_trace_printk("[nat-0]: EGRESS NAT UDP  ip->src: %x->%x\n", ip->src ,egress_value_p->ip_src_new);
      #endif

        bpf_l4_csum_replace(skb, UDP_CSUM_OFFSET, bpf_htonl(ip->src), bpf_htonl(egress_value_p->ip_src_new), IS_PSEUDO | 4);
        //change ip
        bpf_l3_csum_replace(skb, IP_CSUM_OFFSET , bpf_htonl(ip->src), bpf_htonl(egress_value_p->ip_src_new), 4);
        ip->src = egress_value_p->ip_src_new;
      }else{
        #ifdef BPF_TRACE_LOOKUP_FAILED
          bpf_trace_printk("[nat-0]: EGRESS NAT UDP LOOKUP FAILED.\n");
        #endif

        #ifndef FORWARD_NO_NAT_TRAFFIC
          goto DROP;
        #endif
      }

      //redirect packet
      pkt_redirect(skb,md,OUT_IFC);
      return RX_REDIRECT;

    REVERSE_UDP: ;
      //Packet coming back, apply reverse nat translaton
      struct reverse_nat_key reverse_key = {};
      reverse_key.ip_dst = ip->dst;

    #ifdef BPF_TRACE_REVERSE_UDP
      bpf_trace_printk("[nat-0]: REVERSE NAT UDP nextp: 0x%x ip_src: %x ip_dst: %x\n",ip->nextp, ip->src, ip->dst);
    #endif

      struct reverse_nat_value *reverse_value_p = 0;
      reverse_value_p = reverse_nat_table.lookup(&reverse_key);
      if(reverse_value_p){
      #ifdef BPF_TRACE_REVERSE_UDP
        bpf_trace_printk("[nat-0]: REVERSE NAT UDP   ip->src: %x->%x\n", ip->src ,egress_value_p->ip_src_new);
      #endif

        bpf_l4_csum_replace(skb, UDP_CSUM_OFFSET, bpf_htonl(ip->dst), bpf_htonl(reverse_value_p->ip_dst_new), IS_PSEUDO | 4);
        bpf_l3_csum_replace(skb, IP_CSUM_OFFSET , bpf_htonl(ip->dst), bpf_htonl(reverse_value_p->ip_dst_new), 4);
        ip->dst = reverse_value_p->ip_dst_new;
      }else{
        #ifdef BPF_TRACE_LOOKUP_FAILED
          bpf_trace_printk("[nat-0]: REVERSE NAT UDP NO MATCH.\n");
        #endif

        #ifndef FORWARD_NO_NAT_TRAFFIC
          goto DROP;
        #endif
      }

      pkt_redirect(skb,md,IN_IFC);
      return RX_REDIRECT;

      /*END UDP*/
      goto EOP;
  }

    tcp: {
      struct tcp_t *tcp = cursor_advance(cursor, sizeof(*tcp));
      /*BEGIN TCP*/
      switch (md->in_ifc){
        case IN_IFC: goto EGRESS_TCP;
        case OUT_IFC: goto REVERSE_TCP;
      }
      goto DROP;

      EGRESS_TCP: ;
      //Packet exiting the nat, apply nat translation
    #ifdef BPF_TRACE_EGRESS_TCP
      bpf_trace_printk("[nat-0]: EGRESS NAT TCP nextp: 0x%x ip_src: %x ip_dst: %x\n",ip->nextp, ip->src, ip->dst);
    #endif

      struct egress_nat_value *egress_value_p = 0;
      egress_value_p = get_egress_value(ip->src);
      if(egress_value_p){
      #ifdef BPF_TRACE_EGRESS_TCP
        bpf_trace_printk("[nat-0]: EGRESS NAT TCP   ip->src: %x->%x\n", ip->src ,egress_value_p->ip_src_new);
      #endif

        bpf_l4_csum_replace(skb, TCP_CSUM_OFFSET, bpf_htonl(ip->src), bpf_htonl(egress_value_p->ip_src_new), IS_PSEUDO | 4);
        //change ip
        bpf_l3_csum_replace(skb, IP_CSUM_OFFSET , bpf_htonl(ip->src), bpf_htonl(egress_value_p->ip_src_new), 4);
        ip->src = egress_value_p->ip_src_new;
      }else{
        #ifdef BPF_TRACE_LOOKUP_FAILED
          bpf_trace_printk("[nat-0]: EGRESS NAT TCP NO MATCH.\n");
        #endif

        #ifndef FORWARD_NO_NAT_TRAFFIC
          goto DROP;
        #endif
      }

      //redirect packet
      pkt_redirect(skb,md,OUT_IFC);
      return RX_REDIRECT;

    REVERSE_TCP: ;
      //Packet coming back, apply reverse nat translaton
      struct reverse_nat_key reverse_key = {};
      reverse_key.ip_dst = ip->dst;

    #ifdef BPF_TRACE_REVERSE_TCP
      bpf_trace_printk("[nat-0]: REVERSE NAT TCP nextp: 0x%x ports: ip_src: %x ip_dst: %x\n",ip->nextp, ip->src, ip->dst);
    #endif

      struct reverse_nat_value *reverse_value_p = 0;
      reverse_value_p = reverse_nat_table.lookup(&reverse_key);
      if(reverse_value_p){
      #ifdef BPF_TRACE_REVERSE_TCP
        // bpf_trace_printk("[nat-0]: REVERSE NAT TCP   ip->dst: %x->%x\n", ip->dst ,reverse_value_p->ip_dst_new);
        bpf_trace_printk("[nat-0]: REVERSE NAT TCP port->dst: %d->%d\n", tcp->dst_port , reverse_value_p->port_dst_new);
      #endif

        bpf_l4_csum_replace(skb, TCP_CSUM_OFFSET, bpf_htonl(ip->dst), bpf_htonl(reverse_value_p->ip_dst_new), IS_PSEUDO | 4);
        bpf_l3_csum_replace(skb, IP_CSUM_OFFSET , bpf_htonl(ip->dst), bpf_htonl(reverse_value_p->ip_dst_new), 4);
        ip->dst = reverse_value_p->ip_dst_new;
      }else{
        #ifdef BPF_TRACE_LOOKUP_FAILED
          bpf_trace_printk("[nat-0]: REVERSE NAT TCP NO MATCH.\n");
        #endif

        #ifndef FORWARD_NO_NAT_TRAFFIC
          goto DROP;
        #endif
      }

      //redirect packet
      pkt_redirect(skb,md,IN_IFC);
      return RX_REDIRECT;

      /*END TCP*/
      goto EOP;
    }

DROP: ;

EOP: ;
#ifdef BPF_TRACE_DROP
  bpf_trace_printk("[nat-0]: DROP packet.\n");
#endif
  return RX_DROP;
}
`
