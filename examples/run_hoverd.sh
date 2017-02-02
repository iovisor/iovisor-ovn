#!/bin/bash

set -x

#Install Hover daemon
#go install github.com/iovisor/iomodules/hover/hoverd

#Launch Hover daemon
sudo $GOPATH/bin/hoverd -listen 127.0.0.1:5002
