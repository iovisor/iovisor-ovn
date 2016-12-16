# Router

In this example a router module is deployed and three virtual network interfaces
in different network namespaces are connected to it.
In this case the three virtual network interfaces are configured in different
subnets, then a router mechanishm is needed to allow them to exchange data packets.
Connectivity is tested by pinging between the different namespaces

## Preparing the network namespaces

The configuration of the network namespaces and the virtual ethernet interfaces
is very similar to the switch case.
The main difference here is that the three network interfaces are in different
subnets and a default route is configured on each namespace.

Execute the 'setup.sh' script:

```bash
sudo ./setup.sh
```

## Launching hover

Before deploying the router, it is necessary to launch the hover daemon:

```bash
export GOPATH=$HOME/go
sudo $GOPATH/bin/hoverd -listen 127.0.0.1:5002
```

## Deploying the router

The 'test_router.go' script deploys a router module and connects the veth1,
veth2 and veth3 interfaces to it

```bash
export GOPATH=$HOME/go
cd $GOPATH/src/github.com/netgroup-polito/iovisor-ovn/tutorials/router
go run test_router.go -hover http://127.0.0.1:5002
```

## Testing connectivity

Now you are able to test the connectivity pinging between the network interfaces
in the different network spaces, for example:

```bash
# ping ns2 from ns1
sudo ip netns exec ns1 ping 10.0.2.100
# ping ns3 from ns1
sudo ip netns exec ns1 ping 10.0.3.100
# ping ns1 from ns3
sudo ip netns exec ns3 ping 10.0.1.100
```
