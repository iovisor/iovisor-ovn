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
