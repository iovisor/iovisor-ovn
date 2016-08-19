echo "sudo ip netns exec ns1 ping 10.10.1.2 -c 2"
sudo ip netns exec ns1 ping 10.10.1.2 -c 2
echo "sudo ip netns exec ns1 ping 10.10.1.3 -c 2"
sudo ip netns exec ns1 ping 10.10.1.3 -c 2
echo "sudo ip netns exec ns2 ping 10.10.1.1 -c 2"
sudo ip netns exec ns2 ping 10.10.1.1 -c 2
echo "sudo ip netns exec ns2 ping 10.10.1.3 -c 2"
sudo ip netns exec ns2 ping 10.10.1.3 -c 2
echo "sudo ip netns exec ns3 ping 10.10.1.1 -c 2"
sudo ip netns exec ns3 ping 10.10.1.1 -c 2
echo "sudo ip netns exec ns3 ping 10.10.1.2 -c 2"
sudo ip netns exec ns3 ping 10.10.1.2 -c 2

