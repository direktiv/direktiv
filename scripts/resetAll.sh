#!/bin/bash

dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"

sudo -s source  $dir/resetk3s.sh

export KUBECONFIG=/etc/rancher/k3s/k3s.yaml

kubectl create namespace knative-serving
kubectl create namespace direktiv-services-direktiv
kubectl create namespace postgres

# prepare linkerd
kubectl annotate ns knative-serving default direktiv-services-direktiv postgres linkerd.io/inject=enabled

exe='cd /certs && step certificate create root.linkerd.cluster.local ca.crt ca.key \
--profile root-ca --no-password --insecure \
&& step certificate create identity.linkerd.cluster.local issuer.crt issuer.key \
--profile intermediate-ca --not-after 8760h --no-password --insecure \
--ca ca.crt --ca-key ca.key'

mkdir -p $dir/devcerts
dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
rm -Rf $dir/devcerts/*
chmod 777 $dir/devcerts

docker run --user 1000:1000 -v $dir/devcerts:/certs  -i smallstep/step-cli /bin/bash -c "$exe"

helm repo add linkerd https://helm.linkerd.io/stable
exp=$(date -d '+8760 hour' +"%Y-%m-%dT%H:%M:%SZ")

helm install linkerd2 \
  --set-file identityTrustAnchorsPEM=$dir/devcerts/ca.crt \
  --set-file identity.issuer.tls.crtPEM=$dir/devcerts/issuer.crt \
  --set-file identity.issuer.tls.keyPEM=$dir/devcerts/issuer.key \
  --set identity.issuer.crtExpiry=$exp \
  linkerd/linkerd2 --wait

if [ ! -d "$dir/direktiv-charts" ]; then
  git clone git@github.com:direktiv/direktiv-charts.git $dir/direktiv-charts
fi

helm dependency update $dir/direktiv-charts/charts/knative
helm install -n knative-serving knative $dir/direktiv-charts/charts/knative

kubectl delete --all -n postgres persistentvolumeclaims
kubectl delete --all -n default persistentvolumeclaims

helm dependency update $dir/direktiv-charts/charts/direktiv

# install db
helm install -n postgres --set singleNamespace=true postgres $dir/direktiv-charts/charts/pgo --wait
kubectl apply -f $dir/../kubernetes/install/db/pg.yaml

echo "waiting for database secret"
while ! kubectl get secrets -n postgres direktiv-pguser-direktiv
do
    sleep 2
done

echo ""
echo "database:
  host: \"$(kubectl get secrets -n postgres direktiv-pguser-direktiv -o 'go-template={{index .data "host"}}' | base64 --decode)\"
  port: $(kubectl get secrets -n postgres direktiv-pguser-direktiv -o 'go-template={{index .data "port"}}' | base64 --decode)
  user: \"$(kubectl get secrets -n postgres direktiv-pguser-direktiv -o 'go-template={{index .data "user"}}' | base64 --decode)\"
  password: \"$(kubectl get secrets -n postgres direktiv-pguser-direktiv -o 'go-template={{index .data "password"}}' | base64 --decode)\"
  name: \"$(kubectl get secrets -n postgres direktiv-pguser-direktiv -o 'go-template={{index .data "dbname"}}' | base64 --decode)\"
  sslmode: require"
