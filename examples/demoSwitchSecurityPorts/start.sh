#!/bin/bash
set -x

#create 5 ns
sudo ip netns add ns1
sudo ip netns add ns2
sudo ip netns add ns3
sudo ip netns add ns4
sudo ip netns add ns5

#create links
sudo ip link add veth1 type veth peer name veth1_
sudo ip link set veth1 netns ns1
sudo ip link add veth2 type veth peer name veth2_
sudo ip link set veth2 netns ns2
sudo ip link add veth3 type veth peer name veth3_
sudo ip link set veth3 netns ns3
sudo ip link add veth4 type veth peer name veth4_
sudo ip link set veth4 netns ns4
sudo ip link add veth5 type veth peer name veth5_
sudo ip link set veth5 netns ns5
sudo ip netns exec ns1 ip link set dev veth1 up
sudo ip link set dev veth1_ up
sudo ip netns exec ns1 ifconfig veth1 10.10.1.1/24
sudo ip netns exec ns2 ip link set dev veth2 up
sudo ip link set dev veth2_ up
sudo ip netns exec ns2 ifconfig veth2 10.10.1.2/24
sudo ip netns exec ns3 ip link set dev veth3 up
sudo ip link set dev veth3_ up
sudo ip netns exec ns3 ifconfig veth3 10.10.1.3/24
sudo ip netns exec ns4 ip link set dev veth4 up
sudo ip link set dev veth4_ up
sudo ip netns exec ns4 ifconfig veth4 10.10.1.4/24
sudo ip netns exec ns5 ip link set dev veth5 up
sudo ip link set dev veth5_ up
sudo ip netns exec ns5 ifconfig veth5 10.10.1.5/24

go install github.com/iovisor/iomodules/hover/hoverd
gnome-terminal --command="sudo $GOPATH/bin/hoverd -listen 127.0.0.1:5002"

sleep 1

go install github.com/netgroup-polito/iovisor-ovn/daemon
gnome-terminal --command="sudo $GOPATH/bin/daemon -hover http://127.0.0.1:5002"

echo ""
echo "PRESS ANY KEY TO LAUNCH OVN COMMANDS ..."
echo ""

read -r line

#ns1    NO POLICY
#ns2    NO POLICY
#ns3    POLICY ON MAC WRONG mac
#ns4    POLICY ON MAC CORRECT mac
#ns5    POLICY ON MAC CORRECT mac

# mac3=$(sudo ip netns exec ns3 ifconfig | grep veth3 | awk '{print $5}')
mac4=$(sudo ip netns exec ns4 ifconfig | grep veth4 | awk '{print $5}')
mac5=$(sudo ip netns exec ns5 ifconfig | grep veth5 | awk '{print $5}')

# echo "$mac3 $mac4 $mac5"

#ovn
# Create a logical switch named "sw0"
ovn-nbctl ls-add sw0

#Create logical ports on switch 0
#read -r line
ovn-nbctl lsp-add sw0 sw0-port1

#read -r line
ovn-nbctl lsp-add sw0 sw0-port2

#read -r line
ovn-nbctl lsp-add sw0 sw0-port3

#read -r line
ovn-nbctl lsp-add sw0 sw0-port4

#read -r line
ovn-nbctl lsp-add sw0 sw0-port5

#read -r line
# Set a MAC address (no security policy) and Ip for ports
# ovn-nbctl lsp-set-addresses sw0-port1 "00:00:00:00:00:01 192.168.1.1"

#read -r line
# Set a MAC address (no security policy) and Ip for ports
# ovn-nbctl lsp-set-addresses sw0-port2 "00:00:00:00:00:02 192.168.1.2"

#read -r line
# Set a MAC address for each of the two logical ports.
#wrong mac, wrong address
ovn-nbctl lsp-set-port-security sw0-port3 "01:02:03:04:05:06 10.2.2.2"

#read -r line
ovn-nbctl lsp-set-port-security sw0-port4 "$mac4 10.10.1.4"

#read -r line
ovn-nbctl lsp-set-port-security sw0-port5 "$mac5 10.10.1.5"

# Create ports on the local OVS bridge, br-int.  When ovn-controller
# sees these ports show up with an "iface-id" that matches the OVN
# logical port names, it associates these local ports with the OVN
# logical ports.  ovn-controller will then set up the flows necessary
# for these ports to be able to communicate each other as defined by
# the OVN logical topology.
# read -r line
ovs-vsctl --no-wait add-port br-int veth1_ -- set Interface veth1_ external_ids:iface-id=sw0-port1

# read -r line
ovs-vsctl --no-wait add-port br-int veth2_ -- set Interface veth2_ external_ids:iface-id=sw0-port2

# read -r line
ovs-vsctl --no-wait add-port br-int veth3_ -- set Interface veth3_ external_ids:iface-id=sw0-port3

# read -r line
ovs-vsctl --no-wait add-port br-int veth4_ -- set Interface veth4_ external_ids:iface-id=sw0-port4

# read -r line
ovs-vsctl --no-wait add-port br-int veth5_ -- set Interface veth5_ external_ids:iface-id=sw0-port5



gnome-terminal --command="sudo cat /sys/kernel/debug/tracing/trace_pipe"


echo "WAITING FOR CONFIGURATION READY...."

sleep 25

#ns1 -> 2,3,4,5
echo "sudo ip netns exec ns1 ping 10.10.1.2 -c 2"
sudo ip netns exec ns1 ping 10.10.1.2 -c 2
echo "sudo ip netns exec ns1 ping 10.10.1.3 -c 2"
sudo ip netns exec ns1 ping 10.10.1.3 -c 2
echo "sudo ip netns exec ns1 ping 10.10.1.4 -c 2"
sudo ip netns exec ns1 ping 10.10.1.4 -c 2
echo "sudo ip netns exec ns1 ping 10.10.1.5 -c 2"
sudo ip netns exec ns1 ping 10.10.1.5 -c 2

#ns2->1,3,4,5
echo "sudo ip netns exec ns2 ping 10.10.1.1 -c 2"
sudo ip netns exec ns2 ping 10.10.1.1 -c 2
echo "sudo ip netns exec ns2 ping 10.10.1.3 -c 2"
sudo ip netns exec ns2 ping 10.10.1.3 -c 2
echo "sudo ip netns exec ns2 ping 10.10.1.4 -c 2"
sudo ip netns exec ns2 ping 10.10.1.4 -c 2
echo "sudo ip netns exec ns2 ping 10.10.1.5 -c 2"
sudo ip netns exec ns2 ping 10.10.1.5 -c 2

#ns3->1,2,4,5
echo "sudo ip netns exec ns3 ping 10.10.1.1 -c 2"
sudo ip netns exec ns3 ping 10.10.1.1 -c 2
echo "sudo ip netns exec ns3 ping 10.10.1.2 -c 2"
sudo ip netns exec ns3 ping 10.10.1.2 -c 2
echo "sudo ip netns exec ns3 ping 10.10.1.4 -c 2"
sudo ip netns exec ns3 ping 10.10.1.4 -c 2
echo "sudo ip netns exec ns3 ping 10.10.1.5 -c 2"
sudo ip netns exec ns3 ping 10.10.1.5 -c 2

#ns4->1,2,3,5
echo "sudo ip netns exec ns4 ping 10.10.1.1 -c 2"
sudo ip netns exec ns4 ping 10.10.1.1 -c 2
echo "sudo ip netns exec ns4 ping 10.10.1.2 -c 2"
sudo ip netns exec ns4 ping 10.10.1.2 -c 2
echo "sudo ip netns exec ns4 ping 10.10.1.3 -c 2"
sudo ip netns exec ns4 ping 10.10.1.3 -c 2
echo "sudo ip netns exec ns4 ping 10.10.1.5 -c 2"
sudo ip netns exec ns4 ping 10.10.1.5 -c 2

#ns5->1,2,3,4
echo "sudo ip netns exec ns5 ping 10.10.1.1 -c 2"
sudo ip netns exec ns5 ping 10.10.1.1 -c 2
echo "sudo ip netns exec ns5 ping 10.10.1.2 -c 2"
sudo ip netns exec ns5 ping 10.10.1.2 -c 2
echo "sudo ip netns exec ns5 ping 10.10.1.3 -c 2"
sudo ip netns exec ns5 ping 10.10.1.3 -c 2
echo "sudo ip netns exec ns5 ping 10.10.1.4 -c 2"
sudo ip netns exec ns5 ping 10.10.1.4 -c 2
