#!/bin/bash

#set -x

echo "FILE"
echo $1
echo ""

set -x

go install github.com/netgroup-polito/iovisor-ovn/iovisorovnd
sudo $GOPATH/bin/iovisorovnd -file $1 -hover http://127.0.0.1:5002

