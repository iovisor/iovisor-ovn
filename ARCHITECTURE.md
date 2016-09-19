# IOVisor-OVN

---
## Architecture

![IOVisor-OVN architecture](docs/architecture.png)

IOVisor-OVN sits on side of the traditional OVN architecture, it intercepts the
contents of the different databases and based on an implemented logic it deploys
the required network services using the IOVisor technology.

## IOVisor-OVN daemon

Daemon is the running daemon that coordinates all other modules.
The deamon is a central system that coordinates the deployment of network services
based on the contents of the OVN databases.

It is composed of different elements:

### OVN Monitor

Uses the [libovsdb](https://github.com/socketplane/libovsdb) to monitor all the
databases of OVN. (northbound, southbound and the local ovsdb on each compute node)

### Logger

It is a simple system that prints and if configured saves all the logs into a file.

### CLI

The command line interface allows to the user to interact with the system.
It is specially usefull for troubleshooting and debugging purposes

### Main Logic

This module implements the logic for deploying the network services accross the
different compute nodes.

It receives a new service network request, process it and then uses the hover ctrl
interface to deply those services in the different compute nodes.

### Hover ctrl

The hovel ctrl is a wrapper that sends command to the hover instances using the
a REST api.

### IOModule Repo

This module is a database that contains the implementation of the different IOModules.
