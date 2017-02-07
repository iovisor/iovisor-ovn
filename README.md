# IOVisor-OVN

IOVisor-OVN extends the current [Open Virtual Networking (OVN) project](https://github.com/openvswitch/ovs/) with a new backend based on the [IOVisor](https://www.iovisor.org/) technology.
In a nutshell, IOVisor-ION defines a new data plane that is semantically equivalent to the original one, mostly based on Open vSwitch. The new data plane exploits the eBPF virtual machine (also known as IOVisor) and in future it could be integrated with the eXpress Data Path (XDP) technology for improved performance.

### Why?

 - Sophisticated and efficient virtualized network services are becoming important, and cannot directly be realized using the match/action paradigm implemented by current virtual switches;
 - Complex services cannot be carried out with only OpenFlow-based switches (as OvS), and the current model that mixes different technologies (Linux containers, openFlow switches with the associated controller for the control plane, virtual machines, and more) to setup a complex network service, is hard to manage;
 - eBPF is integrated in the Linux kernel and allows to create and deploy (i.e., *inject*) new functions at runtime, without having to upgrade/modify anything in the hosting server.

### How?

 - We replace the current backend of OVN with a new implementation based on IOVisor. This proposal maintains the current OVN architecture that handles orchestration across a datacenter-wise environment, and keeps compatibility with current Cloud Management Systems as OpenStack, Apache Mesos, and other.

## Architecture

<center><a href="images/iovisor-ovn-overview.png"><img src="images/iovisor-ovn-overview.png" width=700></a></center>

IOVisor-OVN sits on side of the traditional OVN architecture, it intercepts the contents of the different databases and deploys the required network services using the IOVisor technology.

For more details about the architecture please see [ARCHITECTURE](./ARCHITECTURE.md).

## Getting Started

IOVisor-OVN can be used in two different ways:

- **Default**: provides a network backend, based on IOVisor and IOModules, for OVN + OpenStack.  
Please read [README_OPENSTACK](./README_OPENSTACK.md).
- **Standalone**: allows to test a single or a chain of IOModules, without the OVN integration. The entire chain can be setup with a simple  yaml configuration file.  
Please read [README_STANDALONE](./README_STANDALONE.md).

## Examples

Some examples intended to provide a step-by-step guide to playing with the existing IOModules are available in the [examples folder](/examples) and can be used in *standalone* mode.


## Repository Organization:

* **iomodules**: contains eBPF code (i.e., available IOModules).
* **cli**: tool that implements the command line interface of IOVisor-OVN daemon.
* **config**: contains a file with the default configuration parameters used when the daemon starts.
* **iovisorovnd**: contains the daemon main program entry point.
* **docs**: documentation about this project, presentations, talks.
* **hoverctl**: hover restful API wrapper and utilities.
* **mainlogic**: tool that performs the mapping between the network configuration of OVN and IOModules.
* **ovnmonitor**:  monitors for OVN northbound database, southbound database and local OvS database. Implemented using libovsdb.
* **examples**: examples using IOModules in the repository.

## Presentations

  * OVS Fall Conference, San Jose, Nov 2016: [Slides](http://openvswitch.org/support/ovscon2016/7/1245-bertrone.pdf), [Video](https://www.youtube.com/watch?v=9cmR2NuAGz0)

## Licence

The IOVisor-OVN is licensed under the [Apache License, Version 2.0](./LICENSE.txt).
