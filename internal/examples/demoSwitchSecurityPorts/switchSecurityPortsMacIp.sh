#!/bin/bash
#set -x

# Clear previous congigurations
# Setup Ns
#
# Create a Switch
# Connect 4 ports with Security Ports on MAC and IP
# Test Ping works
#
#Change mac/IP in namespaces 3-4 -> mac and IP not matching Rules anymore
n=4

echo "cleanup previous namespaces"
for i in `seq 1 $n`;
do
  sudo ip netns del ns${i}
done

echo "cleanup previous OVN Switches"
for i in `seq 1 $n`;
do
  ovs-vsctl --no-wait del-port br-int veth${i}_
  ovn-nbctl lsp-del sw0-port${i}
done

ovn-nbctl ls-del sw0

sleep 1

echo "PRESS A KEY TO START DEMO CONFIG....."
read -r line

echo "setting up namespaces and veth"


mac="a"
for i in `seq 1 $n`;
do
  echo "add ns${i} -> veth${i}_  "
  sudo ip netns add ns${i}
  sudo ip link add veth${i} type veth peer name veth${i}_
  sudo ip link set veth${i} netns ns${i}
  sudo ip netns exec ns${i} ip link set dev veth${i} up
  sudo ip link set dev veth${i}_ up
  sudo ip netns exec ns${i} ifconfig veth${i} 10.10.1.${i}/24
  new_mac=$(sudo ip netns exec ns${i} ifconfig | grep veth${i} | awk '{print $5}')
  mac=(${mac[@]} $new_mac)
done

echo "ovn-nbctl ls-add sw0"
ovn-nbctl ls-add sw0

for i in `seq 1 $n`;
do
  echo "ovn-nbctl lsp-add sw0 sw0-port${i}"
  ovn-nbctl lsp-add sw0 sw0-port${i}
  echo ovn-nbctl lsp-set-port-security sw0-port${i} "${mac[$i]} 10.10.1.${i}"
  ovn-nbctl lsp-set-port-security sw0-port${i} "${mac[$i]} 10.10.1.${i}"
  echo ovs-vsctl --no-wait add-port br-int veth${i}_ -- set Interface veth${i}_ external_ids:iface-id=sw0-port${i}
  ovs-vsctl --no-wait add-port br-int veth${i}_ -- set Interface veth${i}_ external_ids:iface-id=sw0-port${i}
done

echo ""
echo "**CURRENT CONFIGURATION**"
echo "Switch 4 Ports"
echo "Port Security on MAC and IP"
echo ""
echo "In current configuration Ping Should Work for each interface"
echo ""
echo "PRESS A KEY TO START PING....."
read -r line

sudo ip netns exec ns2 ping 10.10.1.1 -c 1

for i in `seq 2 $n`;
do
  sudo ip netns exec ns1 ping 10.10.1.${i} -c 1
done

echo ""
echo "Ping END"
echo ""
echo ""
echo "PRESS A KEY TO CHANGE CONFIG (change some ip address inside ns)....."
read -r line

echo ""
echo "sudo ip netns exec ns3 ifconfig veth3 10.10.1.103/24"
sudo ip netns exec ns3 ifconfig veth3 10.10.1.103/24
echo "sudo ip netns exec ns4 ifconfig veth4 10.10.1.104/24"
sudo ip netns exec ns4 ifconfig veth4 10.10.1.104/24


echo ""
echo "**CURRENT CONFIGURATION**"
echo "Switch 4 Ports"
echo "Port Security on MAC and IP"
echo "Port3 -> WRONG IP -> Switch will DROP Packets"
echo "Port4 -> WRONG IP -> Switch will DROP Packets"
echo ""
echo "In current configuration Ping from/to ns3-4 should NOT work!"
echo ""
echo "PRESS A KEY TO START PING....."
read -r line

echo "PING from: ns1 --> to: ns2 (10.10.1.2)  OK!"
echo ""
sudo ip netns exec ns1 ping 10.10.1.2 -c 1

echo "PING from: ns2 --> to: ns1 (10.10.1.1)  OK!"
echo ""
sudo ip netns exec ns2 ping 10.10.1.1 -c 1

echo "PING from: ns3 --> to: ns1 (10.10.1.1)  DROP"
echo ""
sudo ip netns exec ns3 ping 10.10.1.1 -c 1

echo "PING from: ns4 --> to: ns1 (10.10.1.1)  DROP"
echo ""
sudo ip netns exec ns4 ping 10.10.1.1 -c 1

echo ""
echo "Ping END"
echo ""
echo ""
echo "PRESS A KEY TO CLEANUP....."
read -r line
echo "OVN Cleanup"

for i in `seq 1 $n`;
do
  ovs-vsctl --no-wait del-port br-int veth${i}_
  sleep 0.5
  ovn-nbctl lsp-del sw0-port${i}
  sleep 0.5
done

ovn-nbctl ls-del sw0

sleep 2

for i in `seq 1 $n`;
do
  sudo ip netns del ns${i}
done
