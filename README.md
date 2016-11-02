# IOVisor-OVN

## What is IOVisor-OVN?
IOVisor-OVN project aims to extend the current [OVN](https://github.com/openvswitch/ovs/)
backend with [IOVisor](https://www.iovisor.org/) technology: create a new data plane that is semantically equivalent to the original OVS-based one, but based on IOVisor, that encloses eBPF and XDP technologies.

### Why?

 - Complex and efficient virtualized network services are becoming important.
 - Complex services cannot be implemented with only flow-based switches (as OvS).
 - eBPF is integrated in the linux kernel and allows to create new functions at runtime.

### How?

 - Replace the current backend of OVN with a new implementation based on IOVisor
 - This proposal fits along the current OVN architecture, keeping compatibility with current Cloud Management Systems as OpenStack

## Architecture

![IOVisor-OVN architecture](https://raw.githubusercontent.com/netgroup-polito/iovisor-ovn/master/docs/iovisor-ovn-architecture.png)

IOVisor-OVN sits on side of the traditional OVN architecture, it intercepts the
contents of the different databases and based on an implemented logic it deploys
the required network services using the IOVisor technology.

For more details about the architecture please see [ARCHITECTURE.md](/ARCHITECTURE.md)

## Install

It is possible to install and deploy a complete OpenStack environment with IOVisor-OVN as network backend.
The process is automatically managed by DevStack scripts.
For now only L2 networking is supported.

```
    git clone http://git.openstack.org/openstack-dev/devstack.git
    git clone https://github.com/netgroup-polito/networking-ovn/
    cd devstack
    cp ../networking-ovn/devstack/local.conf.sample local.conf
    ./stack.sh
```

For more details about install and test please see [INSTALL.md](/INSTALL.md)

## Repository Organization

The repository is organized in the following way

* **bpf** contains all ebpf code: switch with security ports

* **cli** contains the command line interface of IOVisor-OVN daemon.

* **config** contains default configuration that can be changed by passing different parameters at the daemon start.

* **daemon** contains the daemon main program entry point.

* **hoverctl** contains hover restful api wrapper and utilities.

* **mainlogic** performs the mapping between the network configuration of OVN and IOModules.

* **ovnmonitor** monitors for ovn northbound database, southbound database and local ovs database. Implemented using libovsdb.
