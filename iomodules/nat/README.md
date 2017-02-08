# NAT IOModule

This module is a NAT that implements source address translation. In particular PAT algorithm is applied: source IP Address and Ports are changed in order to hide intern private ip addresses and exit on the internet (or another network) with only one public IP address.

*notes*:
  * first port is always attached to internal network, second port to external one.
  * nat iomodule should be part of the code of the router. This is not possible for framework issues (hover does not allow to use 1+ eBPF programs inside the same iomodule).
  * this is a *transparent nat*:
   * always attach a nat to a router.
   * the layer 2 (arp request, layer 2 rewrite) is managed by the router.
   * the nat only modifies packet layers 3-4   

## API:

 * **SetPublicIp(ip string)**: Set public ip address
  * ip: public ip address. (e.g. 10.10.1.100)

## How to use

Using iovisor-ovn daemon in standalone mode, the user can deploy and configure a single or a chain of IOModules.
The entire setup can be deployed starting from a YAML configuration file.

```bash
$GOPATH/bin/iovisorovnd -file <configuration.yaml>
```

Some examples are available in [/examples](./../../examples/) folder.
 * [Nat and Router](./../../examples/nat_router/)

Please note that NAT IOModule Must be deployed attached to a Router.

### YAML Configuration Format

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

## Limitations
    * The first port of the nat is always attached to the internal network.

    * The second port of the nat is always attached to the public network.

    * No cleanup is performed on the nat tables entries

    * The mechanism to choose the source port is incremental starting from port 1025.
