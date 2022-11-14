#!/usr/bin/env bash

apt-get install git cmake build-essential pkg-config libssl-dev python3 zlib1g-dev libssh2-1-dev libssh2-1 libmbedtls-dev libpcre3 libpcre3-dev wget -y

mkdir deps
cd deps

wget https://www.libssh2.org/download/libssh2-1.10.0.tar.gz
tar -xvzf libssh2-1.10.0.tar.gz

cd libssh2-1.10.0
mkdir bin
cd bin
cmake .. -DENABLE_ZLIB_COMPRESSION=ON -DBUILD_SHARED_LIBS=ON
cmake --build . --target install
cd ..
cd ..

git clone https://github.com/libgit2/libgit2.git
cd libgit2 
git checkout v1.3.0 
mkdir build 
cd build
cmake .. -DUSE_SSH=ON
cmake --build . --target install
cd ..
cd ..


