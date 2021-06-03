#!/bin/bash

countdown() {
  echo "sleeping for 60 secs"
  secs=60
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

echo "starting k3s"

service k3s start

countdown

export KUBECONFIG=/etc/rancher/k3s/k3s.yaml

source scripts/knative/install-knative.sh

countdown
