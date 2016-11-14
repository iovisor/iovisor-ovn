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

BPF_TABLE("hash", struct mac_t, struct interface, fwdtable, 10240);
BPF_TABLE("array",u32,u32,ports,MAX_PORTS);
BPF_TABLE("hash",struct ifindex,struct mac_t, securitymac, MAX_PORTS*2);
BPF_TABLE("hash",struct ifindex,struct ip_leaf, securityip, MAX_PORTS*2);

static int handle_rx(void *skb, struct metadata *md) {
	u8 *cursor = 0;
	struct ethernet_t *ethernet = cursor_advance(cursor, sizeof(*ethernet));

	#ifdef BPF_TRACE
		bpf_trace_printk("in_ifc=%d\n",md->in_ifc);
	#endif

	//set in-interface for lookup ports security
	struct ifindex in_iface = {};
	in_iface.ifindex = md->in_ifc;

	//port security on source mac
	struct mac_t * mac_lookup;
	mac_lookup = securitymac.lookup(&in_iface);
	if (mac_lookup)
		if (ethernet->src != mac_lookup->mac){
			#ifdef BPF_TRACE
				bpf_trace_printk("mac %lx mismatch %lx -> DROP\n",ethernet->src, mac_lookup->mac);
			#endif
			return RX_DROP;
		}

	//port security on source ip
	if(ethernet->type == 0x0800){
		struct ip_leaf * ip_lookup;
		ip_lookup = securityip.lookup(&in_iface);
		if (ip_lookup){
			struct ip_t *ip = cursor_advance(cursor,sizeof(*ip));
			if (ip->src != ip_lookup->ip){
				#ifdef BPF_TRACE
					bpf_trace_printk("IP %x mismatch %x -> DROP\n",ip->src, ip_lookup->ip);
				#endif
				return RX_DROP;
			}
		}
	}

	#ifdef BPF_TRACE
		bpf_trace_printk("mac src:%lx dst:%lx\n",ethernet->src,ethernet->dst);
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
	if(interface_lookup->ifindex!=md->in_ifc)
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
			bpf_trace_printk("redirect out_ifc=%d\n",dst_interface->ifindex);
		#endif

		return RX_REDIRECT;

	} else {
		//MISS in forwarding table
		#ifdef BPF_TRACE
			bpf_trace_printk("broadcast\n");
		#endif

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


		iface_n = 9; iface_p = ports.lookup(&iface_n);
		if(iface_p)
			if(*iface_p != 0)
				bpf_clone_redirect(skb,*iface_p,0);


		iface_n = 10; iface_p = ports.lookup(&iface_n);
		if(iface_p)
			if(*iface_p != 0)
				bpf_clone_redirect(skb,*iface_p,0);


		iface_n = 11; iface_p = ports.lookup(&iface_n);
		if(iface_p)
			if(*iface_p != 0)
				bpf_clone_redirect(skb,*iface_p,0);

		iface_n = 12; iface_p = ports.lookup(&iface_n);
		if(iface_p)
			if(*iface_p != 0)
				bpf_clone_redirect(skb,*iface_p,0);


		iface_n = 13; iface_p = ports.lookup(&iface_n);
		if(iface_p)
			if(*iface_p != 0)
				bpf_clone_redirect(skb,*iface_p,0);


		iface_n = 14; iface_p = ports.lookup(&iface_n);
		if(iface_p)
			if(*iface_p != 0)
				bpf_clone_redirect(skb,*iface_p,0);


		iface_n = 15; iface_p = ports.lookup(&iface_n);
		if(iface_p)
			if(*iface_p != 0)
				bpf_clone_redirect(skb,*iface_p,0);


		iface_n = 16; iface_p = ports.lookup(&iface_n);
		if(iface_p)
			if(*iface_p != 0)
				bpf_clone_redirect(skb,*iface_p,0);

		iface_n = 17; iface_p = ports.lookup(&iface_n);
		if(iface_p)
			if(*iface_p != 0)
				bpf_clone_redirect(skb,*iface_p,0);


		iface_n = 18; iface_p = ports.lookup(&iface_n);
		if(iface_p)
			if(*iface_p != 0)
				bpf_clone_redirect(skb,*iface_p,0);


		iface_n = 19; iface_p = ports.lookup(&iface_n);
		if(iface_p)
			if(*iface_p != 0)
				bpf_clone_redirect(skb,*iface_p,0);


		iface_n = 20; iface_p = ports.lookup(&iface_n);
		if(iface_p)
			if(*iface_p != 0)
				bpf_clone_redirect(skb,*iface_p,0);


		iface_n = 21; iface_p = ports.lookup(&iface_n);
		if(iface_p)
			if(*iface_p != 0)
				bpf_clone_redirect(skb,*iface_p,0);


		iface_n = 22; iface_p = ports.lookup(&iface_n);
		if(iface_p)
			if(*iface_p != 0)
				bpf_clone_redirect(skb,*iface_p,0);


		iface_n = 23; iface_p = ports.lookup(&iface_n);
		if(iface_p)
			if(*iface_p != 0)
				bpf_clone_redirect(skb,*iface_p,0);


		iface_n = 24; iface_p = ports.lookup(&iface_n);
		if(iface_p)
			if(*iface_p != 0)
				bpf_clone_redirect(skb,*iface_p,0);

		iface_n = 25; iface_p = ports.lookup(&iface_n);
		if(iface_p)
			if(*iface_p != 0)
				bpf_clone_redirect(skb,*iface_p,0);


		iface_n = 26; iface_p = ports.lookup(&iface_n);
		if(iface_p)
			if(*iface_p != 0)
				bpf_clone_redirect(skb,*iface_p,0);


		iface_n = 27; iface_p = ports.lookup(&iface_n);
		if(iface_p)
			if(*iface_p != 0)
				bpf_clone_redirect(skb,*iface_p,0);


		iface_n = 28; iface_p = ports.lookup(&iface_n);
		if(iface_p)
			if(*iface_p != 0)
				bpf_clone_redirect(skb,*iface_p,0);

		iface_n = 29; iface_p = ports.lookup(&iface_n);
		if(iface_p)
			if(*iface_p != 0)
				bpf_clone_redirect(skb,*iface_p,0);

		iface_n = 30; iface_p = ports.lookup(&iface_n);
		if(iface_p)
			if(*iface_p != 0)
				bpf_clone_redirect(skb,*iface_p,0);

		iface_n = 31; iface_p = ports.lookup(&iface_n);
		if(iface_p)
			if(*iface_p != 0)
				bpf_clone_redirect(skb,*iface_p,0);

		iface_n = 32; iface_p = ports.lookup(&iface_n);
		if(iface_p)
			if(*iface_p != 0)
				bpf_clone_redirect(skb,*iface_p,0);

		return RX_DROP;
	}
}
`
