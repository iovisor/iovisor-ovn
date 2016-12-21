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
package l2switch

var SwitchSecurityPolicy = `
#define BPF_TRACE
//#undef BPF_TRACE

//Ports 32+1
#define MAX_PORTS 33

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
BPF_TABLE("hash", struct mac_t, struct interface, fwdtable, 10240);

/*
  The Ports Table (ports) is a fixed length array that identifies the fd (file
  descriptors) of the network interfaces attached to the switch.
  This is a workaround for broadcast implementation, in order to be able to call
  bpf_clone_redirect that accepts as parameter the fd of the network interface.
  This array is not ordered. The index of the array does NOT represent the
  interface number.
*/
BPF_TABLE("array", u32, u32, ports, MAX_PORTS);

/*
  The Security Mac Table (securitymac) associate to each port the allowed mac
  address. If no entry is associated with the port, the port security is not
  applied to the port.
*/
BPF_TABLE("hash", struct ifindex, struct mac_t, securitymac, MAX_PORTS*2);

/*
  The Security Ip Table (securityip) associate to each port the allowed ip
  address. If no entry is associated with the port, the port security is not
  applied to the port.
*/
BPF_TABLE("hash", struct ifindex, struct ip_leaf, securityip, MAX_PORTS*2);

static int handle_rx(void *skb, struct metadata *md) {
  u8 *cursor = 0;
  struct ethernet_t *ethernet = cursor_advance(cursor, sizeof(*ethernet));

  #ifdef BPF_TRACE
    bpf_trace_printk("[switch-%d]: in_ifc=%d\n", md->module_id, md->in_ifc);
  #endif

  //set in-interface for lookup ports security
  struct ifindex in_iface = {};
  in_iface.ifindex = md->in_ifc;

  //port security on source mac
  struct mac_t * mac_lookup;
  mac_lookup = securitymac.lookup(&in_iface);
  if (mac_lookup)
    if (ethernet->src != mac_lookup->mac) {
      #ifdef BPF_TRACE
        bpf_trace_printk("[switch-%d]: mac %lx mismatch %lx -> DROP\n",
          md->module_id, ethernet->src, mac_lookup->mac);
      #endif
      return RX_DROP;
    }

  //port security on source ip
  if (ethernet->type == 0x0800) {
    struct ip_leaf *ip_lookup;
    ip_lookup = securityip.lookup(&in_iface);
    if (ip_lookup) {
      struct ip_t *ip = cursor_advance(cursor, sizeof(*ip));
      if (ip->src != ip_lookup->ip) {
        #ifdef BPF_TRACE
          bpf_trace_printk("[switch-%d]: IP %x mismatch %x -> DROP\n", md->module_id, ip->src, ip_lookup->ip);
        #endif
        return RX_DROP;
      }
    }
  }

  #ifdef BPF_TRACE
    bpf_trace_printk("[switch-%d]: mac src:%lx dst:%lx\n", md->module_id, ethernet->src, ethernet->dst);
  #endif

  //LEARNING PHASE: mapping in_iface with src_interface
  struct mac_t src_key = {};
  struct interface interface = {};

  //set in_iface as key
  src_key.mac = ethernet->src;

  //set in_ifc, and 0 counters as leaf
  interface.ifindex = md->in_ifc;

  //lookup in fwdtable. if no key present initialize with interface
  struct interface *interface_lookup = fwdtable.lookup_or_init(&src_key, &interface);

  //if the same mac has changed interface, update it
  if (interface_lookup->ifindex != md->in_ifc)
    interface_lookup->ifindex = md->in_ifc;

  //FORWARDING PHASE: select interface(s) to send the packet
  struct mac_t dst_mac = {ethernet->dst};

  //lookup in forwarding table fwdtable
  struct interface *dst_interface = fwdtable.lookup(&dst_mac);

  if (dst_interface) {
    //HIT in forwarding table
    //redirect packet to dst_interface
    pkt_redirect(skb, md, dst_interface->ifindex);

    #ifdef BPF_TRACE
      bpf_trace_printk("[switch-%d]: redirect out_ifc=%d\n", md->module_id, dst_interface->ifindex);
    #endif

    return RX_REDIRECT;

  } else {
    //MISS in forwarding table
    #ifdef BPF_TRACE
      bpf_trace_printk("[switch-%d]s: broadcast\n", md->module_id);
    #endif

    u32 i = 0;
    u32 t;
    #pragma unroll
    for (i = 1; i <= 32; i++) {
      u32 *iface_p;
      // For some reason the compiler does not unroll the loop if the 'i'
      // variable is used in the lookup function
      t = i;
      iface_p = ports.lookup(&t);

      if (iface_p)
        if (*iface_p != 0)
          bpf_clone_redirect(skb, *iface_p, 0);
    }

    return RX_DROP;
  }
}
`
