#!/bin/bash

kubectl apply -f https://github.com/knative/eventing/releases/download/v0.26.1/eventing-crds.yaml
kubectl apply -f https://github.com/knative/eventing/releases/download/v0.26.1/eventing-core.yaml
kubectl apply -f https://github.com/knative/eventing/releases/download/v0.26.1/mt-channel-broker.yaml
kubectl apply -f https://github.com/knative/eventing/releases/download/v0.26.1/in-memory-channel.yaml

echo "waiting for ready"
kubectl wait --for=condition=ready pod -l app=mt-broker-controller -n knative-eventing

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

cat <<-EOF | kubectl apply -f -
---
apiVersion: eventing.knative.dev/v1
kind: Broker
metadata:
   name: default
   namespace: default
   annotations:
     eventing.knative.dev/broker.class: MTChannelBasedBroker
EOF
