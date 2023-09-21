#!/bin/bash

# -r: reinstall secure registry
# -d <DB>: reinstall database, can be 'pg' or 'operator'
# -l: reinstall linkerd
# -w: do not wipe k3s
# -k: don not install knative

DB=none
LINKERD=false
REGISTRY=false
ARCH=$(dpkg --print-architecture)
WIPE=true
KNATIVE=true

while getopts rd:lwk flag
do
    case "${flag}" in
        r) REGISTRY=true;;
        d) DB=${OPTARG};;
        l) LINKERD=true;;
        w) WIPE=false;;
        k) KNATIVE=false;;
    esac
done

echo WIPE: $WIPE
echo registry: $REGISTRY
echo linkerd: $LINKERD
echo architecture: $ARCH
echo knative: $KNATIVE
echo database: $DB

# get current dir
dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"

# wipe k3s only if wipe is set
if [[ $WIPE == true ]]; then
  sudo -s source  $dir/resetk3s.sh
fi

export KUBECONFIG=/etc/rancher/k3s/k3s.yaml

if [ ! -d "$dir/direktiv-charts" ]; then
  git clone https://github.com/direktiv/direktiv-charts.git $dir/direktiv-charts;
  git -C $dir/direktiv-charts checkout new-port;
fi

install_knative() {

  lc=$(kubectl get pods -n knative-serving | wc -l)

  if [ $lc != 0 ]; then
    echo "already pods in knative-serving"
    return
  fi

  echo "install knative"  
  kubectl apply -f https://github.com/knative/operator/releases/download/knative-v1.11.5/operator.yaml
  kubectl create ns knative-serving
  kubectl apply -f https://raw.githubusercontent.com/direktiv/direktiv/main/kubernetes/install/knative/basic.yaml
  kubectl apply --filename https://github.com/knative/net-contour/releases/download/knative-v1.11.0/contour.yaml
  kubectl delete namespace contour-external --wait=false
}

install_db() {

  # if no database pod running and nnot set we install pg
  lc=$(kubectl get pods -n postgres --field-selector=status.phase!=Terminated | wc -l)

  if [ $lc == 0 ] && [ $DB == "none" ]; then
    DB=pg
  fi

  if [ $DB == "none" ]; then
    return
  fi

  kubectl create namespace postgres

  kubectl delete -f $dir/../kubernetes/install/db/basic.yaml --timeout=60s
  helm uninstall psql-single -n postgres --wait
  helm uninstall postgres -n postgres --wait

  kubectl delete --all -n postgres persistentvolumeclaims

  dbsecret="psql-single-postgresql"

  if [[ ${DB} == "operator" ]]; then
    echo "install postgres db as operator"

helm repo add linkerd https://helm.linkerd.io/stable;
    helm repo add percona https://percona.github.io/percona-helm-charts/
    helm install -n postgres pg-operator percona/pg-operator --wait
    kubectl apply -n postgres -f $dir/../kubernetes/install/db/basic.yaml 
    dbsecret="direktiv-cluster-pguser-direktiv"
  fi

  if [[ ${DB} == "pg" ]]; then
    echo "install pg db"
    helm repo add bitnami https://charts.bitnami.com/bitnami
    helm install -n postgres psql-single bitnami/postgresql
  fi

  echo "waiting for database secret"
  while ! kubectl get secret --namespace postgres $dbsecret
  do
      sleep 2
  done


}

install_linkerd()  {
  echo "setup linkerd"

  helm uninstall linkerd-crds -n linkerd --wait
  helm uninstall linkerd-control-plane -n linkerd --wait
  
  # prepare linkerd
  kubectl annotate ns default linkerd.io/inject=enabled --overwrite

  # tmp dir
  tempDir=$(mktemp -d)
  $(exe='step certificate create root.linkerd.cluster.local ca.crt ca.key \
--profile root-ca --no-password --insecure \
&& step certificate create identity.linkerd.cluster.local issuer.crt issuer.key \
--profile intermediate-ca --not-after 87600h --no-password --insecure \
--ca ca.crt --ca-key ca.key'; \
  sudo docker run --mount "type=bind,src=${tempDir},dst=/home/step"  -i smallstep/step-cli /bin/bash -c "$exe";);
  echo "certificates in $tempDir"

  helm repo add linkerd https://helm.linkerd.io/stable;
  helm install linkerd-crds linkerd/linkerd-crds -n linkerd --create-namespace 
  helm install linkerd-control-plane \
    -n linkerd \
    --set-file identityTrustAnchorsPEM=$tempDir/ca.crt \
    --set-file identity.issuer.tls.crtPEM=$tempDir/issuer.crt \
    --set-file identity.issuer.tls.keyPEM=$tempDir/issuer.key \
    linkerd/linkerd-control-plane --wait
}

if [[ $REGISTRY == true ]]; then
  $dir/registry/setup.sh
fi

if [[ $LINKERD == true ]]; then
  install_linkerd
fi

install_db

if [[ $KNATIVE == true ]]; then
  install_knative
fi


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

frontend:
  image: "ui"
  tag: "latest"
  logging:
    json: false

api:
  image: "direktiv"
  tag: "latest"

functions:
  namespace: direktiv-services-direktiv
  image: "direktiv"
  tag: "latest"
  sidecar: "direktiv"
  initPodImage: "init-pod"
EOF
fi

if [[ $DB == "none" ]]; then
  echo "no database update"
  exit
fi

# remove old database setting
sed -i '/database:/,+6 d' $dir/dev.yaml

echo "" >> $dir/dev.yaml

if [ $DB == "pg" ]; then
  echo "database:
  host: \"psql-single-postgresql.postgres\"
  port: 5432
  user: \"postgres\"
  password: \"$(kubectl get secret --namespace postgres psql-single-postgresql -o jsonpath="{.data.postgres-password}" | base64 -d)\"
  name: \"postgres\"
  sslmode: disable" >> $dir/dev.yaml
else 
  echo "database:
    host: \"$(kubectl get secrets -n postgres direktiv-cluster-pguser-direktiv -o 'go-template={{index .data "pgbouncer-host"}}' | base64 --decode)\"
    port: $(kubectl get secrets -n postgres direktiv-cluster-pguser-direktiv -o 'go-template={{index .data "pgbouncer-port"}}' | base64 --decode)
    user: \"$(kubectl get secrets -n postgres direktiv-cluster-pguser-direktiv -o 'go-template={{index .data "user"}}' | base64 --decode)\"
    password: \"$(kubectl get secrets -n postgres direktiv-cluster-pguser-direktiv -o 'go-template={{index .data "password"}}' | base64 --decode)\"
    name: \"$(kubectl get secrets -n postgres direktiv-cluster-pguser-direktiv -o 'go-template={{index .data "dbname"}}' | base64 --decode)\"
    sslmode: require" >> $dir/dev.yaml
fi

cat $dir/dev.yaml