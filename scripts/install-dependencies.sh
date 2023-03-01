#!/bin/bash

DIR=$(pwd)

if ! command -v apt-get &> /dev/null
then
    echo "apt-get could not be found and is required for this script to work"
    exit
fi

# Ask user for permissions
sudo -v

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