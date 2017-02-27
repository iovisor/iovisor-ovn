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
package l2switch

var SwitchSecurityPolicy = `
#include <bcc/proto.h>
#include <bcc/helpers.h>

#include <uapi/linux/bpf.h>
#include <uapi/linux/if_ether.h>
#include <uapi/linux/if_packet.h>
#include <uapi/linux/ip.h>
#include <uapi/linux/in.h>
#include <uapi/linux/filter.h>
#include <uapi/linux/pkt_cls.h>

#define BPF_TRACE

#define IP_SECURITY_INGRESS
#define MAC_SECURITY_INGRESS
#define MAC_SECURITY_EGRESS

#define MAX_PORTS 32

struct mac_t {
  u64 mac;
};

struct interface {
  u32 ifindex;
};

struct ifindex{
  u32 ifindex;
};

struct ip_leaf{
  u32 ip;
};

/*
  The Forwarding Table (fwdtable) contains the association between mac Addresses
  and	ports learned by the switch in the learning phase.
  This table is used also in the forwarding phase when the switch has to decide
  the port to use for forwarding the packet.
  The interface number uses the convention of hover, so is an incremental number
  given by hover daemon each time a port is attached to the IOModule (1, 2,..).
*/
BPF_TABLE("hash", struct mac_t, struct interface, fwdtable, 1024);

/*
  The Security Mac Table (securitymac) associate to each port the allowed mac
  address. If no entry is associated with the port, the port security is not
  applied to the port.
*/
BPF_TABLE("hash", struct ifindex, struct mac_t, securitymac, MAX_PORTS + 1);

/*
  The Security Ip Table (securityip) associate to each port the allowed ip
  address. If no entry is associated with the port, the port security is not
  applied to the port.
*/
BPF_TABLE("hash", struct ifindex, struct ip_leaf, securityip, MAX_PORTS + 1);

struct eth_hdr {
  u64   dst:48;
  u64   src:48;
  u16   proto;
} __attribute__((packed));

static int handle_rx(void *skb, struct metadata *md) {
  struct __sk_buff *skb2 = (struct __sk_buff *)skb;
  void *data = (void *)(long)skb2->data;
  void *data_end = (void *)(long)skb2->data_end;
  struct eth_hdr *eth = data;

  if (data + sizeof(*eth) > data_end)
    return RX_DROP;

  #ifdef BPF_TRACE
    bpf_trace_printk("[switch-%d]: in_ifc=%d\n", md->module_id, md->in_ifc);
  #endif

  //set in-interface for lookup ports security
  struct ifindex in_iface = {};
  in_iface.ifindex = md->in_ifc;

  //port security on source mac
  #ifdef MAC_SECURITY_INGRESS
  struct mac_t * mac_lookup;
  mac_lookup = securitymac.lookup(&in_iface);
  if (mac_lookup)
    if (eth->src != mac_lookup->mac) {
      #ifdef BPF_TRACE
        bpf_trace_printk("[switch-%d]: mac INGRESS %lx mismatch %lx -> DROP\n",
          md->module_id, eth->src, mac_lookup->mac);
      #endif
      return RX_DROP;
    }
  #endif

  //port security on source ip
  #ifdef IP_SECURITY_INGRESS
  if (eth->proto == bpf_htons(ETH_P_IP)) {
    struct ip_leaf *ip_lookup = securityip.lookup(&in_iface);
    if (ip_lookup) {
      struct ip_t *ip = data + sizeof(*eth);
      if (data + sizeof(*eth) + sizeof(*ip) > data_end)
        return RX_DROP;

      if (ip->src != ip_lookup->ip) {
        #ifdef BPF_TRACE
          bpf_trace_printk("[switch-%d]: IP INGRESS %x mismatch %x -> DROP\n", md->module_id, ip->src, ip_lookup->ip);
        #endif
        return RX_DROP;
      }
    }
  }
  #endif

  #ifdef BPF_TRACE
    bpf_trace_printk("[switch-%d]: mac src:%lx dst:%lx\n", md->module_id, eth->src, eth->dst);
  #endif

  //LEARNING PHASE: mapping in_iface with src_interface
  struct mac_t src_key = {};
  struct interface interface = {};

  //set in_iface as key
  src_key.mac = (u64) eth->src;

  //set in_ifc, and 0 counters as leaf
  interface.ifindex = md->in_ifc;

  //lookup in fwdtable. if no key present initialize with interface
  struct interface *interface_lookup = fwdtable.lookup_or_init(&src_key, &interface);

  //if the same mac has changed interface, update it
  if (interface_lookup->ifindex != md->in_ifc)
    interface_lookup->ifindex = md->in_ifc;

  //FORWARDING PHASE: select interface(s) to send the packet
  struct mac_t dst_mac = {(u64) eth->dst};

  //lookup in forwarding table fwdtable
  struct interface *dst_interface = fwdtable.lookup(&dst_mac);

  if (dst_interface) {
    //HIT in forwarding table
    //redirect packet to dst_interface

    #ifdef MAC_SECURITY_EGRESS
    struct mac_t * mac_lookup;
    struct ifindex out_iface = {};
    out_iface.ifindex = dst_interface->ifindex;
    mac_lookup = securitymac.lookup(&out_iface);
    if (mac_lookup)
      if (eth->dst != mac_lookup->mac){
        #ifdef BPF_TRACE
          bpf_trace_printk("[switch-%d]: mac EGRESS %lx mismatch %lx -> DROP\n",
            md->module_id, eth->dst, mac_lookup->mac);
        #endif
        return RX_DROP;
      }
    #endif

    /* do not send packet back on the ingress interface */
    if (dst_interface->ifindex == md->in_ifc)
      return RX_DROP;

    pkt_redirect(skb, md, dst_interface->ifindex);

    #ifdef BPF_TRACE
      bpf_trace_printk("[switch-%d]: redirect out_ifc=%d\n", md->module_id, dst_interface->ifindex);
    #endif

    return RX_REDIRECT;

  } else {
    #ifdef BPF_TRACE
      bpf_trace_printk("[switch-%d]: Broadcast\n", md->module_id);
    #endif
    pkt_controller(skb, md, PKT_BROADCAST);
    return RX_CONTROLLER;
  }
}
`
