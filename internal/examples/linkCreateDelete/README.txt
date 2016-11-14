Probable hoverd bug
on Ports DELETE

POST module
link1 = POST link module port1
link2 = POST link module port2
DELETE link1
DELETE link2

connect 2 ports to the module
then remove ports from the module

How to recreate the problem:
prerequisites:
github.com/iovisor/iomodules
github.com/netgroup-polito/iovisor-ovn

github.com/netgroup-polito/iovisor-ovn/examples/linkCreateDelete/hover_util_log.go.diff
contains the diff to enable DEBUG log in hover

1_nsCreate.sh
  Creates 2 namespaces & veth1_ veth2_ Interfaces

2_run_hoverd.sh
  launches hoverd with the correct parameters

(on a new terminal)
3_run_iovisor-ovn_daemon.sh
  launches iovisor-ovn daemon

here you can use the iov-ovn cli!
to recreate the problem
cli@iov-ovn$test

"test" command launched the test, then you can see the output!

(to re-create the problem you have to close hover and iov-ovn daemon and relaunch nsCreate.sh script)

output of hoverd:
INF: 2016/09/29 17:15:59 IOVisor HTTP Daemon starting...
DBG: 2016/09/29 17:15:59 server.go:217: Patch panel modules table loaded: map[key_size:4 leaf_size:4 key_desc:"int" leaf_desc:"int" name:modules fd:4]
DBG: 2016/09/29 17:15:59 netlink_monitor.go:98: link lo master 0
DBG: 2016/09/29 17:15:59 netlink_monitor.go:98: link enp0s3 master 0
DBG: 2016/09/29 17:15:59 graph.go:245: AddNode b:docker0 0
DBG: 2016/09/29 17:15:59 graph.go:245: AddNode b:virbr0 1
DBG: 2016/09/29 17:15:59 netlink_monitor.go:98: link virbr0-nic master 0
DBG: 2016/09/29 17:15:59 netlink_monitor.go:98: link veth4_ master 0
DBG: 2016/09/29 17:15:59 netlink_monitor.go:98: link veth5_ master 0
DBG: 2016/09/29 17:15:59 netlink_monitor.go:98: link veth3_ master 0
DBG: 2016/09/29 17:15:59 netlink_monitor.go:98: link veth1_ master 0
DBG: 2016/09/29 17:15:59 netlink_monitor.go:98: link veth2_ master 0
DBG: 2016/09/29 17:15:59 netlink_monitor.go:74: NewNetlinkMonitor DONE
INF: 2016/09/29 17:15:59 NetlinkMonitor=0xc82005a660
INF: 2016/09/29 17:15:59 Hover Server listening on 127.0.0.1:5002
DBG: 2016/09/29 17:16:08 graph.go:245: AddNode 2stub: 2
DBG: 2016/09/29 17:16:08 graph.go:250: RemoveNode 2stub: 2
DBG: 2016/09/29 17:16:08 graph.go:245: AddNode m:8f894b9b 2
INF: 2016/09/29 17:16:08 POST /modules/ 200
DBG: 2016/09/29 17:16:08 graph.go:245: AddNode i:veth1_ 3
INF: 2016/09/29 17:16:08 Provisioning i:veth1_ (3)
INF: 2016/09/29 17:16:08 Provisioning m:8f894b9b (2)
INF: 2016/09/29 17:16:09 visit: id=3 :: fd=12 :: i:veth1_ :: 50
INF: 2016/09/29 17:16:09     1: m:8f894b9b {[0x10002 0x0 0x0 0x0]}
INF: 2016/09/29 17:16:09 visit: 2 :: m:8f894b9b
INF: 2016/09/29 17:16:09     1: i:veth1_   {[0x10003 0x0 0x0 0x0]}
INF: 2016/09/29 17:16:09 POST /links/ 200
DBG: 2016/09/29 17:16:09 graph.go:245: AddNode i:veth2_ 4
INF: 2016/09/29 17:16:09 Provisioning i:veth1_ (3)
INF: 2016/09/29 17:16:09 Provisioning m:8f894b9b (2)
INF: 2016/09/29 17:16:09 Provisioning i:veth2_ (4)
INF: 2016/09/29 17:16:09 visit: id=3 :: fd=12 :: i:veth1_ :: 50
INF: 2016/09/29 17:16:09     1: m:8f894b9b {[0x10002 0x0 0x0 0x0]}
INF: 2016/09/29 17:16:09 visit: id=4 :: fd=13 :: i:veth2_ :: 52
INF: 2016/09/29 17:16:09     1: m:8f894b9b {[0x20002 0x0 0x0 0x0]}
INF: 2016/09/29 17:16:09 visit: 2 :: m:8f894b9b
INF: 2016/09/29 17:16:09     1: i:veth1_   {[0x10003 0x0 0x0 0x0]}
INF: 2016/09/29 17:16:09     2: i:veth2_   {[0x10004 0x0 0x0 0x0]}
INF: 2016/09/29 17:16:09 POST /links/ 200
INF: 2016/09/29 17:16:11 Provisioning i:veth1_ (3)
INF: 2016/09/29 17:16:11 Provisioning m:8f894b9b (2)
INF: 2016/09/29 17:16:11 Provisioning i:veth2_ (4)
INF: 2016/09/29 17:16:11 visit: id=3 :: fd=12 :: i:veth1_ :: 50
INF: 2016/09/29 17:16:11 visit: id=4 :: fd=13 :: i:veth2_ :: 52
INF: 2016/09/29 17:16:11     1: m:8f894b9b {[0x20002 0x0 0x0 0x0]}
INF: 2016/09/29 17:16:11 visit: 2 :: m:8f894b9b
INF: 2016/09/29 17:16:11     1: i:veth1_   {[0x0 0x0 0x0 0x0]}
INF: 2016/09/29 17:16:11     2: i:veth2_   {[0x10004 0x0 0x0 0x0]}
INF: 2016/09/29 17:16:11 DELETE /links/84e5abbb-fccd-5db3-7586-d02161e4547a 200
INF: 2016/09/29 17:16:11 Provisioning i:veth1_ (3)
INF: 2016/09/29 17:16:11 Provisioning m:8f894b9b (2)
INF: 2016/09/29 17:16:11 Provisioning i:veth2_ (4)
INF: 2016/09/29 17:16:11 visit: id=3 :: fd=12 :: i:veth1_ :: 50
/*ERROR HERE!!*/
ERR: 2016/09/29 17:16:11 Invalid # edges for node i:veth1_, must be 2, got 1
INF: 2016/09/29 17:16:11 DELETE /links/84e5abbb-fccd-5db3-7586-d02661e4547a 400
