package bpf

var DummySwitch2 = `
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
`

var DummySwitch2count = `
struct mac_key {
	u64 mac;
};

struct host_info {
	u32 ifindex;
	u64 rx_pkts;
	u64 tx_pkts;
};

BPF_TABLE("hash", struct mac_key, struct host_info, mac2host, 10240);

BPF_TABLE("array",u32,u32,count,10);

static int handle_rx(void *skb, struct metadata *md) {
	u32 key = 0, value = 0, *cnt;

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
			key = 1;cnt = count.lookup_or_init(&key,&value);(*cnt)++;
			return RX_REDIRECT;
		}
		if(md->in_ifc==2){
			pkt_redirect(skb,md,1);
			key = 2;cnt = count.lookup_or_init(&key,&value);(*cnt)++;
			return RX_REDIRECT;
		}
		return RX_DROP;
	}
}
`

var DummySwitch3 = `
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
	//debug init for broadcast_count
	u32 key = 0, value = 0, *cnt;
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
		//increment tx_pkts counter
		lock_xadd(&dst_host->tx_pkts, 1);

		//redirect packet to correspondent interface
		pkt_redirect(skb, md, dst_host->ifindex);

		return RX_REDIRECT;

	} else {
		//miss in forwarding table
		//broadcast test implementation
			if(md->in_ifc==1){
				bpf_clone_redirect(skb, 2, 0);
				bpf_clone_redirect(skb, 3, 0);
				return RX_OK;
			}
			if(md->in_ifc==2){
				bpf_clone_redirect(skb, 1, 0);
				bpf_clone_redirect(skb, 3, 0);
				return RX_OK;
			}
			if(md->in_ifc==3){
				bpf_clone_redirect(skb, 1, 0);
				bpf_clone_redirect(skb, 2, 0);
				return RX_OK;
			}
		return RX_DROP;
	}
}
`