# NAT IOModule

This module is a One to One NAT that implements internal to external and reverse natting.
In particular each internal ip address (if mapped), is translated into a new external address.

*notes*:
  * first port is always attached to internal network, second port to external one.
  * nat iomodule should be part of the code of the router. This is not possible for framework issues (hover does not allow to use 1+ eBPF programs inside the same iomodule).
  * this is a *transparent nat*:
   * always attach a nat to a router.
   * the layer 2 (arp request, layer 2 rewrite) is managed by the router.
   * the nat only modifies packet layers 3   

## API:

  * **SetAddressAssociation(internal_ip string, external_ip string)**: Set the NAT rules.
  * internal_ip: internal ip address.
  * external_ip: external ip address.


## How to use

Using iovisor-ovn daemon in standalone mode, the user can deploy and configure a single or a chain of IOModules.
The entire setup can be deployed starting from a YAML configuration file.

```bash
$GOPATH/bin/iovisorovnd -file <configuration.yaml>
```

Some examples are available in [/examples](./../../examples/) folder:
  * [Nat One-to-One and Router](./../../examples/nat-one-to-one_router/)

Please note that NAT One-to-One IOModule Must be deployed attached to a Router.

### YAML Configuration Format

The following is an example of the configuration of a NAT:
```
[...]
- name: Nat
  type: onetoonenat
  config:
    nat_entries:
    - internal_ip: 10.0.1.1
      external_ip: 130.192.1.1

    - internal_ip: 10.0.1.2
      external_ip: 130.192.1.2

[...]
```

  * **nat_entries**: defines the ip mapping.
  * **internl_ip**: is the internal IP address.
  * **external_ip**: is the correspondent external IP address.

## Limitations

  * The first port of the nat is always attached to the internal network.
  * The second port of the nat is always attached to the public network.
