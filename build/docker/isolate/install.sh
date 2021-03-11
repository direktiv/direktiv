#!/bin/sh

FC="v0.23.3"
CNI="v0.9.1"

install_firecracker()
{

  echo "installing firecracker $FC"

  mkdir -p /srv/jailer
  mkdir -p /usr/local/bin/

  wget https://github.com/firecracker-microvm/firecracker/releases/download/$FC/Firecracker-$FC-x86_64.tgz

  tar --strip 1 -C /usr/local/bin/ -xvf Firecracker-$FC-x86_64.tgz release-$FC/firecracker-$FC-x86_64; \
  mv /usr/local/bin/firecracker-$FC-x86_64  /usr/local/bin/firecracker; \
  tar --strip 1 -C /usr/local/bin/ -xvf Firecracker-$FC-x86_64.tgz release-$FC/jailer-$FC-x86_64; \
  mv /usr/local/bin/jailer-$FC-x86_64  /usr/local/bin/jailer; \
  rm Firecracker-$FC-x86_64.tgz

}

install_cni()
{
  echo "installing cni"
  mkdir -p /opt/cni/bin

  wget https://github.com/containernetworking/plugins/releases/download/$CNI/cni-plugins-linux-amd64-$CNI.tgz
  tar -C /opt/cni/bin -xvf cni-plugins-linux-amd64-$CNI.tgz
  rm -Rf cni-plugins-linux-amd64-$CNI.tgz

  wget https://github.com/awslabs/tc-redirect-tap/archive/master.zip;
  unzip -o master.zip;
  cd tc-redirect-tap-master && GOPATH=/tmp GOCACHE=/tmp make && cp tc-redirect-tap /opt/cni/bin && cd ..
  rm -Rf master.zip && rm -Rf tc-redirect-tap-master

  apk del go
}

apk update

apk add ca-certificates iptables ip6tables device-mapper udev make go openssl util-linux
apk add podman --repository=http://dl-cdn.alpinelinux.org/alpine/edge/community

install_firecracker
install_cni
