#!/bin/bash

KUBECONFIG=/etc/rancher/k3s/k3s.yaml helm install -f debug.yaml direktiv .
