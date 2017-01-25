#! /bin/bash

set -x

for i in `seq 1 3`;
do
	# remove ns and veth pairs if already created
	sudo ip netns del ns${i}
	sudo ip link del veth${i}
done

for i in `seq 1 3`;
do
	# create ns and veth pairs
	sudo ip netns add ns${i}
	sudo ip link add veth${i}_ type veth peer name veth${i}
	sudo ip link set veth${i}_ netns ns${i}
	sudo ip netns exec ns${i} ip link set dev veth${i}_ up
	sudo ip link set dev veth${i} up
	sudo ip netns exec ns${i} ifconfig veth${i}_ 10.0.${i}.1/24

  sudo ethtool --offload veth${i} rx off tx off
  sudo ethtool -K veth${i} gso off
done

sudo ip netns exec ns3 ifconfig veth3_ 10.10.1.1/24

sudo ip netns exec ns1 sudo route add default gw 10.0.1.254 veth1_
sudo ip netns exec ns2 sudo route add default gw 10.0.2.254 veth2_
sudo ip netns exec ns3 sudo route add default gw 10.10.1.100 veth3_
