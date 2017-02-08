# DHCP Server IOModule

This module implements a basic version of a DHCP server.

## API

- **ConfigureParameters(...)**
Configures the parameters required to the operation of the DHCP server.
The received arguments are the same used in the YAML configuration file.

## How to use

Using iovisor-ovn daemon in standalone mode, the user can deploy and configure a single or a chain of IOModules.
The entire setup can be deployed starting from a YAML configuration file.

```bash
$GOPATH/bin/iovisorovnd -file <configuration.yaml>
```

Some examples are available in [/examples](./../../examples/) folder:
 * [DHCP](./../../examples/dhcp/)

### YAML Configuration Format

The following is an example of the configuration of a dhcp server:
```
[...]
  - name: mydhcp
    type: dhcp
    config:
      netmask: 255.255.255.0
      addr_low: 192.168.1.100
      addr_high: 192.168.1.150
      dns: 8.8.8.8
      router: 192.168.1.1
      lease_time: 3600
      server_ip: 192.168.1.250
      server_mac: "b6:87:f8:5a:40:23"
[...]
```

 - **netmask**: mask of the network segment where the DHCP server is.
 - **addr_low**: first ip address that the server can assign.
 - **addr_high**: last ip address that the server can assign.
 - **dns**: DNS ip address that the server offers to the clients.
 - **router**: default gateway assigned to clients.
 - **lease_time**: default time that an address is leased to a client.
 - **server_ip**: IP address of the DHCP server.
 - **mac_ip**: MAC address of the DHCP server.

## Limitations

- Maximum 10 addresses can be assigend by the DHCP server.
- There is not support for address rebinding.
- There is not support for address releasing.
