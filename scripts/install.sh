#!/bin/sh

set -e

check_binaries() 
{
    if ! which kind > /dev/null; then
        echo "kind not installed. please visit: https://kind.sigs.k8s.io/docs/user/quick-start/"
        exit 1
    fi 
    if ! which kubectl > /dev/null; then
        echo "kubectl not installed. please visit: https://kubernetes.io/docs/tasks/tools/"
        exit 1
    fi 
    if ! which helm > /dev/null; then
        echo "helm not installed. please visit: https://helm.sh/docs/intro/install/"
        exit 1
    fi 
}

create_cluster() 
{
    kind delete clusters direktiv
    (
cat << EOF
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
name: direktiv-cluster
nodes:
- role: control-plane
  extraPortMappings:
  - containerPort: 80
    hostPort: 9090
    protocol: TCP
EOF
    ) | kind create cluster --config -
}

install_dependencies() 
{
    kubectl apply -f https://raw.githubusercontent.com/direktiv/direktiv/refs/heads/main/kind/postgres.yaml
	kubectl apply -f https://raw.githubusercontent.com/direktiv/direktiv/refs/heads/main/kind/deploy-ingress-nginx.yaml
	kubectl apply -f https://raw.githubusercontent.com/direktiv/direktiv/refs/heads/main/kind/svc-configmap.yaml
	kubectl apply -f https://raw.githubusercontent.com/direktiv/direktiv/refs/heads/main/kind/knative-a-serving-operator.yaml
	kubectl apply -f https://raw.githubusercontent.com/direktiv/direktiv/refs/heads/main/kind/knative-b-serving-ns.yaml
	kubectl apply -f https://raw.githubusercontent.com/direktiv/direktiv/refs/heads/main/kind/knative-c-serving-basic.yaml
	kubectl apply -f https://raw.githubusercontent.com/direktiv/direktiv/refs/heads/main/kind/knative-d-serving-countour.yaml
	kubectl apply -f https://raw.githubusercontent.com/direktiv/direktiv/refs/heads/main/kind/knative-d-serving-countour.yaml
	kubectl delete -f https://raw.githubusercontent.com/direktiv/direktiv/refs/heads/main/kind/knative-e-serving-ns-delete.yaml
}

install_direktiv() 
{
    echo "waiting for nginx ingress controller"
    kubectl wait -n ingress-nginx --for=condition=ready pod --selector=app.kubernetes.io/component=controller --timeout=180s
	kubectl wait -n ingress-nginx --for=condition=complete job --selector=app.kubernetes.io/component=admission-webhook --timeout=180s
	helm install --set database.host=postgres.default.svc \
	--set database.port=5432 \
	--set database.user=admin \
	--set database.password=password \
	--set database.name=direktiv \
	--set database.sslmode=disable \
	--set ingress-nginx.install=false \
	--set pullPolicy=IfNotPresent \
	direktiv direktiv/direktiv

	kubectl wait --for=condition=ready pod -l app=direktiv-flow --timeout=60s
}

check_binaries
create_cluster
install_dependencies
install_direktiv

echo "access direktiv at http://localhost:9090"