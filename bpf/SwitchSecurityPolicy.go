package bpf

var SwitchSecurityPolicy = `
#define MAX_PORTS 8

struct mac_key {
	u64 mac;
};

struct host_info {
	u32 ifindex;
	u64 rx_pkts;
	u64 tx_pkts;
};

struct iface_key{
	u32 ifindex;
};

struct mac_leaf{
	u64 mac;
};

struct ip_leaf{
	u32 ip;
};

BPF_TABLE("hash", struct mac_key, struct host_info, mac2host, 10240);

BPF_TABLE("array",u32,u32,ports,MAX_PORTS);

BPF_TABLE("hash",struct iface_key,struct mac_leaf, securitymac, MAX_PORTS);

BPF_TABLE("hash",struct iface_key,struct ip_leaf, securityip, MAX_PORTS);


static int handle_rx(void *skb, struct metadata *md) {

	u8 *cursor = 0;

	struct ethernet_t *ethernet = cursor_advance(cursor, sizeof(*ethernet));

	bpf_trace_printk("md->in_ifc = %d\n",md->in_ifc);

	struct iface_key ifc_key = {};
	ifc_key.ifindex = md->in_ifc;

	struct mac_leaf * mac_lookup;

	bpf_trace_printk("Policy on MAC\n");
	mac_lookup = securitymac.lookup(&ifc_key);
	if (mac_lookup)
		if (ethernet->src != mac_lookup->mac){
			bpf_trace_printk("MAC %x mismatch %x\n",ethernet->src, mac_lookup->mac);
			return RX_DROP;
		}

	bpf_trace_printk("Policy on IP\n");
	bpf_trace_printk("Ethertype %x\n",ethernet->type);
	//IP Packets
	if(ethernet->type == 0x0800){
		bpf_trace_printk("IP Packet\n");

		struct ip_t *ip = cursor_advance(cursor,sizeof(*ip));

		struct ip_leaf * ip_lookup;

		ip_lookup = securityip.lookup(&ifc_key);
		if (ip_lookup){
			bpf_trace_printk("IP Lookup hit\n");
			if (ip->src != ip_lookup->ip){
				bpf_trace_printk("IP %x mismatch %x\n",ip->src, ip_lookup->ip);
				return RX_DROP;
			}
		}else{
			bpf_trace_printk("IP Lookup miss\n");
		}
	}
	else{
			bpf_trace_printk("NO IP Packet\n");
	}

	//LEARNING PHASE: mapping src_mac with src_interface
	struct mac_key src_key = {};
	struct host_info src_info = {};

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

		// iface_n = 7; iface_p = ports.lookup(&iface_n);
		// if(iface_p)
		// 	if(*iface_p != 0)
		// 		bpf_clone_redirect(skb,*iface_p,0);
		//
		// iface_n = 8; iface_p = ports.lookup(&iface_n);
		// if(iface_p)
		// 	if(*iface_p != 0)
		// 		bpf_clone_redirect(skb,*iface_p,0);

		return RX_DROP;
	}
}
`
