# IOVisor-OVN

## What is IOVisor-OVN?
IOVisor-OVN project aims to extend the current [OVN](https://github.com/openvswitch/ovs/)
backend with [IOVisor](https://www.iovisor.org/) technology: create a new data plane that is semantically equivalent to the original OVS-based one, but based on IOVisor, that encloses eBPF and XDP technologies.

##Architecture


## Repository Organization

The repository is organized in the following way

* bpf

Contains all ebpf code

* common

Contains common utilities

* hoverctl

Contains hover api wrapper

* ovnmonitor

Using libovsdb monitors ovs, ovn sb,nb

* Examples

Some Examples, demos and useful files

### Documentation is under construction
