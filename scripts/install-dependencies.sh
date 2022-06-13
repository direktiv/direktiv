#!/bin/bash

LIB_GIT=/usr/local/lib/libgit2.so.1.3
DIR=$(pwd)

if ! command -v apt-get &> /dev/null
then
    echo "apt-get could not be found and is required for this script to work"
    exit
fi

# Ask user for permissions
sudo -v

# Installing libgit2
if [ -e ${LIB_GIT} ]
then
    echo "- libgit2 is installed - SKIPPING"
else
    libgit_dir=$(mktemp -d -t libgit2-XXXXXXXXXX)
    echo "- libgit2 is not installed - INSTALLING"
    echo "- WORKING DIRECTORY: ${libgit_dir}"
    cd ${libgit_dir}
    sudo apt-get install git cmake build-essential -y
    sudo apt-get install pkg-config libssl-dev python3 zlib1g-dev libssh2-1-dev libssh2-1 -y
    sudo apt-get install libmbedtls-dev -y 
    sudo apt-get install libpcre3 libpcre3-dev -y
    sudo apt-get install wget -y

    git clone https://github.com/libgit2/libgit2.git
    cd libgit2
    git checkout v1.3.0
    mkdir build
    cd build

    wget https://www.libssh2.org/download/libssh2-1.10.0.tar.gz
    tar -xvzf libssh2-1.10.0.tar.gz

    mkdir libssh2-1.10.0/bin
    cd libssh2-1.10.0/bin
    cmake .. -DENABLE_ZLIB_COMPRESSION=ON -DBUILD_SHARED_LIBS=ON
    sudo cmake --build . --target install

    cd -
    cmake .. -DUSE_SSH=ON
    sudo cmake --build . --target install
    cd ${DIR}
fi



# Install protobuff
if command -v protoc-gen-go &> /dev/null
then
    echo "- protobuff is installed: SKIPPING"
else
    # Install complier
    proto_dir=$(mktemp -d -t proto-XXXXXXXXXX)
    echo "- protobuff is not installed: INSTALLING"
    echo "- WORKING DIRECTORY: ${proto_dir}"
    cd ${proto_dir}
    wget https://github.com/protocolbuffers/protobuf/releases/download/v21.1/protoc-21.1-linux-x86_64.zip
    unzip protoc-21.1-linux-x86_64.zip -d protoc
    sudo chmod +x ./protoc/bin/protoc
    sudo mv -v ./protoc/bin/protoc /usr/local/bin/protoc

    sudo mkdir -p /usr/local/bin/include/
    sudo mv  -v ./protoc/include/* /usr/local/bin/include/

    # Install plugin
    go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
fi