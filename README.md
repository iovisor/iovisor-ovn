# IOVisor-OVN

## What is IOVisor-OVN?
IOVisor-OVN project aims to extend the current [OVN](https://github.com/openvswitch/ovs/)
backend with [IOVisor](https://www.iovisor.org/) technology: create a new data plane that is semantically equivalent to the original OVS-based one, but based on IOVisor, that encloses eBPF and XDP technologies.

### Additional Documents
To know more details about the architecture please see [Architecture.md](/ARCHITECTURE.md)

To know how to install and test it please see [Install.md](/INSTALL.md)

## Repository Organization

The repository is organized in the following way

* bpf

Contains all ebpf code: switch with security ports

* cli

Contains the command line interface of IOVisor-OVN daemon.

* config

Contains default configuration that can be changed by passing different parameters at the daemon start.

* hoverctl

Contains hover api wrapper

* ovnmonitor

Using libovsdb monitors ovs, ovn sb,nb

* Examples

Some Examples, demos and useful files

