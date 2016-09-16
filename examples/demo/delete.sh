#!/bin/bash
set -x


ovs-vsctl --no-wait del-port br-int veth1_

read -r line
ovs-vsctl --no-wait del-port br-int veth2_

read -r line
ovn-nbctl lsp-del sw0-port1

read -r line
ovn-nbctl lsp-del sw0-port2

read -r line
ovn-nbctl ls-del sw0
