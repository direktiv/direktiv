#!/bin/bash

dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"

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

echo "stopping k3s"

service k3s stop

echo "deleting k3s data"

rm -Rf /etc/rancher/k3s
rm -Rf /var/lib/rancher/k3s
rm -rf /var/lib/cni/networks/cbr0

for name in $(ip -o link show | awk -F': ' '{print $2}' | sed  's/@.*//' | grep veth)
do
    r=`ip link show $name | grep cni0`
    if [ "$r" != "" ]; then
      echo "deleting $name"
      ip link delete $name
    fi
done

echo "starting k3s"

service k3s start

countdown

export KUBECONFIG=/etc/rancher/k3s/k3s.yaml

kubectl create namespace knative-serving
kubectl create namespace direktiv-services-direktiv

# prepare linkerd
kubectl annotate ns knative-serving default direktiv-services-direktiv linkerd.io/inject=enabled

exe='cd /certs && step certificate create root.linkerd.cluster.local ca.crt ca.key \
--profile root-ca --no-password --insecure \
&& step certificate create identity.linkerd.cluster.local issuer.crt issuer.key \
--profile intermediate-ca --not-after 8760h --no-password --insecure \
--ca ca.crt --ca-key ca.key'

dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
rm -Rf $dir/devcerts/*
chmod 777 $dir/devcerts

docker run -v $dir/devcerts:/certs  -i smallstep/step-cli /bin/bash -c "$exe"

helm repo add linkerd https://helm.linkerd.io/stable
exp=$(date -d '+8760 hour' +"%Y-%m-%dT%H:%M:%SZ")

helm install linkerd2 \
  --set-file identityTrustAnchorsPEM=$dir/devcerts/ca.crt \
  --set-file identity.issuer.tls.crtPEM=$dir/devcerts/issuer.crt \
  --set-file identity.issuer.tls.keyPEM=$dir/devcerts/issuer.key \
  --set identity.issuer.crtExpiry=$exp \
  linkerd/linkerd2 --wait

helm dependency update $dir/../kubernetes/charts/knative
helm install -n knative-serving -f $dir/../kubernetes/charts/knative/debug-knative.yaml  knative $dir/../kubernetes/charts/knative

# install database
# delete stuff first
kubectl delete --all -n yugabyte persistentvolumeclaims
kubectl delete  job.batch/yugabyte-setup-credentials -n yugabyte

kubectl create namespace yugabyte
helm repo add yugabytedb https://charts.yugabyte.com

helm install -n yugabyte -f $dir/../kubernetes/install/db/tiny.yaml yugabyte yugabytedb/yugabyte

helm dependency update $dir/../kubernetes/charts/direktiv

countdown
