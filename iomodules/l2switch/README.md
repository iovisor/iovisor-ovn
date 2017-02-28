# Switch IOModule

This module is an Ethernet Switch that implements the MAC learning algorithm.

## API

- **AddForwardingTableEntry(mac string, ifaceName string)**
Adds a static entry in the forwarding table of the switch
-- mac: MAC address. It must be in the "xx:yy:zz:xx:xx:xx" format
-- ifaceName: name of the port where MAC can be reached

## How to use

Using iovisor-ovn daemon in standalone mode, the user can deploy and configure a single or a chain of IOModules.
The entire setup can be deployed starting from a YAML configuration file.

```bash
$GOPATH/bin/iovisorovnd -file <configuration.yaml>
```

Some examples are available in [/examples](./../../examples/) folder:
 * [Switch](./../../examples/switch/)

### YAML Configuration Format

The following is an example of the configuration of a switch:
```
[...]
  - name: myswitch
    type: switch
    config:
      forwarding_table:
      - port: veth1
        mac: "b2:1b:34:5d:9b:2d"
      - port: veth2
        mac: "b2:1b:34:5d:9b:2e"
[...]
```

 - **forwarding _table**: defines static entries for the forwarding table of the switch. Please note that this configuration parameter is optional. It's useful when the programmer wants to force entries in the forwarding table.

## Limitations

- There is not a mechanishm to clean up the forwarding table, all the entries remain there until the module is unloaded.  This behavior could cause issues when the number of entries reaches the maximum.
This issue is not on the immediate roadmap, however could be solved using some sort of timeout mechanishm provided by the eBPF maps.
