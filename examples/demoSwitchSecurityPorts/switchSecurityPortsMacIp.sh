#!/bin/bash
#set -x

# Clear previous congigurations
# Setup Ns
#
# Create a Switch
# Connect 7 ports with Security Ports on MAC and IP
# Test Ping works
#
# Configure Ports Security
# p1 no policy              OK
# p2 mac+ip correct policy  OK
# p3 mac correct            OK
# p4 ip correct             OK
# p5 mac wrong              NO
# p6 ip wrong               NO
# p7 MAC + IP wrong         NO

n=7

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

echo "adding logical-switch"
ovn-nbctl ls-add sw0

for i in `seq 1 $n`;
do
  echo "adding logical-switch sw0 port sw0-port${i}"
  ovn-nbctl lsp-add sw0 sw0-port${i}
  ovn-nbctl lsp-set-port-security sw0-port${i} "${mac[$i]} 10.10.1.${i}"
  ovs-vsctl --no-wait add-port br-int veth${i}_ -- set Interface veth${i}_ external_ids:iface-id=sw0-port${i}
done

echo "**SWITCH 7 ports with Ports security on MAC and IP**"
echo "PRESS A KEY TO START PING....."
read -r line

# sudo ip netns exec ns2 ping 10.10.1.1 -c 1

for i in `seq 2 $n`;
do
  sudo ip netns exec ns1 ping 10.10.1.${i} -c 1
done

echo "PRESS A KEY TO CHANGE CONFIG....."

echo "**SWITCH 7 ports with Ports security on MAC and IP**"
echo "Port1 No Policy->OK"
echo "Port2 Policy MAC +IP Correct->OK"
echo "Port3 MAC Correct->OK"
echo "Port4 IP Correct->OK"
echo "Port5 MAC Wrong->KO!"
echo "Port6 IP Wrong->KO!"
echo "Port7 MAC+IP Wrong->KO!"

read -r line

ovn-nbctl lsp-set-port-security sw0-port1 ""
ovn-nbctl lsp-set-port-security sw0-port2 "${mac[2]} 10.10.1.2"
ovn-nbctl lsp-set-port-security sw0-port3 "${mac[3]} "
ovn-nbctl lsp-set-port-security sw0-port4 "10.10.1.4"
ovn-nbctl lsp-set-port-security sw0-port5 "01:00:02:00:03:00"
ovn-nbctl lsp-set-port-security sw0-port6 "10.1.2.33"
ovn-nbctl lsp-set-port-security sw0-port7 "00:11:00:22:00:33 10.1.2.77"


sudo ip netns exec ns2 ping 10.10.1.1 -c 1

for i in `seq 2 $n`;
do
  sudo ip netns exec ns1 ping 10.10.1.${i} -c 1
done

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
