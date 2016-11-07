# IOVisor-OVN

IOVisor-OVN aims at extending the current [OVN](https://github.com/openvswitch/ovs/)
backend with [IOVisor](https://www.iovisor.org/) technology: creates a new data plane that is semantically equivalent to the original OvS-based one, but based on IOVisor, that is based on the eBPF virtual machine and can be integrated with the XDP technology.

### Why?

 - Complex and efficient virtualized network services are becoming important
 - Complex services cannot be implemented with only OpenFlow-based switches (as OvS), and the current model that mixes different technologies (Linux containers, openFlow switches, VMs, and more) in order to setup a complex network service is difficult to manage
 - eBPF is integrated in the Linux kernel and allows to create and deploy new functions at runtime.

### How?

 - Replace the current backend of OVN with a new implementation based on IOVisor
 - This proposal fits along the current OVN architecture, keeping compatibility with current Cloud Management Systems as OpenStack

## Architecture

![IOVisor-OVN architecture](https://raw.githubusercontent.com/netgroup-polito/iovisor-ovn/master/docs/iovisor-ovn-overview.png)

IOVisor-OVN sits on side of the traditional OVN architecture, it intercepts the contents of the different databases and based on an implemented logic it deploys the required network services using the IOVisor technology.

For more details about the architecture please see [ARCHITECTURE.md](/ARCHITECTURE.md)

## Install

It is possible to install and deploy a complete OpenStack environment with IOVisor-OVN as network backend.
The process is automatically managed by DevStack scripts.
Currently only L2 networks are supported.

For more details about install and test please see [INSTALL.md](/INSTALL.md)

## Repository Organization

The repository is organized in the following way

* **bpf** contains all eBPF code: switch with security ports

* **cli** contains the command line interface of IOVisor-OVN daemon.

* **config** contains default configuration that can be changed by passing different parameters at the daemon start.

* **daemon** contains the daemon main program entry point.

* **hoverctl** contains hover restful api wrapper and utilities.

* **mainlogic** performs the mapping between the network configuration of OVN and IOModules.

* **ovnmonitor** monitors for OVN northbound database, southbound database and local ovs database. Implemented using libovsdb.
