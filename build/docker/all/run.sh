#!/bin/sh

echo "starting k3s"

/bin/k3s server --disable traefik --write-kubeconfig-mode=644 > /var/log/k3s.log 2>&1 &

until [ -f /etc/rancher/k3s/k3s.yaml ]
do
    echo "waiting..."
    sleep 1
done

echo "k3s running, waiting for 10 secs"

sleep 10

export KUBECONFIG=/etc/rancher/k3s/k3s.yaml

cd direktiv/scripts/knative &&  ./install-knative.sh && cd ../../..

kubectl get pods -A

cd direktiv/kubernetes/charts/direktiv && /helm install -f /debug.yaml direktiv .

echo "direktiv installed, pulling containers can take several minutes"
sleep 5

a=0
until [ $a -gt 50 ]
do
  clear
  kubectl get pods -A
  sleep 10
  a=`expr $a + 1`
done

echo "UI/API: localhost:8080"

sleep infinity
