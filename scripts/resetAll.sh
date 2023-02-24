#!/bin/bash


dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"

# check if with docker registry
if [[ $1 == "registry" ]]; then
  $dir/registry/setup.sh
fi

sudo -s source  $dir/resetk3s.sh

export KUBECONFIG=/etc/rancher/k3s/k3s.yaml

kubectl create namespace postgres

# prepare linkerd
kubectl annotate ns default linkerd.io/inject=enabled

certDir=$(exe='step certificate create root.linkerd.cluster.local ca.crt ca.key \
--profile root-ca --no-password --insecure \
&& step certificate create identity.linkerd.cluster.local issuer.crt issuer.key \
--profile intermediate-ca --not-after 87600h --no-password --insecure \
--ca ca.crt --ca-key ca.key'; \
  sudo docker run --mount "type=bind,src=$(pwd),dst=/home/step"  -i smallstep/step-cli /bin/bash -c "$exe";  \
echo $(pwd));

helm repo add linkerd https://helm.linkerd.io/stable;
helm repo add direktiv https://chart.direktiv.io

helm install linkerd-crds linkerd/linkerd-crds -n linkerd --create-namespace 

helm install linkerd-control-plane \
  -n linkerd \
  --set-file identityTrustAnchorsPEM=$certDir/ca.crt \
  --set-file identity.issuer.tls.crtPEM=$certDir/issuer.crt \
  --set-file identity.issuer.tls.keyPEM=$certDir/issuer.key \
  linkerd/linkerd-control-plane --wait

if [ ! -d "$dir/direktiv-charts" ]; then
  git clone https://github.com/direktiv/direktiv-charts.git $dir/direktiv-charts
fi

cd $dir/direktiv-charts/charts/knative-instance && helm dependency update $dir/direktiv-charts/charts/knative-instance

kubectl apply -f https://github.com/knative/operator/releases/download/knative-v1.8.1/operator.yaml 
kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=knative-operator

CACERT=$dir/registry/share/out/ca.cert.pem


# knative
kubectl create namespace knative-serving
#kubectl annotate ns knative-serving linkerd.io/inject=enabled

echo "checking for ca $CACERT"
if test -f "$CACERT"; then
  echo "using ca-cert"
  kubectl create secret generic self-signed-registry -n knative-serving --from-file=ca.crt=$CACERT
  helm install -n knative-serving --set certificate=self-signed-registry knative-serving direktiv/knative-instance
else
  echo "not using ca-cert"
  helm install -n knative-serving knative-serving direktiv/knative-instance
fi

kubectl delete --all -n postgres persistentvolumeclaims
kubectl delete --all -n default persistentvolumeclaims

cd  $dir/direktiv-charts/charts/direktiv && helm dependency update $dir/direktiv-charts/charts/direktiv

# install db
helm install -n postgres --set singleNamespace=true postgres $dir/direktiv-charts/charts/pgo --wait
kubectl apply -f $dir/../kubernetes/install/db/basic.yaml

echo "waiting for database secret"
while ! kubectl get secrets -n postgres direktiv-pguser-direktiv
do
    sleep 2
done


if [ ! -f "$dir/dev.yaml" ]; then
cat <<EOF > $dir/dev.yaml
registry: localhost:5000
pullPolicy: Always
debug: "true"

secrets:
  image: "direktiv"
  tag: "latest"

flow:
  image: "direktiv"
  dbimage: "direktiv"
  tag: "latest"

ui:
  image: "ui"
  tag: "latest"

api:
  image: "direktiv"
  tag: "latest"

functions:
  namespace: direktiv-services-direktiv
  image: "direktiv"
  tag: "latest"
  sidecar: "sidecar"
  initPodImage: "init-pod"
EOF
fi

# remove old database setting
sed -i '/database:/,+6 d' $dir/dev.yaml

echo "" >> $dir/dev.yaml
echo "database:
  host: \"$(kubectl get secrets -n postgres direktiv-pguser-direktiv -o 'go-template={{index .data "host"}}' | base64 --decode)\"
  port: $(kubectl get secrets -n postgres direktiv-pguser-direktiv -o 'go-template={{index .data "port"}}' | base64 --decode)
  user: \"$(kubectl get secrets -n postgres direktiv-pguser-direktiv -o 'go-template={{index .data "user"}}' | base64 --decode)\"
  password: \"$(kubectl get secrets -n postgres direktiv-pguser-direktiv -o 'go-template={{index .data "password"}}' | base64 --decode)\"
  name: \"$(kubectl get secrets -n postgres direktiv-pguser-direktiv -o 'go-template={{index .data "dbname"}}' | base64 --decode)\"
  sslmode: require" >> $dir/dev.yaml

cat $dir/dev.yaml