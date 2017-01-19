# IOVisor-OVN Standalone Installation

This alternative way of using IOVisor-OVN is intended to provide a fast and easy way to test the different IOModules.

## Installation
This guide includes the instructions to install *bcc, go, hover* and *iovisor-ovn*.
A Linux kernel with a version 4.9 or newer is required, we recommend to use Ubuntu as all the examples have been tested on that platform.

### Automatic Installation

In order to automatically install all the required components please execute the following commands, please consider that they only work on Debian based distros.

```bash
git clone https://github.com/netgroup-polito/iovisor-ovn.git
cd iovisor-ovn/tutorials/
./install-minimal.sh
```

### Manual Installation

Please follow these steps to install the required components manually.

#### bcc

Follow the steps indicated in [BCC Installation Guide](https://github.com/iovisor/bcc/blob/master/INSTALL.md) in order to install bcc, please consider that the version 0.2.0 is required.

#### go

Please follow the [Go Installation Guide](https://golang.org/doc/install), Go 1.4 or higher is required

#### hover
In order to install hover please follow the steps in [Installing Hover](https://github.com/iovisor/iomodules/#installing-hover)

Please bet sure that the this [patch](https://github.com/mvbpolito/iomodules/commit/7409078fcb158263dcc2b6b58b508e7033865d5f) is applied before the installation.

#### iovisor-ovn

Installing iovisor-ovn is very easy, just use the go get command:

```bash
go get github.com/netgroup-polito/iovisor-ovn/iovisorovnd
```

## How to use?

In this case the configuration and topology of the service topology to be deployed is passed through a .yaml file.

### Configuration File

```
# list of modules to be deployed
modules:
  - name: modulename
    type: moduletype
    config:
      # module configuration

# links between modules
links:
  - from: moduleName
    to: moduleName

# connection to network interfaces
external_interfaces:
  - module: moduleName
    iface: interface name
```
The file is composed of three sections: modules, links and external_interfaces.

1. **modules**: This section contains the modules to be deployed. 
The name and type are mandatory, while the configuration is optional and different for each kind of IOModules. 
Please see the documentation of each single IOModule to get information about the configuration parameters.

2. **links**: These are the links between the different IOModules, "from" and "to" must correspond to the name of modules in the "modules" section.

3. **external_interfaces**:  The connection to the network interfaces are defined in this section. Module should be a module defined in the "modules" section and iface should be the name of the interface on the system.

### Launching IOVisor-OVN

In order to deploy a service topology from a file, the -file parameter should be passed to the IOVisor-OVN daemon.

```
export GOPATH=$HOME/go
cd $GOPATH/src/github.com/netgroup-polito/iovisor-ovn/examples/switch
$GOPATH/bin/iovisorovnd -file <file.yaml> -hover <hover_url>
```

### Examples

Some examples with a complete explanation are provided in [examples](./examples)

