#!/bin/bash
# This script install/reinstall simple base stack.

dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"

# check if with docker registry
if [[ $1 == "registry" ]]; then
  $dir/registry/setup.sh
fi

sudo -s source  $dir/resetk3s.sh

export KUBECONFIG=/etc/rancher/k3s/k3s.yaml

kubectl create namespace postgres

helm repo add direktiv https://chart.direktiv.io

if [ ! -d "$dir/direktiv-charts" ]; then
  git clone https://github.com/direktiv/direktiv-charts.git $dir/direktiv-charts;
  git -C $dir/direktiv-charts checkout dev;
fi

cd $dir/direktiv-charts/charts/knative-instance && helm dependency update $dir/direktiv-charts/charts/knative-instance

kubectl apply -f https://github.com/knative/operator/releases/download/knative-v1.9.2/operator.yaml
kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=knative-operator

# knative
kubectl create namespace knative-serving

helm install -n knative-serving knative-serving $dir/direktiv-charts/charts/knative-instance

echo "delete persistent volume claims"
kubectl delete --all -n postgres persistentvolumeclaims
kubectl delete --all -n default persistentvolumeclaims

cd $dir/direktiv-charts/charts/direktiv && helm dependency update $dir/direktiv-charts/charts/direktiv

# install postgres
helm repo add bitnami https://charts.bitnami.com/bitnami
helm install psql-single bitnami/postgresql

echo "waiting for database secret"
while ! kubectl get secret --namespace default psql-single-postgresql
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
  sidecar: "direktiv"
  initPodImage: "init-pod"
EOF
fi

# remove old database setting
sed -i '/database:/,+6 d' $dir/dev.yaml

echo "" >> $dir/dev.yaml
echo "database:
  host: \"psql-single-postgresql\"
  port: 5432
  user: \"postgres\"
  password: \"$(kubectl get secret --namespace default psql-single-postgresql -o jsonpath="{.data.postgres-password}" | base64 -d)\"
  name: \"postgres\"
  sslmode: disable" >> $dir/dev.yaml

echo "generated dev.yaml:\n"

cat $dir/dev.yaml