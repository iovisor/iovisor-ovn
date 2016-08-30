# Polito Controller

---
## Architecture
```
cli interface (iovisor-ovn commandline)
|
|monitor(ovs; ovn Nb, Sb) ------------//----------> OVN Architerture
||
||logger (to print output of daemon)
|||
*iovisor-ovn (daemon)*
|
|bpf modules repository
||
engine (main logic)------------ db (local database)
|
hoverctl(wrapper for Hover Apis)
|
|
//
|
|
hover
|
in-kernel dataplane impl

```


## daemon

Daemon is the running daemon that coordinates all other modules.

##  bpf

Contains all ebpf code

## common
Contains common utilities

## hoverctl

Contains hover api wrapper

## monitor

Using libovsdb monitors ovs, ovn sb,nb


# Examples

Some Examples

---
# Overview

Polito Ctrl (*daemon*) must have a library package of hoverctl, to perform various requests to Hover(s).

from cli (*iovisor-ovn*) or ovn-monitor I receive a request.
This request is routed into the engine, that performs the request to hover.
