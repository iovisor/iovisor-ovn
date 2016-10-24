# Test Environment

# Install Hover
```bash
# prereqs
# make sure you have exported $GOPATH to your workspace directory.
go get github.com/vishvananda/netns
go get github.com/willf/bitset
go get github.com/gorilla/mux
# to pull customized fork of netlink
go get github.com/vishvananda/netlink
cd $GOPATH/src/github.com/vishvananda/netlink
git remote add drzaeus77 https://github.com/drzaeus77/netlink
git fetch drzaeus77
git reset --hard drzaeus77/master

go get github.com/iovisor/iomodules/hover
go install github.com/iovisor/iomodules/hover/hoverd
```

# Launch Hover
```bash
#Launch in a diffent terminal
sudo $GOPATH/bin/hoverd -listen 127.0.0.1:5002
```

# Launch Iovisor-Ovn Daemon
```bash
#Launch in a different terminal
go install github.com/netgroup-polito/iovisor-ovn/daemon
sudo $GOPATH/bin/daemon -hover http://127.0.0.1:5002 -sandbox -sbsock /ovs/tutorial/sandbox/ovnsb_db.sock -nbsock /ovs/tutorial/sandbox/ovnnb_db.sock -ovssock /ovs/tutorial/sandbox/db.sock
```

# Launch Ovn Sandbox
```bash
#in /ovs/ main folder
make sandbox SANDBOXFLAGS="--ovn"
```

# Launch Test Script inside Ovn Sandbox Shell
```bash
./testScript.sh
```
