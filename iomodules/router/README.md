# Router IOModule

This module is a Router that implements a simplified version of routing algorithm, port management and arp responder.

## YAML Configuration Format

The following is an example of the configuration of a router:
```
[...]
- name: myrouter
  type: router
  config:
    interfaces:
      - name: Switch1
        ip: 10.0.1.254
        netmask: 255.255.255.0
        mac: "7e:ee:c2:01:01:01"

      - name: Switch2
        ip: 10.0.2.254
        netmask: 255.255.255.0
        mac: "7e:ee:c2:02:02:02"

      - name: Nat
        ip: 0.0.0.0
        netmask: 0.0.0.0
        mac: "7e:ee:c2:03:03:03"

    static_routes:
      - network: 130.192.0.0
        netmask: 255.255.0.0
        interface: Switch3
        next_hop: 10.2.2.254

      - network: 10.1.1.0
        netmask: 255.255.255.0
        interface: Switch4
        #If no next_hop set, route of type local

[...]
```

 - **interfaces**: defines local interfaces of the router. This primitive set automatically also the routing table entries of type local in order to reach hosts belonging to the defined network.

## API:
 * **AddRoutingTableEntry(network string, netmask string, outputIface string, nexthop string)**: Adds a routing table entry in the routing table of the router
  * network: network address to reach. (e.g. 10.0.1.0)
  * netmask: network address netmask. (e.g 255.255.255.0)
  * outputIface: output interface (router port).
  * nexthop: represents the next hop (e.g. 130.192.123.123). If 0 the network is locally reachable.

 * **AddRoutingTableEntryLocal(network string, netmask string, outputIface string)**: As previous API, but forces the nexthop parameter to 0 and force the network to be local to a router port.


 * **ConfigureInterface(ifaceName string, ip string, netmask string, mac string)**: Configure the interface Ip and Mac address of the router port. In addition set the correspondent routing table entry to reach the local network attachet to the port.
  * ifaceName: string that identifies the interface name (e.g. if the interface was previously attached to Switch1 use "Switch1")
  * ip: ip address of the interface
  * netmask: netmask address
  * mac: mac address of the port

## Limitations
 * The control plane must put ordered entries into the routing table, from longest to shortest prefix.

 * Routing table contains maximum 10 entries.

 * When the router has to forward a packet to an ip address without know the mac address, sent it in broadcast. When someone performs an arp request with that address router arp table is updated and l2 address is successfully completed.
