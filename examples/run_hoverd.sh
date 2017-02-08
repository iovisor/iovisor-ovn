#!/bin/bash

set -x

#This script (install) and launches Hocer daemon, the framework used to deploy IOModules

#Install Hover daemon
#go install github.com/iovisor/iomodules/hover/hoverd

#Launch Hover daemon
sudo $GOPATH/bin/hoverd -listen 127.0.0.1:5002
