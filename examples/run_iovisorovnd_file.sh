#!/bin/bash

#set -x

#This script (install) and launches iovisor-ovn daemon in standalone mode.
#The standalone mode allows the user to deploy a single or a chain of IOModules
#using a YAML configuration file.
#The -file parameter starts the Standalone mode

echo "FILE"
echo $1
echo ""

set -x

#Install iovisorovnd
#go install github.com/iovisor/iovisor-ovn/iovisorovnd

#Launch iovisorovndn using file parameter
sudo $GOPATH/bin/iovisorovnd -file $1 -hover http://127.0.0.1:5002
