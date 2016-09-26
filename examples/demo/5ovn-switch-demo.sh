#!/bin/bash
set -x
#launch hover
#go install github.com/mbertrone/iomodules/hover/hoverd
#gnome-terminal --command="sudo $GOPATH/bin/hoverd -listen 127.0.0.1:5002"

#launch iovisor-ovn daemon (&monitors)
#go install github.com/netgroup-polito/iovisor-ovn/daemon
#gnome-terminal --command="sudo $GOPATH/bin/daemon -hover http://127.0.0.1:5002"


#ovn

# Create a logical switch named "sw0"

ovn-nbctl ls-add sw0

read -r line
# Create two logical ports on "sw0".
ovn-nbctl lsp-add sw0 sw0-port1

read -r line
ovn-nbctl lsp-add sw0 sw0-port2

read -r line
# Create two logical ports on "sw0".
ovn-nbctl lsp-add sw0 sw0-port3

read -r line
ovn-nbctl lsp-add sw0 sw0-port4

read -r line
# Create two logical ports on "sw0".
ovn-nbctl lsp-add sw0 sw0-port5


read -r line
# Set a MAC address for each of the two logical ports.
ovn-nbctl lsp-set-addresses sw0-port1 00:00:00:00:00:01

read -r line
ovn-nbctl lsp-set-addresses sw0-port2 00:00:00:00:00:02

# Set up port security for the two logical ports.  This ensures that
# the logical port mac address we have configured is the only allowed
# source and destination mac address for these ports.
#ovn-nbctl lsp-set-port-security sw0-port1 00:00:00:00:00:01
#ovn-nbctl lsp-set-port-security sw0-port2 00:00:00:00:00:02

# Create ports on the local OVS bridge, br-int.  When ovn-controller
# sees these ports show up with an "iface-id" that matches the OVN
# logical port names, it associates these local ports with the OVN
# logical ports.  ovn-controller will then set up the flows necessary
# for these ports to be able to communicate each other as defined by
# the OVN logical topology.
read -r line
ovs-vsctl --no-wait add-port br-int veth1_ -- set Interface veth1_ external_ids:iface-id=sw0-port1

read -r line
ovs-vsctl --no-wait add-port br-int veth2_ -- set Interface veth2_ external_ids:iface-id=sw0-port2

read -r line
ovs-vsctl --no-wait add-port br-int veth3_ -- set Interface veth3_ external_ids:iface-id=sw0-port3

read -r line
ovs-vsctl --no-wait add-port br-int veth4_ -- set Interface veth4_ external_ids:iface-id=sw0-port4

read -r line
ovs-vsctl --no-wait add-port br-int veth5_ -- set Interface veth5_ external_ids:iface-id=sw0-port5
