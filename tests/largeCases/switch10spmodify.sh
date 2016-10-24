#!/bin/bash
#set -x

#test switch 10 ports and modify security policies

n=10

echo "cleanup previous namespaces"
for i in `seq 1 $n`;
do
  sudo ip netns del ns${i}
done

for i in `seq 1 $n`;
do
  ovs-vsctl --no-wait del-port br-int veth${i}_
  ovn-nbctl lsp-del sw0-port${i}
done
sleep 1

mac="a"

echo "PRESS A KEY TO START TEST CONFIG....."
read -r line

echo "setting up namespaces and veth"

for i in `seq 1 $n`;
do
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
  echo "adding logical-switch port sw0-port${i}"
  ovn-nbctl lsp-add sw0 sw0-port${i}
  ovn-nbctl lsp-set-port-security sw0-port${i} "${mac[$i]} 10.10.1.${i}"
  ovs-vsctl --no-wait add-port br-int veth${i}_ -- set Interface veth${i}_ external_ids:iface-id=sw0-port${i}
done

echo "PRESS A KEY TO START PING....."
read -r line
echo "ping namespaces.."
for i in `seq 2 $n`;
do
  sudo ip netns exec ns1 ping 10.10.1.${i} -c 1
  echo $?
done

echo "NOW Setting wrong security policy on mac,right on IP"
for i in `seq 1 $n`;
do
  ovn-nbctl lsp-set-port-security sw0-port${i} "01:00:02:00:03:00 10.10.1.${i}"
done

echo "PRESS A KEY TO START PING....."
read -r line
echo "ping namespaces.."
for i in `seq 2 $n`;
do
  sudo ip netns exec ns1 ping 10.10.1.${i} -c 1
  echo $?
done

echo "NOW Setting right sp on mac,wrong on IP"
for i in `seq 1 $n`;
do
  ovn-nbctl lsp-set-port-security sw0-port${i} "${mac[$i]} 10.10.2.1"
done

echo "PRESS A KEY TO START PING....."
read -r line
echo "ping namespaces.."
for i in `seq 2 $n`;
do
  sudo ip netns exec ns1 ping 10.10.1.${i} -c 1
  echo $?
done


echo "NOW Delete SP"
for i in `seq 1 $n`;
do
  ovn-nbctl lsp-set-port-security sw0-port${i} ""
done

echo "PRESS A KEY TO START PING....."
read -r line
echo "ping namespaces.."
for i in `seq 2 $n`;
do
  sudo ip netns exec ns1 ping 10.10.1.${i} -c 1
  echo $?
done

echo "OVN Cleanup"

for i in `seq 1 $n`;
do
  ovs-vsctl --no-wait del-port br-int veth${i}_
  ovn-nbctl lsp-del sw0-port${i}
done

ovn-nbctl ls-del sw0
