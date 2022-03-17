#!/bin/bash

SCRIPT="$( cd -- "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )"
HOSTNAME=$(echo $(hostname) | awk '{print tolower($0)}')

echo "creating registry for ${HOSTNAME}"

docker run --user 0:1000 -v $SCRIPT/share:/share flat/ubuntu-dev /share/run.sh ${HOSTNAME}

echo "copying cert into docker dir"
sudo mkdir -p /etc/docker/certs.d/${HOSTNAME}:5443
sudo cp $SCRIPT/share/out/${HOSTNAME}.crt /etc/docker/certs.d/${HOSTNAME}:5443/ca.crt

echo "restarting docker"
sudo service docker restart

docker stop registry-secure
docker rm registry-secure

echo "starting registry"
docker run -d \
  --restart=always \
  --name registry-secure \
  -v $SCRIPT/share/out:/certs \
  -e REGISTRY_HTTP_ADDR=0.0.0.0:443 \
  -e REGISTRY_HTTP_TLS_CERTIFICATE=/certs/${HOSTNAME}.crt \
  -e REGISTRY_HTTP_TLS_KEY=/certs/${HOSTNAME}.key \
  -p 5443:443 \
  registry:2

echo "register ca / registry with k3s"
SCRIPT_ESCAPE=$(echo $SCRIPT | sed 's/\//\\\//ig')
sed "s/HOSTNAME/$HOSTNAME/g; s/PATH/$SCRIPT_ESCAPE/g;" $SCRIPT/registries.yaml > /tmp/registries.yaml

sudo cp -r /tmp/registries.yaml /etc/rancher/k3s/registries.yaml
sudo service k3s restart