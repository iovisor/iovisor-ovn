# IOVisor-OVN

---
## Architecture

![IOVisor-OVN architecture](https://raw.githubusercontent.com/netgroup-polito/iovisor-ovn/master/docs/Architecture.png)

IOVisor-OVN sits on side of the traditional OVN architecture, it intercepts the
contents of the different databases and based on an implemented logic it deploys
the required network services using the IOVisor technology.

## IOVisor-OVN daemon

Daemon is the running daemon that coordinates all other modules.
The daemon is a central system that coordinates the deployment of network services
based on the contents of the OVN databases.

It is composed of different elements:

### OVN Monitor

Uses the [libovsdb](https://github.com/socketplane/libovsdb) to monitor all the
databases of OVN. (northbound, southbound and the local ovsdb on each compute node)

### Logger

It is a simple system that prints and if configured saves all the logs into a file.

### CLI

The command line interface allows to the user to interact with the system.
It is specially useful for troubleshooting and debugging purposes.
Provides an easy way to access the realtime modules status on the hypervisors, look at the current status of the mapping between OVN network elements and IOModules.

### Main Logic

This module implements the logic for deploying the network services across the
different compute nodes.

It receives a new service network request from OVN, process it and then uses the hover ctrl
interface to deploy those services in the different compute nodes.

### Hover Controller

The hover controller is a wrapper to sends command to the hover instances using
a RESTful API.

### IOModules Repository

This module is a local Repository that contains the implementation of the different IOModules.
