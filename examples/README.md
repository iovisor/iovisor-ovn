# Examples

This folder contains a few examples that illustrate the capabilities of
iovisor-ovn.

Each subfolder contains a different example, usually on these there are three files

* README.md: presents the details of the example
* setup.sh: prepares the environment
* *.yaml: contains the configuration of the IOModules and its connections.

Before running the examples it is necessary to install some components.
Please see [Readme Standalone](../README_STANDALONE.md).

## Available Examples
* [switch](switch): L2 switch connected to two virtual network interfaces
* [router](router): L3 router connected to three virtual network interfaces
* [dhcp](dhcp): DHCP IOModule connected to a switch
* [Nat and Router](nat_router): Nat iomodule connected to a Router
* [Nat One-to-One and Router]: Nat One-to-One connected to a Router
