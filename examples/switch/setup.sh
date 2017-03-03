#! /bin/bash

set -x
set -e

for i in `seq 1 2`;
do
	# remove ns and veth pairs if already created
	if [ -e /var/run/netns/ns${i} ]; then
		sudo ip netns del ns${i}
		sudo ip link del veth${i}
	fi

	sudo ip netns add ns${i}
	sudo ip link add veth${i}_ type veth peer name veth${i}
	sudo ip link set veth${i}_ netns ns${i}
	sudo ip netns exec ns${i} ip link set dev veth${i}_ up
	sudo ip link set dev veth${i} up
	sudo ip netns exec ns${i} ifconfig veth${i}_ 10.0.0.${i}/24
done
