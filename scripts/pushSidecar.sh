#!/bin/bash

# call from direktiv as scripts/pushSidecar.sh

make docker-sidecar && docker tag sidecar localhost:5000/sidecar

docker push localhost:5000/sidecar

KUBECONFIG=/etc/rancher/k3s/k3s.yaml kubectl apply -f  scripts/test-action.yaml
