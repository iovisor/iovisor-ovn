# Installing

This guide includes the instructions to install bcc, go, hover and iovisor-ovnm
those are the necessary components to run the examples.

A linux kernel with a version 4.1 or newer is required, we recommend to use
Ubuntu Xenial as all the examples have been tested on that platform.

## Automatic Installation

The script install.sh tries to automatically download and install all the
components that are necessary, it is only designed to work on Debian based distros.

## Manual Installation

Please follow these steps to install the required components manually.

### bcc

Follow the steps indicated in
[BCC Installation Guide](https://github.com/iovisor/bcc/blob/master/INSTALL.md)
in order to install bcc, please consider that the version 0.2.0 is required.

### go

Please follow the [Go Installation Guide](https://golang.org/doc/install),
Go 1.4 or higher is required

### Hover
In order to install hover please follow the steps in
[Installing Hover](https://github.com/iovisor/iomodules/#installing-hover)

### iovisor-ovn

Intalling iovisor-ovn is very easy, just use the go get command:

```bash
go get github.com/netgroup-polito/iovisor-ovn/daemon
```

After these steps you are ready to run the examples.
