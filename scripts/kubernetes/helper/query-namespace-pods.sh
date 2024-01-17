#!/bin/sh

kubectl get pod -A -o jsonpath="{.items[?(@.metadata.labels.direktiv\.io/namespace-name==\"$1\")].metadata.name}" 

echo ""