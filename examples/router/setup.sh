#! /bin/bash

set -x
set -e

for i in `seq 1 3`;
do
	# remove ns and veth pairs if already created
	if [ -e /var/run/netns/ns${i} ]; then
		sudo ip netns del ns${i}
		sudo ip link del veth${i}
	fi

	# create ns and veth pairs
	sudo ip netns add ns${i}
	sudo ip link add veth${i}_ type veth peer name veth${i}
	sudo ip link set veth${i}_ netns ns${i}
	sudo ip netns exec ns${i} ip link set dev veth${i}_ up
	sudo ip link set dev veth${i} address 82:73:8d:f3:62:${i} up
	sudo ip netns exec ns${i} ifconfig veth${i}_ 10.0.${i}.100/24

	sudo ip netns exec ns${i} route add default gw 10.0.${i}.1 veth${i}_
done
