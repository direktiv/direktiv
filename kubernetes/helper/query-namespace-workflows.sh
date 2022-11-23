#!/bin/sh

kubectl get pod -A -o jsonpath="{.items[?(@.metadata.labels.direktiv\.io/workflow-name==\"$1\")].metadata.name}" 
echo ""