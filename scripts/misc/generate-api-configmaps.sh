#!/bin/bash

PWD2=$(pwd)

rm -rf direktiv-apps
git clone https://github.com/vorteil/direktiv-apps.git
cd direktiv-apps/.direktiv/.templates

fromfiles=""

for i in *.yml; do
	fromfiles="$fromfiles --from-file=$i"
done

kubectl delete cm api-wftemplates-cm || true
kubectl create cm api-wftemplates-cm  $fromfiles
kubectl get cm api-wftemplates-cm  -o yaml > $PWD2/kubernetes/charts/direktiv/templates/api-wftemplates-cm.yaml
sed -i 's/  namespace: default/  namespace: {{ .Release.Namespace }}/g' $PWD2/kubernetes/charts/direktiv/templates/api-wftemplates-cm.yaml
kubectl delete cm api-wftemplates-cm  || true

cd ../
fromfiles=""

for i in *.json; do
  fromfiles="$fromfiles --from-file=$i"
done

kubectl delete cm api-actiontemplates-cm|| true
kubectl create cm api-actiontemplates-cm $fromfiles
kubectl get cm api-actiontemplates-cm -o yaml > $PWD2/kubernetes/charts/direktiv/templates/api-actiontemplates-cm.yaml
sed -i 's/  namespace: default/  namespace: {{ .Release.Namespace }}/g' $PWD2/kubernetes/charts/direktiv/templates/api-actiontemplates-cm.yaml
kubectl delete cm api-actiontemplates-cm || true
