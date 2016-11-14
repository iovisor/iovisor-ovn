#!/bin/bash
set -x


ovs-vsctl --no-wait del-port br-int veth1_

read -r line
ovs-vsctl --no-wait del-port br-int veth2_

read -r line
ovs-vsctl --no-wait del-port br-int veth3_

read -r line
ovs-vsctl --no-wait del-port br-int veth4_

read -r line
ovs-vsctl --no-wait del-port br-int veth5_



read -r line
ovn-nbctl lsp-del sw0-port1

read -r line
ovn-nbctl lsp-del sw0-port2

read -r line
ovn-nbctl lsp-del sw0-port3

read -r line
ovn-nbctl lsp-del sw0-port4

read -r line
ovn-nbctl lsp-del sw0-port5


read -r line
ovn-nbctl ls-del sw0
