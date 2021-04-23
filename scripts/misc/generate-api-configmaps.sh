#!/bin/bash

rm -rf direktiv-apps
git clone https://github.com/vorteil/direktiv-apps.git
cd direktiv-apps/.direktiv/.templates

fromfiles=""

for i in *.yml; do
	fromfiles="$fromfiles --from-file=$i"
done

kubectl delete cm api-cm-wftemplates || true
kubectl create cm api-cm-wftemplates $fromfiles
kubectl get cm api-cm-wftemplates -o yaml > /tmp/api-cm-wftemplates.yml
kubectl delete cm api-cm-wftemplates || true

cd ../
fromfiles=""

for i in *.json; do
  fromfiles="$fromfiles --from-file=$i"
done

kubectl delete cm api-cm-actiontemplates || true
kubectl create cm api-cm-actiontemplates $fromfiles
kubectl get cm api-cm-actiontemplates -o yaml > /tmp/api-cm-actiontemplates.yml
kubectl delete cm api-cm-actiontemplates || true