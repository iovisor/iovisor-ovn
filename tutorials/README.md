# Tutorials

This folder contains a few examples that illustrate the capabilities of
iovisor-ovn.

Each subfolder contains a different example, usually on these there are three files

* README.md: presents the details of the example
* setup.sh: prepares the environment
* example.go: injects and configures the eBPF IOModules

Before running the examples it is necessary to install some components.
Please see [Installing](#installing).

# Examples
* [switch](/switch/): L2 switch connected to two virtual network interfaces
* [router](/router/): L3 router connected to three virtual network interfaces

# Installing Minimal

In order to execute the tutorials a minimal setup is required.

```bash
git clone https://github.com/netgroup-polito/iovisor-ovn.git
cd iovisor-ovn/tutorials/
./install-minimal.sh
```

For more details please follow [INSTALL-MINIMAL](INSTALL-MINIMAL.md).
