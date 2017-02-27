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
package dhcp

// This module perform a filtering on the packets that arrive, only
// DHCP packets destinated to the server IP and MAC addressed (or to
// brodcast address) are sent to the control plane

var DhcpServer = `
#include <bcc/proto.h>
#include <bcc/helpers.h>

#include <uapi/linux/bpf.h>
#include <uapi/linux/if_ether.h>
#include <uapi/linux/if_packet.h>
#include <uapi/linux/ip.h>
#include <uapi/linux/in.h>
#include <uapi/linux/filter.h>
#include <uapi/linux/pkt_cls.h>

#define ETHERNET_BROADCAST 0xffffffffffffULL
#define IP_BROADCAST 0xffffffffUL // 255.255.255.255

/* See RFC 2131 */
struct dhcp_packet {
  uint8_t op;
  uint8_t htype;
  uint8_t hlen;
  uint8_t hops;
  uint32_t xid;
  uint16_t secs;
  uint16_t flags;
  uint32_t ciaddr;
  uint32_t yiaddr;
  uint32_t siaddr_nip;
  uint32_t gateway_nip;
  uint8_t chaddr[16];
  uint8_t sname[64];
  uint8_t file[128];
  uint32_t cookie;
  //uint8_t options[DHCP_OPTIONS_BUFSIZE + EXTEND_FOR_BUGGY_SERVERS];
} BPF_PACKET_HEADER;

struct config_t {
  u32 server_ip;
  u64 server_mac;
};

/* config: contains the configuration of the module */
BPF_TABLE("array", unsigned int, struct config_t, config, 1);

static int handle_rx(void *skb, struct metadata *md) {
  u8 *cursor = 0;
  unsigned int zero = 0;

  /* get server configuration */
  struct config_t *cfg = config.lookup(&zero);
  if (!cfg) {
    //bpf_trace_printk("[dhcp] no config found\n");
    goto DROP;
  }

  /* check that the packet is a dhcp packet*/
  struct ethernet_t *ethernet = cursor_advance(cursor, sizeof(*ethernet));

  if (ethernet->dst != ETHERNET_BROADCAST && ethernet->dst != cfg->server_mac) {
    //bpf_trace_printk("[dhcp] no dst mac\n");
    goto DROP;
  }

  if (ethernet->type != 0x0800) {
    //bpf_trace_printk("[dhcp] no IP\n");
    goto DROP;
  }

  struct ip_t *ip = cursor_advance(cursor, sizeof(*ip));

  if (ip->dst != IP_BROADCAST && ip->dst != cfg->server_ip) {
    goto DROP;
  }

  // is the packet UDP?
  if (ip->nextp != 0x11) {
    //bpf_trace_printk("[dhcp] packet is not udp\n");
    goto DROP;
  }

  struct udp_t *udp = cursor_advance(cursor, sizeof(*udp));

  if (udp->dport != 67 || udp->sport != 68) {
    //bpf_trace_printk("[dhcp] packet has invalid port numbers\n");
    goto DROP;
  }

  struct dhcp_packet *dhcp = cursor_advance(cursor, sizeof(*dhcp));

  if (dhcp->op != 1) {
    //bpf_trace_printk("[dhcp] dhcp packet has dhcp->op != 1\n");
    goto DROP;
  }

  bpf_trace_printk("[dhcp] send packet to controller\n");
  return RX_CONTROLLER;

DROP:
  bpf_trace_printk("[dhcp] dropping packet\n");
  return RX_DROP;
}
`
