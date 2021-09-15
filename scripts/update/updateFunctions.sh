#!/bin/bash
./scripts/push/pushFunctions.sh
echo "Deleting Old Pod"
kubectl delete pod -l app.kubernetes.io/name=direktiv-functions
echo "Tailing New Pod"
kubectl logs -f -l app.kubernetes.io/name=direktiv-functions -c functions-controller
