#include <bcc/proto.h>
#include <uapi/linux/pkt_cls.h>

enum {
	RX_OK,
	RX_REDIRECT,
	RX_DROP,
	RX_RECIRCULATE,
	RX_ERROR,
};

struct chain {
	u32 hops[4];
};
static inline u16 chain_ifc(struct chain *c, int id) {
	return c->hops[id] >> 16;
}
static inline u16 chain_module(struct chain *c, int id) {
	return c->hops[id] & 0xffff;
}

struct type_value {
	u64 type:8;
	u64 value:56;
};
struct metadata {
	// An array of type/value pairs for the module to do with as it pleases. The
	// array is initialized to zero when the event first enters the module chain.
	// The values are preserved across modules.
	struct type_value data[4];

	// A field reserved for use by the wrapper and helper functions.
	u32 is_egress:1;
	u32 flags:31;

	// The length of the packet currently being processed. Read-only.
	u32 pktlen;

	// The module id currently processing the packet.
	int module_id;

	// The interface on which a packet was received. Numbering is local to the
	// module.
	int in_ifc;

	// If the module intends to forward the packet, it must call pkt_redirect to
	// set this field to determine the next-hop.
	int redir_ifc;

	int clone_ifc;
};

// iomodule must implement this function to attach to the networking stack
static int handle_rx(void *pkt, struct metadata *md);
static int handle_tx(void *pkt, struct metadata *md);

static int pkt_redirect(void *pkt, struct metadata *md, int ifc);
static int pkt_mirror(void *pkt, struct metadata *md, int ifc);
static int pkt_drop(void *pkt, struct metadata *md);


//BPF_TABLE("extern", int, struct metadata, metadata, NUMCPUS);
BPF_TABLE("extern", int, int, modules, MAX_MODULES);
BPF_TABLE("array", int, struct chain, forward_chain, MAX_INTERFACES);

static int forward(struct __sk_buff *skb, int out_ifc) {
	struct chain *cur = (struct chain *)skb->cb;
	struct chain *next = forward_chain.lookup(&out_ifc);
	if (next) {
		cur->hops[0] = chain_ifc(next, 0);
		cur->hops[1] = next->hops[1];
		cur->hops[2] = next->hops[2];
		cur->hops[3] = next->hops[3];
		//bpf_trace_printk("fwd:%d=0x%x %d\n", out_ifc, next->hops[0], chain_module(next, 0));
		modules.call(skb, chain_module(next, 0));
	}
	//bpf_trace_printk("fwd:%d=0\n", out_ifc);
	return TC_ACT_SHOT;
}

static int chain_pop(struct __sk_buff *skb) {
	struct chain *cur = (struct chain *)skb->cb;
	struct chain orig = *cur;
	cur->hops[0] = chain_ifc(&orig, 1);
	cur->hops[1] = cur->hops[2];
	cur->hops[2] = cur->hops[3];
	cur->hops[3] = 0;
	if (cur->hops[0]) {
		modules.call(skb, chain_module(&orig, 1));
	}

	//bpf_trace_printk("pop empty\n");
	return TC_ACT_OK;
}

int handle_rx_wrapper(struct __sk_buff *skb) {
	//bpf_trace_printk("" MODULE_UUID_SHORT ": rx:%d\n", skb->cb[0]);
	struct metadata md = {};
	md.in_ifc = skb->cb[0];

	int rc = handle_rx(skb, &md);

	// TODO: implementation
	switch (rc) {
		case RX_OK:
			return chain_pop(skb);
		case RX_REDIRECT:
			return forward(skb, md.redir_ifc);
		//case RX_RECIRCULATE:
		//	modules.call(skb, 1);
		//	break;
		case RX_DROP:
			return TC_ACT_SHOT;
	}
	//return rc;
	return TC_ACT_SHOT;
}

static int pkt_redirect(void *pkt, struct metadata *md, int ifc) {
	md->redir_ifc = ifc;
	return TC_ACT_OK;
}

static int pkt_mirror(void *pkt, struct metadata *md, int ifc) {
	md->clone_ifc = ifc;
	return TC_ACT_OK;
}


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
