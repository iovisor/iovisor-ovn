# IOVisor-OVN Installation Guide
Currently there is only support to test it with devstack.
## Devstak
In order to create a test environment with a single compute node you should perform the following steps.
1. Install a test system
    We recommend to use Ubuntu 16.04 as it is the only tested platform. Please take into consideration that devstack performs deep changes on the system, so the system should only be dedicated to this purpose.
2. Install devstack with iovisor-ovn as network provider
```
    git clone http://git.openstack.org/openstack-dev/devstack.git
    git clone https://github.com/netgroup-polito/networking-ovn/
    cd devstack
    cp ../networking-ovn/devstack/local.conf.sample local.conf
    ./stack.sh
```
3. When devstack finishes it shows something like
```
This is your host ip: 192.168.122.8
Horizon is now available at http://192.168.122.8/
Keystone is serving at http://192.168.122.8:5000/
The default users are: admin and demo
The password: password
2015-04-30 22:02:40.220 | stack.sh completed in 515 seconds.
```
After it you can instantiate virtual machines that will use iovisor-ovn as network provider.

Please note that currently only L2 networks are supported.
