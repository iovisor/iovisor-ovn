#! /bin/bash

_pwd=$(pwd)

# install bcc and dependencies
echo "installing bcc dependencies"
sudo apt-get install -y bison build-essential cmake flex git libedit-dev \
libllvm3.7 llvm-3.7-dev libclang-3.7-dev python zlib1g-dev libelf-dev

# echo installing bcc
cd $DATA_DIR
rm -rf bcc  # be sure that there is not al already git clone of bcc
git clone https://github.com/iovisor/bcc.git
cd bcc
mkdir build; cd build
cmake .. -DCMAKE_INSTALL_PREFIX=/usr
make -j
sudo make install

echo "installing go"
curl -O https://storage.googleapis.com/golang/go1.6.2.linux-amd64.tar.gz
tar -xvf go1.6.2.linux-amd64.tar.gz

sudo rm -rf /usr/local/go
sudo mv go /usr/local

sudo rm -f /usr/bin/go
sudo ln -s /usr/local/go/bin/go /usr/bin

export GOPATH=$HOME/go

# Hover
echo "installing hover dependencies"
go get github.com/vishvananda/netns
go get github.com/willf/bitset
go get github.com/gorilla/mux
# to pull customized fork of netlink
go get github.com/vishvananda/netlink
cd $HOME/go/src/github.com/vishvananda/netlink
git remote add drzaeus77 https://github.com/drzaeus77/netlink
git fetch drzaeus77
git reset --hard drzaeus77/master

echo "installing hover"
go get github.com/iovisor/iomodules/hover
# use custom version of hover
cd $GOPATH/src/github.com/iovisor/iomodules
git remote add mvbpolito https://github.com/mvbpolito/iomodules
git fetch mvbpolito
git reset --hard mvbpolito/master
go install github.com/iovisor/iomodules/hover/hoverd

echo "installing iovisor ovn"
go get github.com/iovisor/iovisor-ovn/iovisorovnd

cd $_pwd
