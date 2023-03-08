#!/bin/bash

cat <<-EOF | kubectl apply -f -
---
apiVersion: v1
kind: Namespace
metadata:
  name: knative-eventing
---
apiVersion: operator.knative.dev/v1beta1
kind: KnativeEventing
metadata:
  name: knative-eventing
  namespace: knative-eventing
spec:
  config:
    default-ch-webhook:
      default-ch-config: |
        clusterDefault:
          apiVersion: messaging.knative.dev/v1
          kind: InMemoryChannel
          spec:
            delivery:
              backoffDelay: PT0.5S
              backoffPolicy: exponential
EOF
 
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
