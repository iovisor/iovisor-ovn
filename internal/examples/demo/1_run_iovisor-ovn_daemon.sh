go install github.com/netgroup-polito/iovisor-ovn/daemon
sudo $GOPATH/bin/daemon -hover http://127.0.0.1:5002 -sandbox
