#!/bin/bash
set -x

#clean ovs database
ovs-vsctl --no-wait del-port br-int veth1_
ovs-vsctl --no-wait del-port br-int veth2_
ovs-vsctl --no-wait del-port br-int veth3_
ovs-vsctl --no-wait del-port br-int veth4_
ovs-vsctl --no-wait del-port br-int veth5_

#clean ovn databse
ovn-nbctl lsp-del sw0-port1
ovn-nbctl lsp-del sw0-port2
ovn-nbctl lsp-del sw0-port3
ovn-nbctl lsp-del sw0-port4
ovn-nbctl lsp-del sw0-port5
ovn-nbctl ls-del sw0

#delete links
sudo ip link del veth1
sudo ip link del veth2
sudo ip link del veth3
sudo ip link del veth4
sudo ip link del veth5

sudo ip link del veth1_
sudo ip link del veth2_
sudo ip link del veth3_
sudo ip link del veth4_
sudo ip link del veth5_

#delete namespaces
sudo ip netns del ns1
sudo ip netns del ns2
sudo ip netns del ns3
sudo ip netns del ns4
sudo ip netns del ns5
