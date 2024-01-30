#!/bin/bash

export $(cat /env | xargs)

# start k3s and wait for pod to be up
curl -sfL https://get.k3s.io | sh -s - --disable traefik --write-kubeconfig-mode=644
echo "waiting for k3s"

while ! kubectl wait --for=condition=ready -n kube-system pod -l k8s-app=metrics-server
do
    echo "waiting for k3s pods"
    sleep 1
done

# knative
kubectl apply -f https://github.com/knative/operator/releases/download/knative-v1.12.2/operator.yaml
kubectl create ns knative-serving

if [ -n "${HTTPS_PROXY+1}" ]; then
  echo "adding proxy to knative"
  curl https://raw.githubusercontent.com/direktiv/direktiv/main/scripts/kubernetes/install/knative/basic.yaml > knative.yaml

  ln=$(awk '/name: controller/ {print FNR}' knative.yaml)
  num=$((ln + 1))

  VAR=$(cat <<-END
    env:
      - container: controller
        envVars:
        - name: HTTP_PROXY
          value: "${HTTP_PROXY}"
        - name: HTTPS_PROXY
          value: "${HTTP_PROXY}"
        - name: NO_PROXY
          value: "${NO_PROXY}"
END
)

    ed -v knative.yaml <<END
${num}i
${VAR}
.
w
q
END

  kubectl apply -f knative.yaml
else
  kubectl apply -f https://raw.githubusercontent.com/direktiv/direktiv/main/scripts/kubernetes/install/knative/basic.yaml
fi

# database
mkdir pv
cat <<EOF > pv/pv.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: postgres
spec:
  replicas: 1
  selector:
    matchLabels:
      app: postgres
  template:
    metadata:
      labels:
        app: postgres
    spec:
      volumes:
        - name: postgres-pv-storage
          persistentVolumeClaim:
            claimName: postgres-pv-claim
      containers:
        - name: postgres
          image: postgres:13.4
          imagePullPolicy: IfNotPresent
          ports:
            - containerPort: 5432
          env:
            - name: POSTGRES_USER
              value: direktiv
            - name: POSTGRES_DB
              value: direktiv
            - name: POSTGRES_PASSWORD
              value: direktivdirektiv
            - name: PGDATA
              value: /var/lib/postgresql/data/pgdata
          volumeMounts:
            - mountPath: /var/lib/postgresql/data
              name: postgres-pv-storage
EOF

cat <<EOF > pv/storage.yaml
apiVersion: v1
kind: PersistentVolume
metadata:
  name: postgres-pv-volume
spec:
  storageClassName: local-path
  capacity:
    storage: 1Gi
  accessModes:
    - ReadWriteOnce
  hostPath:
    path: "/tmp/pgdata"
---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: postgres-pv-claim
spec:
  storageClassName: local-path
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 1Gi
---
apiVersion: v1
kind: Service
metadata:
  name: postgres
  labels:
    app: postgres
spec:
  type: ClusterIP
  ports:
   - port: 5432
  selector:
    app: postgres
EOF

kubectl apply -f pv

# docker registry
mkdir registry

cat <<EOF > registry/registry.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: registry-ns
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: registry-sa
  namespace: registry-ns
---
apiVersion: v1
kind: Pod
metadata:
  name: docker-registry-pod
  namespace: registry-ns
  labels:
    app: registry
spec:
  serviceAccountName: registry-sa
  containers:
  - name: registry
    image: registry:2.7.1
---
apiVersion: v1
kind: Service
metadata:
  name: docker-registry
  namespace: registry-ns
spec:
  type: NodePort
  selector:
    app: registry
  ports:
  - port: 5000
    targetPort: 5000
    nodePort: 31212
EOF

kubectl apply -f registry

# waiting for db
while ! kubectl wait --for=condition=ready pod -l app=postgres
do
    echo "waiting for database"
    sleep 1
done


# direktiv
cat <<EOF > direktiv.yaml
pullPolicy: IfNotPresent
debug: "true"

eventing:
  enabled: true

$(if [ -n "${HTTPS_PROXY+1}" ]; then
  echo "  http_proxy: ${HTTP_PROXY}"
  echo "  https_proxy: ${HTTPS_PROXY}"
  echo "  no_proxy: ${NO_PROXY}"
fi)

database:
  # -- database host
  host: "postgres"
  # -- database port
  port: 5432
  # -- database user
  user: "direktiv"
  # -- database password
  password: "direktivdirektiv"
  # -- database name, auto created if it does not exist
  name: "direktiv"
  # -- sslmode for database
  sslmode: disable

$(if [ -n "${HTTPS_PROXY+1}" ]; then
  echo "http_proxy: ${HTTP_PROXY}"
  echo "https_proxy: ${HTTPS_PROXY}"
  echo "no_proxy: ${NO_PROXY}"
fi)

$(if [ -n "${APIKEY+1}" ]; then
  echo "apikey: \"${APIKEY}\""
fi)
EOF

helm repo add direktiv https://charts.direktiv.io && \
helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx && \
helm repo add prometheus https://prometheus-community.github.io/helm-charts && \
KUBECONFIG=/etc/rancher/k3s/k3s.yaml helm install -f /direktiv.yaml direktiv direktiv/direktiv

# eventing
kubectl create ns knative-eventing

cat <<EOF > eventing.yaml
apiVersion: operator.knative.dev/v1beta1
kind: KnativeEventing
metadata:
  name: knative-eventing
  namespace: knative-eventing
EOF

kubectl apply -f eventing.yaml

# contour
kubectl apply --filename https://github.com/knative/net-contour/releases/download/knative-v1.12.2/contour.yaml
kubectl delete ns contour-external &

# waiting for direktiv
while ! kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=direktiv
do
    echo "waiting for direktiv"
    sleep 1
done

n=0
until [ "$n" -ge 24 ]
do
   curl -X PUT http://localhost/api/namespaces/examples \
     -H "Content-Type: application/json" \
     -H "Direktiv-Token: ${APIKEY}" \
     --write-out "%{http_code}" \
     -f \
     -d '{ "url": "https://github.com/direktiv/direktiv-examples.git", "ref": "main" }' && break
   n=$((n+1))
   sleep 10
done



