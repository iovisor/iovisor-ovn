#!/bin/bash
#set -x

#POSTing 5 switch  with 10 ports each

n=10
nswitch=5

echo "cleanup previous namespaces"
for j in `seq 1 $nswitch`;
do
  for i in `seq 1 $n`;
  do
    sudo ip netns del ns${j}${i}
  done
done

for j in `seq 1 $nswitch`;
do
  for i in `seq 1 $n`;
  do
    ovs-vsctl --no-wait del-port br-int veth${j}${i}_
    ovn-nbctl lsp-del sw${j}-port${i}
  done
done
sleep 1

mac="a"

echo "PRESS A KEY TO START TEST CONFIG....."
read -r line

echo "setting up namespaces and veth"

for j in `seq 1 $nswitch`;
do
  for i in `seq 1 $n`;
  do
    sudo ip netns add ns${j}${i}
    sudo ip link add veth${j}${i} type veth peer name veth${j}${i}_
    sudo ip link set veth${j}${i} netns ns${j}${i}
    sudo ip netns exec ns${j}${i} ip link set dev veth${j}${i} up
    sudo ip link set dev veth${j}${i}_ up
    sudo ip netns exec ns${j}${i} ifconfig veth${j}${i} 10.10.1.${i}/24
    # new_mac=$(sudo ip netns exec ns${j}${i} ifconfig | grep veth${j}${i} | awk '{print $5}')
    # mac=(${mac[@]} $new_mac)
  done
done
echo "adding logical-switch"
for j in `seq 1 $nswitch`;
do
  ovn-nbctl ls-add sw${j}
done

for j in `seq 1 $nswitch`;
do
  for i in `seq 1 $n`;
  do
    echo "adding logical-switch port sw${j}-port${i}"
    ovn-nbctl lsp-add sw${j} sw${j}-port${i}
    # ovn-nbctl lsp-set-port-security sw0-port${i} "${mac[$i]} 10.10.1.${i}"
    ovs-vsctl --no-wait add-port br-int veth${j}${i}_ -- set Interface veth${j}${i}_ external_ids:iface-id=sw${j}-port${i}
  done
done

echo "PRESS A KEY TO START PING....."
read -r line

echo "ping namespaces.."
success=1

for j in `seq 1 $nswitch`;
do
  for i in `seq 2 $n`;
  do
    echo "ns${j}1"
    sudo ip netns exec ns${j}1 ping 10.10.1.${i} -c 1
  done
done

echo "OVN Cleanup"

for j in `seq 1 $nswitch`;
do
  for i in `seq 1 $n`;
  do
    ovs-vsctl --no-wait del-port br-int veth${j}${i}_
    ovn-nbctl lsp-del sw${j}-port${i}
  done
done

for j in `seq 1 $nswitch`;
do
  ovn-nbctl ls-del sw${j}
done

sleep 1

for j in `seq 1 $nswitch`;
do
  for i in `seq 1 $n`;
  do
    sudo ip netns del ns${j}${i}
  done
done
