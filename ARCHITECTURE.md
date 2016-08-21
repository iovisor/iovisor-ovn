# Polito Controller

---
## Architecture

cli interface (politoctrl commandline)
|
|monitor(ovs; ovn Nb, Sb) ------------//----------> OVN Architerture
||
||logger (to print output of politod)
|||
*politod (daemon)*
|
|bpf modules repository
||
engine/orchestator ------------ db (local database)
|
helper(wrapper for Hover Apis)
|
|
//
|
|
hover
|
in-kernel dataplane impl
---

## politod

Polito Daemon is the running daemon that coordinates all other modules.

##  bpf

Contains all ebpf code

## common
Contains common utilities

## helper

Contains hover api wrapper

## monitor

Using libovsdb monitors ovs, ovn sb,nb


# Examples

Some Examples

---
# Overview

Polito Ctrl (*politod*) must have a library package of helper, to perform various requests to Hover(s).

from cli (*politoctrl*) or ovn-monitor I receive a request.
This request is routed into the engine, that performs the request to hover.
---
