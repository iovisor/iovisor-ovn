
//****IOMODULES WRAPPER********

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

/*******CODE OF MY IOMODULE**********/

#define MAX_PORTS 8

struct mac_key {
	u64 mac;
};

struct host_info {
	u32 ifindex;
	u64 rx_pkts;
	u64 tx_pkts;
};

BPF_TABLE("hash", struct mac_key, struct host_info, mac2host, 10240);

BPF_TABLE("array",u32,u32,ports,MAX_PORTS);

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

	bpf_trace_printk("pkt in_ifc:%d from:%x -> to:%x\n",md->in_ifc,ethernet->src,ethernet->dst);

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
		//HIT in forwarding table
		//increment tx_pkts counter
		lock_xadd(&dst_host->tx_pkts, 1);

		//redirect packet to correspondent interface
		pkt_redirect(skb, md, dst_host->ifindex);

		bpf_trace_printk("lookup hit -> redirect to %d\n",dst_host->ifindex);

		return RX_REDIRECT;

	} else {
		//MISS in forwarding table
		bpf_trace_printk("lookup miss->  broadcast\n");

		u32  iface_n = 1;
		u32 *iface_p;

		//Manual unroll to avoid ebpf verifier error due to loops
		//#pragma unroll seems not to work

		iface_n = 1; iface_p = ports.lookup(&iface_n);
		if(iface_p)
			if(*iface_p != 0)
				bpf_clone_redirect(skb,*iface_p,0);

		iface_n = 2; iface_p = ports.lookup(&iface_n);
		if(iface_p)
			if(*iface_p != 0)
				bpf_clone_redirect(skb,*iface_p,0);

		iface_n = 3; iface_p = ports.lookup(&iface_n);
		if(iface_p)
			if(*iface_p != 0)
				bpf_clone_redirect(skb,*iface_p,0);

		iface_n = 4; iface_p = ports.lookup(&iface_n);
		if(iface_p)
			if(*iface_p != 0)
				bpf_clone_redirect(skb,*iface_p,0);

		iface_n = 5; iface_p = ports.lookup(&iface_n);
		if(iface_p)
			if(*iface_p != 0)
				bpf_clone_redirect(skb,*iface_p,0);

		iface_n = 6; iface_p = ports.lookup(&iface_n);
		if(iface_p)
			if(*iface_p != 0)
				bpf_clone_redirect(skb,*iface_p,0);

		iface_n = 7; iface_p = ports.lookup(&iface_n);
		if(iface_p)
			if(*iface_p != 0)
				bpf_clone_redirect(skb,*iface_p,0);

		iface_n = 8; iface_p = ports.lookup(&iface_n);
		if(iface_p)
			if(*iface_p != 0)
				bpf_clone_redirect(skb,*iface_p,0);

		return RX_DROP;
	}
}
