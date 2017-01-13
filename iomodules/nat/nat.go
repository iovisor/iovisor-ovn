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
package nat

var NatCode = `
#include <uapi/linux/bpf.h>
#include <uapi/linux/if_ether.h>
#include <uapi/linux/if_packet.h>
#include <uapi/linux/ip.h>
#include <uapi/linux/in.h>
#include <uapi/linux/filter.h>
#include <uapi/linux/pkt_cls.h>

#include <bcc/proto.h>

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

#define PUBLIC_IP 0x0a0a0164 //10.10.1.100

/*Egress Nat key*/
struct egress_nat_key{
  u32 ip_src;
  u32 ip_dst;
  u16 port_src;
  u16 port_dst;
};

/*Egress Nat Value*/
struct egress_nat_value{
  u32 ip_src_new;
  u16 port_src_new;
};

/*Reverse Nat key*/
struct reverse_nat_key{
  u32 ip_src;
  u32 ip_dst;
  u16 port_src;
  u16 port_dst;
};

/*Reverse Nat Value*/
struct reverse_nat_value{
  u32 ip_dst_new;
  u16 port_dst_new;
};

/*
  Egress Nat Table. This table translate and mantains the association between
  (ip_src, ip_dst, port_src, port_dst) and the correspondent mapping into
  new (new_ip_src, new_port_src).
*/
BPF_TABLE("hash", struct egress_nat_key, struct egress_nat_value, egress_nat_table, EGRESS_NAT_TABLE_DIM);

/*
  Reverse Nat Table. This table contains the association for the reverse
  translation of packet of a session coming back to the nat.
*/
BPF_TABLE("hash", struct reverse_nat_key, struct reverse_nat_value, reverse_nat_table, REVERSE_NAT_TABLE_DIM);

/*
  First implementation of a pool of ports. (incremental counter).
*/
BPF_TABLE("array", u32, u16, first_free_port, 1);

/*
  Allocates and returns the first free port for instantiate a new session.
*/
static inline u16 get_free_port(){
  u32 i = 0;
  u16 *new_port_p = 0;
  new_port_p = first_free_port.lookup(&i);
  if (new_port_p){
    //TODO Remove (to be performed by ControlPlane)
    if(*new_port_p == 0)
      *new_port_p = 1024;
    *new_port_p = *new_port_p + 1;
    return *new_port_p;
  }
  return 0;
}

/*given ip(src,dst) and port(src,dst) returns the correspondent association in
  the nat table.
  If no association is present, allocates egress and reverse nat table entries,
  and returns the new association just created.
*/
static inline struct egress_nat_value * get_egress_value(u32 ip_src, u32 ip_dst, u16 port_src, u16 port_dst){
  //Lookup in the egress_nat_table
  struct egress_nat_key egress_key = {};
  egress_key.ip_src = ip_src;
  egress_key.ip_dst = ip_dst;
  egress_key.port_src = port_src;
  egress_key.port_dst = port_dst;

  struct egress_nat_value *egress_value_p = 0;
  egress_value_p = egress_nat_table.lookup(&egress_key);
  if (!egress_value_p){
    //Create rule for egress
    struct egress_nat_value egress_value = {};
    egress_value.ip_src_new = PUBLIC_IP;
    egress_value.port_src_new = get_free_port();

    //push egress rule
    egress_nat_table.lookup_or_init(&egress_key, &egress_value);
    egress_value_p = &egress_value;

    //Create rule for reverse
    struct reverse_nat_key reverse_key= {};
    reverse_key.ip_src = egress_key.ip_dst;
    reverse_key.ip_dst = egress_value.ip_src_new;
    reverse_key.port_src = egress_key.port_dst;
    reverse_key.port_dst = egress_value.port_src_new;

    struct reverse_nat_value reverse_value= {};
    reverse_value.ip_dst_new = egress_key.ip_src;
    reverse_value.port_dst_new = egress_key.port_src;

    //push reverse nat rule
    reverse_nat_table.lookup_or_init(&reverse_key, &reverse_value);
  }
  return egress_value_p;
}

static int handle_rx(void *skb, struct metadata *md) {
  u8 *cursor = 0;
  struct ethernet_t *ethernet = cursor_advance(cursor, sizeof(*ethernet));

  // bpf_trace_printk("[nat-0]: eth_type:%x mac_src:%lx mac_dst:%lx\n",
  // ethernet->type, ethernet->src, ethernet->dst);

  switch (ethernet->type){
    case ETH_TYPE_IP: goto IP;
    case ETH_TYPE_ARP: goto DROP;
    default: goto EOP;
  }

  IP: ;
  struct ip_t *ip = cursor_advance(cursor, sizeof(*ip));

  switch (ip->nextp){
    case IP_UDP: goto udp;
    case IP_TCP: goto tcp;
    default: goto EOP;
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
      // bpf_trace_printk("[nat-0]: EGRESS NAT UDP nextp: 0x%x ports: %d->%d\n",ip->nextp, udp->sport,udp->dport);

      struct egress_nat_value *egress_value_p = 0;
      egress_value_p = get_egress_value(ip->src, ip->dst, udp->sport, udp->dport);
      if(egress_value_p){
        // bpf_trace_printk("[nat-0]: EGRESS NAT UDP   ip->src: %x->%x\n", ip->src ,egress_value_p->ip_src_new);
        // bpf_trace_printk("[nat-0]: EGRESS NAT UDP port->src: %d->%d\n", udp->sport , egress_value_p->port_src_new);

        bpf_l4_csum_replace(skb, UDP_CSUM_OFFSET, bpf_htonl(ip->src), bpf_htonl(egress_value_p->ip_src_new), IS_PSEUDO | 4);
        bpf_l4_csum_replace(skb, UDP_CSUM_OFFSET, bpf_htons(udp->sport), bpf_htons(egress_value_p->port_src_new), 2);
        udp->sport = egress_value_p->port_src_new;

        //change ip
        bpf_l3_csum_replace(skb, IP_CSUM_OFFSET , bpf_htonl(ip->src), bpf_htonl(egress_value_p->ip_src_new), 4);
        ip->src = egress_value_p->ip_src_new;
      }else{
        //bpf_trace_printk("[nat-0]: EGRESS NAT UDP LOOKUP FAILED -> DROP packet.\n");
        goto DROP;
      }

      //redirect packet
      pkt_redirect(skb,md,OUT_IFC);
      return RX_REDIRECT;

    REVERSE_UDP: ;
      //Packet coming back, apply reverse nat translaton
      struct reverse_nat_key reverse_key = {};
      reverse_key.ip_src = ip->src;
      reverse_key.ip_dst = ip->dst;
      reverse_key.port_src = udp->sport;
      reverse_key.port_dst = udp->dport;

      // bpf_trace_printk("[nat-0]: REVERSE NAT UDP nextp: 0x%x ports: %d->%d\n",ip->nextp, udp->sport,udp->dport);

      struct reverse_nat_value *reverse_value_p = 0;
      reverse_value_p = reverse_nat_table.lookup(&reverse_key);
      if(reverse_value_p){
        // bpf_trace_printk("[nat-0]: REVERSE NAT UDP   ip->src: %x->%x\n", ip->src ,egress_value_p->ip_src_new);
        // bpf_trace_printk("[nat-0]: REVERSE NAT UDP port->src: %d->%d\n", udp->sport , egress_value_p->port_src_new);

        bpf_l4_csum_replace(skb, UDP_CSUM_OFFSET, bpf_htonl(ip->dst), bpf_htonl(reverse_value_p->ip_dst_new), IS_PSEUDO | 4);
        bpf_l4_csum_replace(skb, UDP_CSUM_OFFSET, bpf_htons(udp->dport), bpf_htons(reverse_value_p->port_dst_new), 2);
        udp->dport = reverse_value_p->port_dst_new;

        bpf_l3_csum_replace(skb, IP_CSUM_OFFSET , bpf_htonl(ip->dst), bpf_htonl(reverse_value_p->ip_dst_new), 4);
        ip->dst = reverse_value_p->ip_dst_new;

        pkt_redirect(skb,md,IN_IFC);
        return RX_REDIRECT;

      }else{
        // bpf_trace_printk("[nat-0]: REVERSE NAT UDP NO MATCH -> DROP packet.\n");
        goto DROP;
      }
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
      // bpf_trace_printk("[nat-0]: EGRESS NAT TCP nextp: 0x%x ports: %d->%d\n",ip->nextp, tcp->src_port,tcp->dst_port);

      struct egress_nat_value *egress_value_p = 0;
      egress_value_p = get_egress_value(ip->src, ip->dst, tcp->src_port, tcp->dst_port);
      if(egress_value_p){
        // bpf_trace_printk("[nat-0]: EGRESS NAT TCP   ip->src: %x->%x\n", ip->src ,egress_value_p->ip_src_new);
        // bpf_trace_printk("[nat-0]: EGRESS NAT TCP port->src: %d->%d\n", tcp->src_port , egress_value_p->port_src_new);

        bpf_l4_csum_replace(skb, TCP_CSUM_OFFSET, bpf_htonl(ip->src), bpf_htonl(egress_value_p->ip_src_new), IS_PSEUDO | 4);
        bpf_l4_csum_replace(skb, TCP_CSUM_OFFSET, bpf_htons(tcp->src_port), bpf_htons(egress_value_p->port_src_new), 2);
        tcp->src_port = egress_value_p->port_src_new;

        //change ip
        bpf_l3_csum_replace(skb, IP_CSUM_OFFSET , bpf_htonl(ip->src), bpf_htonl(egress_value_p->ip_src_new), 4);
        ip->src = egress_value_p->ip_src_new;
      }else{
        //bpf_trace_printk("[nat-0]: EGRESS NAT TCP LOOKUP FAILED -> DROP packet.\n");
        goto DROP;
      }

      //redirect packet
      pkt_redirect(skb,md,OUT_IFC);
      return RX_REDIRECT;

    REVERSE_TCP: ;
      //Packet coming back, apply reverse nat translaton
      struct reverse_nat_key reverse_key = {};
      reverse_key.ip_src = ip->src;
      reverse_key.ip_dst = ip->dst;
      reverse_key.port_src = tcp->src_port;
      reverse_key.port_dst = tcp->dst_port;

      // bpf_trace_printk("[nat-0]: REVERSE NAT TCP nextp: 0x%x ports: %d->%d\n",ip->nextp, tcp->src_port,tcp->dst_port);

      struct reverse_nat_value *reverse_value_p = 0;
      reverse_value_p = reverse_nat_table.lookup(&reverse_key);
      if(reverse_value_p){
        // bpf_trace_printk("[nat-0]: REVERSE NAT TCP   ip->dst: %x->%x\n", ip->dst ,reverse_value_p->ip_dst_new);
        // bpf_trace_printk("[nat-0]: REVERSE NAT TCP port->dst: %d->%d\n", tcp->dst_port , reverse_value_p->port_dst_new);

        bpf_l4_csum_replace(skb, TCP_CSUM_OFFSET, bpf_htonl(ip->dst), bpf_htonl(reverse_value_p->ip_dst_new), IS_PSEUDO | 4);
        bpf_l4_csum_replace(skb, TCP_CSUM_OFFSET, bpf_htons(tcp->dst_port), bpf_htons(reverse_value_p->port_dst_new), 2);
        tcp->dst_port = reverse_value_p->port_dst_new;

        bpf_l3_csum_replace(skb, IP_CSUM_OFFSET , bpf_htonl(ip->dst), bpf_htonl(reverse_value_p->ip_dst_new), 4);
        ip->dst = reverse_value_p->ip_dst_new;

        //redirect packet
        pkt_redirect(skb,md,IN_IFC);
        return RX_REDIRECT;
      }else{
        // bpf_trace_printk("[nat-0]: REVERSE NAT TCP NO MATCH -> DROP packet.\n");
        goto DROP;
      }
      /*END TCP*/
      goto EOP;
    }

DROP: ;

EOP: ;
  // bpf_trace_printk("[nat-0]: DROP packet.\n");
  return RX_DROP;
}
`
