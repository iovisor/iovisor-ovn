#! /bin/bash

set -x

for i in `seq 1 3`;
do
	# remove ns and veth pairs if already created
	sudo ip netns del ns${i}
	sudo ip link del veth${i}

	# create ns and veth pairs
	sudo ip netns add ns${i}
	sudo ip link add veth${i}_ type veth peer name veth${i}
	sudo ip link set veth${i}_ netns ns${i}
	sudo ip netns exec ns${i} ip link set dev veth${i}_ up
	sudo ip link set dev veth${i} up
done
