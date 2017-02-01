#!/bin/bash

set -x

sudo ip netns exec ns1 sudo arping -c 1 10.0.1.254
sudo ip netns exec ns2 sudo arping -c 1 10.0.2.254
sudo ip netns exec ns3 sudo arping -c 1 10.10.1.100

sudo ip netns exec ns1 sudo ping -c 1 10.0.1.254
sudo ip netns exec ns2 sudo ping -c 1 10.0.2.254
sudo ip netns exec ns3 sudo ping -c 1 10.10.1.100
