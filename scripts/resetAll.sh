#!/bin/bash

dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"

countdown() {
  echo "sleeping for 30 secs"
  secs=30
  shift
  while [ $secs -gt 0 ]
  do
    printf "\r\033[Kwaiting %.d seconds" $((secs--))
    sleep 1
  done
  echo
}

echo "stopping k3s"

service k3s stop

echo "deleting k3s data"

rm -Rf /etc/rancher/k3s
rm -Rf /var/lib/rancher/k3s
rm -rf /var/lib/cni/networks/cbr0

for name in $(ip -o link show | awk -F': ' '{print $2}' | sed  's/@.*//' | grep veth)
do
    r=`ip link show $name | grep cni0`
    if [ "$r" != "" ]; then
      echo "deleting $name"
      ip link delete $name
    fi
done

echo "starting k3s"

service k3s start

countdown

export KUBECONFIG=/etc/rancher/k3s/k3s.yaml

helm install --set development=true knative $dir/../kubernetes/charts/knative

countdown
