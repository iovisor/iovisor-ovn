sudo ip netns del ns1
sudo ip netns del ns2
sudo ip netns del ns3

sudo ip link del veth1
sudo ip link del veth2
sudo ip link del veth3

sudo ip link del veth1_
sudo ip link del veth2_
sudo ip link del veth3_

sudo ip netns add ns1
sudo ip netns add ns2
sudo ip netns add ns3


sudo ip link add veth1 type veth peer name veth1_
sudo ip link set veth1 netns ns1
sudo ip link add veth2 type veth peer name veth2_
sudo ip link set veth2 netns ns2
sudo ip link add veth3 type veth peer name veth3_
sudo ip link set veth3 netns ns3
sudo ip netns exec ns1 ip link set dev veth1 up
sudo ip link set dev veth1_ up
sudo ip netns exec ns1 ifconfig veth1 10.10.1.1/24
sudo ip netns exec ns2 ip link set dev veth2 up
sudo ip link set dev veth2_ up
sudo ip netns exec ns2 ifconfig veth2 10.10.1.2/24
sudo ip netns exec ns3 ip link set dev veth3 up
sudo ip link set dev veth3_ up
sudo ip netns exec ns3 ifconfig veth3 10.10.1.3/24

