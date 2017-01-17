#! /bin/bash

set -x

for i in `seq 1 2`;
do
	# remove ns and veth pairs if already created
	sudo ip netns del ns${i}
	sudo ip link del veth${i}
done

for i in `seq 1 2`;
do
	# create ns and veth pairs
	sudo ip netns add ns${i}
	sudo ip link add veth${i}_ type veth peer name veth${i}
	sudo ip link set veth${i}_ netns ns${i}
	sudo ip netns exec ns${i} ip link set dev veth${i}_ up
	sudo ip link set dev veth${i} up
	sudo ip netns exec ns${i} ifconfig veth${i}_ 10.10.1.${i}/24

  sudo ethtool --offload veth${i} rx off tx off
  sudo ethtool -K veth${i} gso off
done

mac1=$(sudo ip netns exec ns1 ifconfig | grep veth1 | awk '{print $5}')
mac2=$(sudo ip netns exec ns2 ifconfig | grep veth2 | awk '{print $5}')

sudo ip netns exec ns1 sudo arp -s 10.10.1.2 $mac2
