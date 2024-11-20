#!/bin/bash

set -e

# Default cluster name if not provided
DEFAULT_CLUSTER_NAME="kind"
DEFAULT_CLUSTER_NAME="kind"
CLUSTER_NAME="${1:-$DEFAULT_CLUSTER_NAME}"
DEFAULT_CIDR="172.22.0.240/28"
CIRD="${2:-$DEFAULT_CIDR}"
#CIRD=$(docker network inspect kind | jq -r '.[0].IPAM.Config[0].Subnet' 2>/dev/null)
CLUSTER_NAME="${1:-$DEFAULT_CLUSTER_NAME}"

echo "Cleaning old cluster $CLUSTER_NAME if any..."
kind delete cluster --name "$CLUSTER_NAME" || true

# Variables
KIND_CONFIG="/tmp/kind-config.yaml"
METALLB_CONFIG="/tmp/metallb-config.yaml"
POSTGRES_INGRESS="/tmp/postgres-ingress.yaml"
REGISTRY="/tmp/registry-deployment.yaml"

touch $KIND_CONFIG
touch $METALLB_CONFIG
touch $POSTGRES_INGRESS
# Create Kind Configuration
cat <<EOF > $KIND_CONFIG
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
  - role: control-plane
EOF

# Create Kind Cluster
echo "Creating Kind cluster $CLUSTER_NAME..."
kind create cluster --name "$CLUSTER_NAME" --config $KIND_CONFIG
# Set kubectl context to the new cluster
sleep 1
kubectl cluster-info --context kind-$CLUSTER_NAME

# Preloading images (after cluster creation)
echo Preloading images

# List of images to preload
IMAGES=(
  "quay.io/metallb/controller:v0.14.8"
  "nginx:latest"
  "bitnami/postgresql:latest"
)

# Pull and load each image into the Kind cluster
for IMAGE in "${IMAGES[@]}"; do
  echo "Pulling image: $IMAGE"
  docker pull $IMAGE
  echo "Loading image: $IMAGE into the Kind cluster"
  kind load docker-image $IMAGE --name $CLUSTER_NAME
done

# Install MetalLB
echo "Installing MetalLB..."
kubectl apply -f https://raw.githubusercontent.com/metallb/metallb/v0.14.8/config/manifests/metallb-native.yaml

# Wait for MetalLB components to be ready
sleep 3
echo "Waiting for MetalLB to be ready..."
kubectl wait --namespace metallb-system --for=condition=Ready pod -l app=metallb --timeout=120s

# Create the MetalLB configuration file dynamically
cat <<EOF > $METALLB_CONFIG
apiVersion: metallb.io/v1beta1
kind: IPAddressPool
metadata:
  name: kind-pool
  namespace: metallb-system
spec:
  addresses:
  - "$CIRD"
---
apiVersion: metallb.io/v1beta1
kind: L2Advertisement
metadata:
  name: kind-l2
  namespace: metallb-system
spec: {}
EOF

# Apply MetalLB Configuration
kubectl apply -f $METALLB_CONFIG

# Install Knative
echo "Installing Knative..."
kubectl create namespace knative-serving || true
kubectl apply -f https://github.com/knative/operator/releases/download/knative-v1.12.2/operator.yaml
kubectl apply -f https://raw.githubusercontent.com/direktiv/direktiv/main/scripts/kubernetes/install/knative/basic.yaml
kubectl apply -f https://github.com/knative/net-contour/releases/download/knative-v1.11.1/contour.yaml
kubectl delete namespace contour-external || true
kubectl create namespace postgres || true

# Install PostgreSQL
helm repo add bitnami https://charts.bitnami.com/bitnami
helm install postgres bitnami/postgresql \
  --namespace postgres \
  --set primary.persistence.enabled=false

# Wait for PostgreSQL to be ready
kubectl wait --namespace postgres --for=condition=Ready pod --selector=app.kubernetes.io/name=postgresql --timeout=120s

# Create PostgreSQL Ingress
echo "Creating Ingress for PostgreSQL..."
cat <<EOF > $POSTGRES_INGRESS
apiVersion: v1
kind: Service
metadata:
  name: postgres
  namespace: postgres
spec:
  selector:
    app.kubernetes.io/name: postgresql
  ports:
    - protocol: TCP
      port: 5432
      targetPort: 5432
  type: LoadBalancer
EOF

# Apply PostgreSQL Ingress
kubectl apply -f $POSTGRES_INGRESS

# Apply PostgreSQL Ingress
kubectl apply -f $POSTGRES_INGRESS

# Deploy the Docker registry
kubectl create namespace registry
cat <<EOF > $REGISTRY
apiVersion: apps/v1
kind: Deployment
metadata:
  name: registry
  namespace: registry
spec:
  replicas: 1
  selector:
    matchLabels:
      app: registry
  template:
    metadata:
      labels:
        app: registry
    spec:
      containers:
        - name: registry
          image: registry:2
          ports:
            - containerPort: 5000
          volumeMounts:
            - name: registry-data
              mountPath: /var/lib/registry
      volumes:
        - name: registry-data
          emptyDir: {}  # Non-persistent volume
---
apiVersion: v1
kind: Service
metadata:
  name: registry
  namespace: registry
spec:
  type: LoadBalancer
  ports:
    - port: 5000
      targetPort: 5000
  selector:
    app: registry
EOF

# Apply Registry Deployment and Service
kubectl apply -f $REGISTRY

# Wait for registry pod to be ready
echo "Waiting for registry pod to be ready..."
kubectl wait --namespace registry --for=condition=Ready pod -l app=registry --timeout=120s

export REGISTRY_HOST=(`kubectl --namespace registry get services registry --output jsonpath='{.status.loadBalancer.ingress[0].ip}'`)

echo your image registry is here: $REGISTRY_HOST

echo "cluster is ready!"
