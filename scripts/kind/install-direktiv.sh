#!/bin/bash

set -e

# Default cluster name if not provided
DEFAULT_CLUSTER_NAME="kind"
CLUSTER_NAME="${1:-$DEFAULT_CLUSTER_NAME}"

DIREKTIV_VERSION="v0.8.9"

# Preloading images (after cluster creation)
echo Preloading images

echo "Tag Direktiv images as dev"
docker pull direktiv/frontend:$DIREKTIV_VERSION
docker tag direktiv/frontend:$DIREKTIV_VERSION direktiv/frontend:dev

docker build -t direktiv/direktiv:$DIREKTIV_VERSION .
docker tag direktiv/direktiv:$DIREKTIV_VERSION direktiv/direktiv:dev

echo "Load images into Kind cluster"
kind load docker-image direktiv/frontend:dev --name $CLUSTER_NAME
kind load docker-image direktiv/direktiv:dev --name $CLUSTER_NAME
echo "Images tagged as 'dev' and loaded into Kind."

echo "Setting up Direktiv in Kind cluster with name $CLUSTER_NAME..."
# Set kubectl context to the new cluster
kubectl cluster-info --context kind-$CLUSTER_NAME

# Retrieve PostgreSQL password
POSTGRES_PASSWORD=$(kubectl get secret --namespace postgres postgres-postgresql -o jsonpath="{.data.postgres-password}" | base64 -d)
echo "PostgreSQL password retrieved: $POSTGRES_PASSWORD"

DIREKTIV_CONFIG="/tmp/direktiv.yaml"
# Create Direktiv Configuration
cat <<EOF > $DIREKTIV_CONFIG
pullPolicy: Never
tag: dev
flow:
  debug: true
database:
  host: "postgres-postgresql.postgres.svc"
  port: 5432
  user: "postgres"
  password: "$POSTGRES_PASSWORD"
  name: "postgres"
  sslmode: disable

frontend:
  image: "direktiv/frontend"
  tag: dev
nats:
  install: true
EOF

# Install Direktiv
echo "Installing Direktiv..."
helm repo add direktiv https://charts.direktiv.io
helm upgrade -i direktiv -f $DIREKTIV_CONFIG ./charts/direktiv
export DIREKTIV_HOST=(`kubectl get services direktiv-ingress-nginx-controller --output jsonpath='{.status.loadBalancer.ingress[0].ip}'`)
echo Direktiv is at http://$DIREKTIV_HOST
echo "setup completed!"
