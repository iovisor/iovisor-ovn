
sudo ip netns exec ns1 ping 2.2.2.2 -c 2
sudo ip netns exec ns1 ping 3.3.3.3 -c 2

sudo ip netns exec ns2 ping 1.1.1.1 -c 2
sudo ip netns exec ns2 ping 3.3.3.3 -c 2

sudo ip netns exec ns3 ping 1.1.1.1 -c 2
sudo ip netns exec ns3 ping 2.2.2.2 -c 2
