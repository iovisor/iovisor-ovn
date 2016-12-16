# Switch

In this example a switch module is deployed and two virtual network interfaces
in different network namespaces are connected to it.
Connectivity is tested by pinging between the different namespaces

<center><a href="../../images/switch_tutorial.png"><img src="../../images/switch_tutorial.png" width=500></a></center>

## Preparing the network namespaces

In order to work, it is necessary to create two network namespaces and set two
pairs of veth interfaces.

Execute the 'setup.sh' script:

```bash
sudo ./setup.sh
```

## Launching hover

Before deploying the switch, it is necessary to lanuch the hover daemon:

```bash
export GOPATH=$HOME/go
sudo $GOPATH/bin/hoverd -listen 127.0.0.1:5002
```

## Deploying the switch

The 'test_switch.go' script deploys a switch module and connects the veth1 and veth2
interfaces to it

```bash
export GOPATH=$HOME/go
cd $GOPATH/src/github.com/netgroup-polito/iovisor-ovn/tutorials/switch
go run test_switch.go -hover http://127.0.0.1:5002
```

## Testing connectivity

Now you are able to test the connectivity pinging between the network interfaces
in the different network spaces

```bash
sudo ip netns exec ns1 ping 10.0.0.2
```
