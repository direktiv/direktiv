.PHONY: cluster-create
cluster-create: cluster-image-cache-start ## Creates cluster and requires kind
	mkdir -p ${HOME}/gobase
	chmod 777 ${HOME}/gobase
	cat kind-config.yaml | sed 's?MYDIR?'`pwd`'?' | sed 's?GOBASE?'`echo ${HOME}/gobase`'?' | kind create cluster --config -
# @for node in $(shell kind get nodes --name direktiv-cluster); do \
# 	echo $$node;\
# 	docker exec "$$node" mkdir -p "/etc/containerd/certs.d/localhost:5001"; \
# 	docker exec "$$node" bash -c 'echo [host."http://kind-registry:5000"] > /etc/containerd/certs.d/localhost:5001/hosts.toml'; \
# done
	@if [ "$(shell docker inspect -f='{{json .NetworkSettings.Networks.kind}}' "kind-registry")" = 'null' ]; then \
		docker network connect "kind" "kind-registry"; \
	fi

.PHONY: cluster-delete
cluster-delete: ## Deletes cluster
	kind delete cluster --name direktiv-cluster

.PHONY: cluster-prepare
cluster-prepare: 
	kubectl apply -f kind/postgres.yaml
	kubectl apply -f kind/deploy-ingress-nginx.yaml
	kubectl apply -f kind/svc-configmap.yaml
	kubectl apply -f kind/knative-a-serving-operator.yaml
	kubectl apply -f kind/knative-b-serving-ns.yaml
	kubectl apply -f kind/knative-c-serving-basic.yaml
	kubectl apply -f kind/knative-d-serving-countour.yaml
	kubectl apply -f kind/knative-d-serving-countour.yaml
	kubectl delete -f kind/knative-e-serving-ns-delete.yaml

# mirrord kind
.PHONY: cluster-init 
cluster-init: cluster-create cluster-prepare cluster-build

.PHONY: cluster-ui-build
cluster-ui-build: ## Builds UI for cluster
	DOCKER_BUILDKIT=1 docker build --push -t localhost:5001/frontend:source ui/ 

.PHONY: cluster-build
cluster-build: ## Builds direktiv for cluster
	DOCKER_BUILDKIT=1 docker build --push -f Dockerfile.source -t localhost:5001/direktiv:source .
	DOCKER_BUILDKIT=1 docker build --push -t localhost:5001/direktiv:dev .

.PHONY: cluster-direktiv
cluster-direktiv: ## Installs direktiv in cluster
	helm install --set database.host=postgres.default.svc \
	--set database.port=5432 \
	--set database.user=admin \
	--set database.password=password \
	--set database.name=direktiv \
	--set database.sslmode=disable \
	--set ingress-nginx.install=false \
	--set frontend.image=frontend \
	--set frontend.tag=source \
	--set image=direktiv \
	--set registry=localhost:5001 \
	--set tag=source \
	--set pullPolicy=IfNotPresent \
	--set flow.sidecar=localhost:5001/direktiv:dev \
	--set-json flow.command='[]' \
	--set-json flow.extraVolumes='[{"name":"source-files","hostPath":{"path":"/source"}},{"name":"gobase","hostPath":{"path":"/gobase"}}]' \
	--set-json flow.extraVolumeMounts='[{"name":"source-files","mountPath":"/source"},{"name":"gobase","mountPath":"/gobase"}]' \
	direktiv charts/direktiv

.PHONY: cluster-image-cache-start
cluster-image-cache-start:  
	@if [ "$$(docker ps -f name=proxy-docker-hub --format {{.Names}})" != "proxy-docker-hub" ]; then \
		docker run -d --name proxy-docker-hub --restart=always \
		--net=kind \
		-e REGISTRY_PROXY_REMOTEURL=https://registry-1.docker.io \
		registry:2; \
	fi

	@if [ "$$(docker ps -f name=proxy-quay --format {{.Names}})" != "proxy-quay" ]; then \
		docker run -d --name proxy-quay --restart=always \
		--net=kind \
		-e REGISTRY_PROXY_REMOTEURL=https://quay.io \
		registry:2; \
	fi

	@if [ "$$(docker ps -f name=proxy-gcr --format {{.Names}})" != "proxy-gcr" ]; then \
		docker run -d --name proxy-gcr --restart=always \
		--net=kind \
		-e REGISTRY_PROXY_REMOTEURL=https://gcr.io \
		registry:2; \
	fi

	@if [ "$$(docker ps -f name=proxy-k8s-gcr --format {{.Names}})" != "proxy-k8s-gcr" ]; then \
		docker run -d --name proxy-k8s-gcr --restart=always \
		--net=kind \
		-e REGISTRY_PROXY_REMOTEURL=https://k8s.gcr.io \
		registry:2; \
	fi

	@if [ "$$(docker ps -f name=proxy-registry-k8s-io --format {{.Names}})" != "proxy-registry-k8s-io" ]; then \
		docker run -d --name proxy-registry-k8s-io --restart=always \
		--net=kind \
		-e REGISTRY_PROXY_REMOTEURL=https://registry.k8s.io \
		registry:2; \
	fi

	@if [ "$$(docker ps -f name=proxy-cr-fluentbit-io --format {{.Names}})" != "proxy-cr-fluentbit-io" ]; then \
		docker run -d --name proxy-cr-fluentbit-io --restart=always \
		--net=kind \
		-e REGISTRY_PROXY_REMOTEURL=https://cr.fluentbit.io \
		registry:2; \
	fi

	@if [ "$$(docker ps -f name=kind-registry --format {{.Names}})" != "kind-registry" ]; then \
		docker run -d -p "127.0.0.1:5001:5000" --network bridge --name kind-registry --restart=always \
		registry:2; \
	fi

.PHONY: cluster-image-cache-stop
cluster-image-cache-stop:  
	@docker kill kind-registry proxy-docker-hub proxy-quay proxy-gcr proxy-k8s-gcr proxy-registry-k8s-io proxy-cr-fluentbit-io 2>/dev/null || true
	@docker rm -f kind-registry proxy-docker-hub proxy-quay proxy-gcr proxy-k8s-gcr proxy-registry-k8s-io proxy-cr-fluentbit-io 2>/dev/null || true

.PHONY: cluster-direktiv-run
cluster-direktiv-run:  
	kubectl delete pod -l app.kubernetes.io/name=direktiv,app.kubernetes.io/instance=direktiv 
	kubectl wait --for=condition=ready pod -l "app=direktiv-flow"
	kubectl logs -f -l "app=direktiv-flow"


# sudo sysctl fs.inotify.max_user_watches=524288
# sudo sysctl fs.inotify.max_user_instances=512
# NS=`kubectl get ns |grep Terminating | awk 'NR==1 {print $1}'` && kubectl get namespace "$NS" -o json   | tr -d "\n" | sed "s/\"finalizers\": \[[^]]\+\]/\"finalizers\": []/"   | kubectl replace --raw /api/v1/namespaces/$NS/finalize -f -
	
