# NAT IOModule

This module is a NAT that implements source address translation. In particular PAT algorithm is applied: source IP Address and Ports are changed in order to hide intern private ip addresses and exit on the internet (or another network) with only one public IP address.

## YAML Configuration Format

The following is an example of the configuration of a NAT:
```
[...]
- name: Nat
  type: nat
  config:
    public_ip: 10.10.1.100

[...]
```

  * **public_ip**: defines public ip address.

## API:
 * **SetPublicIp(ip string)**: Set public ip address
  * ip: public ip address. (e.g. 10.10.1.100)

## Limitations
 * The first port of the nat is always attached to the internal network.

 * The second port of the nat is always attached to the public network.

 * No cleanup is performed on the nat tables entries

 * The mechanism to choose the source port is incremental starting from port 1025.
