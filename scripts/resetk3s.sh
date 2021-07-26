#!/bin/sh

service k3s stop

rm -Rf /etc/rancher/k3s
rm -Rf /var/lib/rancher/k3s
rm -Rf /var/lib/cni/networks/cbr0

for name in $(ip -o link show | awk -F': ' '{print $2}' | sed  's/@.*//' | grep veth)
do
    r=`ip link show $name | grep cni0`
    if [ "$r" != "" ]; then
      echo "deleting $name"
      ip link delete $name
    fi
done

service k3s start

sleep 30
