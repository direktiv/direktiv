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

install_minio()
{
  wget -O /bin/minio https://dl.min.io/server/minio/release/linux-amd64/minio
}

create_certs()
{
  openssl req -new -newkey rsa:4096 -days 3650 -nodes -x509 -subj "/C=AU/ST=Varsity/L=GoldCoast/O=direktiv/CN=direktiv.user" -keyout key.pem -out cert.pem >/dev/null 2>&1
  mkdir /miniocerts && mv key.pem /miniocerts/private.key && mv cert.pem /miniocerts/public.crt
  apk del openssl
}

apk update

if [ "$1" = "true" ] ; then
  mkdir -p /run/postgresql/
  mkdir -p /var/lib/postgresql/data
  apk add postgresql postgresql-contrib
fi

apk add ca-certificates iptables ip6tables device-mapper udev make go openssl util-linux
apk add podman --repository=http://dl-cdn.alpinelinux.org/alpine/edge/community

create_certs
install_firecracker
install_cni
install_minio
rm /builder.sh

# start the second script
touch /done
