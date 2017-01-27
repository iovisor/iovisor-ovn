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

/* module configuration parameters */
#define POOL_SIZE 10
#define N_DHCP_OPTIONS 10 // max number of options that can be processed

/* DHCP protocol definitions
 * took from: https://github.com/aldebaran/connman/blob/master/gdhcp/common.h */
#define DHCP_MAGIC              0x63825363
#define DHCP_OPTIONS_BUFSIZE    308
#define BOOTREQUEST             1
#define BOOTREPLY               2

#define BROADCAST_FLAG	0x8000

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


#define DHCP_PADDING		0x00
#define DHCP_SUBNET		0x01
#define DHCP_ROUTER		0x03
#define DHCP_TIME_SERVER	0x04
#define DHCP_NAME_SERVER	0x05
#define DHCP_DNS_SERVER		0x06
#define DHCP_HOST_NAME		0x0c
#define DHCP_DOMAIN_NAME	0x0f
#define DHCP_NTP_SERVER		0x2a
#define DHCP_REQUESTED_IP	0x32
#define DHCP_LEASE_TIME		0x33
#define DHCP_OPTION_OVERLOAD	0x34
#define DHCP_MESSAGE_TYPE	0x35
#define DHCP_SERVER_ID		0x36
#define DHCP_PARAM_REQ		0x37
#define DHCP_ERR_MESSAGE	0x38
#define DHCP_MAX_SIZE		0x39
#define DHCP_VENDOR		0x3c
#define DHCP_CLIENT_ID		0x3d
#define DHCP_END		0xff

#define OPT_CODE		0
#define OPT_LEN			1
#define OPT_DATA		2
#define OPTION_FIELD		0
#define FILE_FIELD		1
#define SNAME_FIELD		2

/* DHCP_MESSAGE_TYPE values */
#define DHCPDISCOVER		1
#define DHCPOFFER		2
#define DHCPREQUEST		3
#define DHCPDECLINE		4
#define DHCPACK			5
#define DHCPNAK			6
#define DHCPRELEASE		7
#define DHCPINFORM		8
#define DHCP_MINTYPE DHCPDISCOVER
#define DHCP_MAXTYPE DHCPINFORM

#define DHCP_OPTIONS_OFFSET (sizeof(struct ethernet_t) + sizeof(struct ip_t) + \
  sizeof(struct udp_t) + sizeof(struct dhcp_packet))
#define UDP_CRC_OFF (sizeof(struct ethernet_t) + sizeof(struct ip_t) + \
  offsetof(struct udp_t, crc))
#define IP_CRC_OFF (sizeof(struct ethernet_t) + offsetof(struct ip_t, hchecksum))

#define ETHERNET_BROADCAST 0xffffffffffffULL
#define IP_BROADCAST 0xffffffffUL // 255.255.255.255

#define IP_CSUM_OFFSET 10
#define IS_PSEUDO 0x10

enum {
  AVAILABLE = 0, /* IP can be assigned to a client */
  OFFERED, /* IP has been offered to a client but it has not requested it yet */
  LEASED /* IP has been assigned to a client */
};

struct pool_entry {
  u32 ip;
  u32 status; /* AVAILABLE, OFFERED or LEASED */
  u64 time; /* when the IP was offered or leased */
  u64 mac; /* mac of the cient that is using this ip*/
};

struct config_t {
  u64 subnet_mask;
  u32 dns;
  u32 lease_time;
  u32 router;
  u32 server_ip;
  u64 server_mac;
};

/* pool_entry: contains all the status of the server */
BPF_TABLE("array", unsigned int, struct pool_entry, pool, POOL_SIZE);

/* ip_to_index: allows to get the index within the pool_entry map of a given ip */
BPF_TABLE("hash", u32, unsigned int, ip_to_index, POOL_SIZE);

/* config: contains the configuration of the module */
BPF_TABLE("array", unsigned int, struct config_t, config, 1);

struct dhcp_options {
  u8 message_type_id;
  u8 message_type_length;
  u8 message_type;

  u8 server_id_id;
  u8 server_id_length;
  u32 server_id;

  u8 lease_time_id;
  u8 lease_time_length;
  u32 lease_time;

  u8 subnet_mask_id;
  u8 subnet_mask_length;
  u32 subnet_mask;

  u8 router_id;
  u8 router_length;
  u32 router;

  u8 dns_id;
  u8 dns_length;
  u32 dns;

  u8 end;
  u8 padding[2];

} __attribute__((packed));

static inline void fill_dhcp_options(struct dhcp_options *r) {
  r->message_type_id = DHCP_MESSAGE_TYPE;
  r->message_type_length = 1;

  r->server_id_id = DHCP_SERVER_ID;
  r->server_id_length = 4;

  r->lease_time_id = DHCP_LEASE_TIME;
  r->lease_time_length = 4;

  r->subnet_mask_id = DHCP_SUBNET;
  r->subnet_mask_length = 4;

  r->router_id = DHCP_ROUTER;
  r->router_length = 4;

  r->dns_id = DHCP_DNS_SERVER;
  r->dns_length = 4;

  r->end = DHCP_END;
}

static inline struct pool_entry *find_first_available_ip(void) {
  unsigned int i;
  struct pool_entry *ret;
  #pragma unroll
  for (i = 0; i < POOL_SIZE; i++) {
    unsigned int t = i; /* workaround for unrolling loop problem*/
    ret = pool.lookup(&t);
    if (!ret) /* it should never happen as map is of type array */
      return NULL;
    if (ret->status == AVAILABLE)
      return ret;
  }
  /*TODO: does it make sense to look for offered Ips here?*/
  return NULL;
}

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

  /* controls about hardware address lenght and this stuff are omitted */
  u8 len = 0;

  /* options to be parsed */
  u32 requested_ip = 0;
  u8 message_type = 0;
  u32 server_id = 0;

  unsigned short i = 0;
  unsigned int index = DHCP_OPTIONS_OFFSET;

  #pragma unroll
  for (i = 0; i < N_DHCP_OPTIONS; i++) {
    u8 option = load_byte(skb, index++);
    if (option == DHCP_END)
      break;

    u8 len = load_byte(skb, index++);

    switch(option) {
      case DHCP_MESSAGE_TYPE:
        message_type = load_byte(skb, index);
        break;
      case DHCP_REQUESTED_IP:
        requested_ip = load_word(skb, index);
        break;
      case DHCP_SERVER_ID:
        server_id = load_word(skb, index);
        break;
    }
    index += len;
  }

  struct pool_entry *ip_to_offer = 0;
  struct dhcp_options options = {};

  int csum;
  int err;

  if (message_type == DHCPDISCOVER) {
    //bpf_trace_printk("[dhcp] DHCPDISCOVER\n");
    ip_to_offer = find_first_available_ip();
    if (!ip_to_offer)
      goto DROP;
    options.message_type = DHCPOFFER;
    ip_to_offer->status = OFFERED;
    /* TODO: set leased time */
    goto REPLY;
  }

  if (message_type == DHCPREQUEST) {
    //bpf_trace_printk("[dhcp] DHCPREQUEST\n");
    /*
      * A dhcp request message can arrive for three different reasons:
      * 1. As a response to a dhcp offer message:
      *  Server id and request ip options are present
      * 2. A client wants to validate a previous known address:
      *  Only the request ip option is present
      * 3. A client wants to extend the lease:
      *  Nor the server id or the request ip options are present. Ip that wants
      *  to be re-leased is in 'ciaddr'
      */
    if (!server_id) { /* this is a reply to an offer message */
      //bpf_trace_printk("[dhcp] no server id\n");
      goto DROP;
    }

    u32 server_ip = cfg->server_ip;
    if (server_id != server_ip) {
      /* TODO: set IP status to available*/
      //bpf_trace_printk("[dhcp] server_id != server_ip. %x\n", server_id);
      goto DROP;
    }
    int *index = ip_to_index.lookup(&requested_ip);
    if (!index) {
      /* TODO: send NACK*/
      //bpf_trace_printk("[dhcp] no index found\n");
      goto DROP;
    }

    int index2 = *index;
    ip_to_offer = pool.lookup(&index2);
    if (!ip_to_offer) { /* This error should be impossible. */
      goto DROP;
    }

    if (ip_to_offer->status != OFFERED || ip_to_offer->ip != requested_ip) {
      /* TODO: send NACK */
      //bpf_trace_printk("[dhcp] ip_to_offer->ip != requested_ip\n");
      goto DROP;
    }

    ip_to_offer->status = LEASED;
    options.message_type = DHCPACK;
    /* TODO: set lease time */
    goto REPLY;
  }

  /* TODO: support other dhcpmessages */
  goto DROP;

REPLY:
  err = bpf_l4_csum_replace(skb, UDP_CRC_OFF, bpf_htonl(dhcp->yiaddr),
    bpf_htonl(ip_to_offer->ip), IS_PSEUDO | 4);
  if (err) {
    goto ERROR;
  }
  dhcp->yiaddr = ip_to_offer->ip;

  err = bpf_l4_csum_replace(skb, UDP_CRC_OFF, bpf_htonl(dhcp->siaddr_nip),
    bpf_htonl(cfg->server_ip), IS_PSEUDO | 4);
  if (err) {
    goto ERROR;
  }
  dhcp->siaddr_nip = cfg->server_ip;

  fill_dhcp_options(&options);
  options.server_id = bpf_htonl(cfg->server_ip);
  options.lease_time = bpf_htonl(cfg->lease_time);
  options.subnet_mask = bpf_htonl(cfg->subnet_mask);
  options.router = bpf_htonl(cfg->router);
  options.dns = bpf_htonl(cfg->dns);

  /* TODO: trim or extend packet before copying options in */

  struct __sk_buff *skb2 = (struct __sk_buff *) skb;
  void *data = (void *)(long)skb2->data;
  void *data_end = (void *)(long)skb2->data_end;
  /* be sure that we access to allow content on the packet*/
  if (data + DHCP_OPTIONS_OFFSET + sizeof(options) > data_end)
    goto DROP;
  csum = bpf_csum_diff(data + DHCP_OPTIONS_OFFSET, sizeof(options), &options,
    sizeof(options), 0);
  //bpf_trace_printk("[dhcp] csum: %d\n", csum);
  err = bpf_l4_csum_replace(skb, UDP_CRC_OFF, 0, csum, 0);
  if (err) {
    goto ERROR;
  }

  err = bpf_skb_store_bytes(skb, DHCP_OPTIONS_OFFSET, &options, sizeof(options), 0);
  if (err) {
    goto ERROR;
  }

  /* set packets field before sending it */
  /* L4 */
  err = bpf_l4_csum_replace(skb, UDP_CRC_OFF, bpf_htons(udp->dport), bpf_htons(68),
    IS_PSEUDO | 2);
  if (err) {
    goto ERROR;
  }
  udp->dport = 68;

  err = bpf_l4_csum_replace(skb, UDP_CRC_OFF, bpf_htons(udp->sport), bpf_htons(67),
    IS_PSEUDO | 2);
  if (err) {
    goto ERROR;
  }
  udp->sport = 67;

  /* L3 */
  err = bpf_l3_csum_replace(skb, IP_CRC_OFF, bpf_htonl(ip->dst),
    bpf_htonl((IP_BROADCAST)), 4);
  if (err) {
    goto ERROR;
  }

  err = bpf_l4_csum_replace(skb, UDP_CRC_OFF, bpf_htonl(ip->dst),
    bpf_htonl((IP_BROADCAST)), IS_PSEUDO | 4);
  if (err) {
    goto ERROR;
  }
  ip->dst = (IP_BROADCAST);

  err = bpf_l4_csum_replace(skb, UDP_CRC_OFF, bpf_htonl(ip->src),
    bpf_htonl(cfg->server_ip), IS_PSEUDO | 4);
  if (err) {
    goto ERROR;
  }

  err = bpf_l3_csum_replace(skb, IP_CRC_OFF, bpf_htonl(ip->src),
    bpf_htonl(cfg->server_ip), 4);
  if (err) {
    goto ERROR;
  }
  ip->src = cfg->server_ip;

  /* L2 */
  ethernet->src = cfg->server_mac;

  pkt_redirect(skb, md, md->in_ifc);
  //bpf_trace_printk("[dhcp] sending packet back\n");
  return RX_REDIRECT;

DROP:
  //bpf_trace_printk("[dhcp] dropping packet\n");
  return RX_DROP;

ERROR:
  //bpf_trace_printk("[dhcp] ERROR: %d\n", err);
  return RX_DROP;
}
`
