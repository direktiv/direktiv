KIND_CONFIG ?= kind-config.yaml

.PHONY: cluster-setup
cluster-setup: cluster-create cluster-prep cluster-direktiv

.PHONY: cluster-create
cluster-create: 
	kind delete clusters --all
	kind create cluster --config ${KIND_CONFIG}

	if ! docker inspect proxy-docker-hub >/dev/null 2>&1; then \
		docker run -d --name proxy-docker-hub --restart=always \
		--net=kind \
		-e REGISTRY_PROXY_REMOTEURL=https://registry-1.docker.io \
		registry:2;\
	fi

	if ! docker inspect kind-registry >/dev/null 2>&1; then \
		docker run -d -p "127.0.0.1:5001:5000" --network bridge --name kind-registry --restart=always registry:2; \
		docker network connect kind kind-registry; \
	fi

	DOCKER_BUILDKIT=1 docker build --push -t localhost:5001/direktiv:dev .

	if ! docker inspect proxy-quay >/dev/null 2>&1; then \
		docker run -d --name proxy-quay --restart=always \
		--net=kind \
		-e REGISTRY_PROXY_REMOTEURL=https://quay.io \
		registry:2;\
	fi

	if ! docker inspect proxy-gcr >/dev/null 2>&1; then \
		docker run -d --name proxy-gcr --restart=always \
		--net=kind \
		-e REGISTRY_PROXY_REMOTEURL=https://gcr.io \
		registry:2;\
	fi

	if ! docker inspect proxy-k8s-gcr >/dev/null 2>&1; then \
		docker run -d --name proxy-k8s-gcr --restart=always \
		--net=kind \
		-e REGISTRY_PROXY_REMOTEURL=https://k8s.gcr.io \
		registry:2;\
	fi

	if ! docker inspect proxy-registry-k8s-io >/dev/null 2>&1; then \
		docker run -d --name proxy-registry-k8s-io --restart=always \
		--net=kind \
		-e REGISTRY_PROXY_REMOTEURL=https://registry.k8s.io \
		registry:2;\
	fi

	if ! docker inspect proxy-cr-fluentbit-io >/dev/null 2>&1; then \
		docker run -d --name proxy-cr-fluentbit-io --restart=always \
		--net=kind \
		-e REGISTRY_PROXY_REMOTEURL=https://cr.fluentbit.io \
		registry:2;\
	fi

.PHONY: cluster-prep
cluster-prep: 
	kubectl apply -f kind/postgres.yaml
	kubectl apply -f kind/deploy-ingress-nginx.yaml
	kubectl apply -f kind/svc-configmap.yaml
	kubectl apply -f kind/knative-a-serving-operator.yaml
	kubectl apply -f kind/knative-b-serving-ns.yaml
	kubectl apply -f kind/knative-c-serving-basic.yaml
	kubectl apply -f kind/knative-d-serving-countour.yaml
	kubectl apply -f kind/knative-d-serving-countour.yaml
	kubectl delete -f kind/knative-e-serving-ns-delete.yaml

.PHONY: cluster-build
cluster-build: ## Builds direktiv for cluster
	DOCKER_BUILDKIT=1 docker build --push -t localhost:5001/direktiv:dev .

.PHONY: cluster-direktiv-delete
cluster-direktiv-delete: ## Deletes direktiv from cluster
	kubectl get namespace "direktiv-services-direktiv" -o json   | tr -d "\n" | sed "s/\"finalizers\": \[[^]]\+\]/\"finalizers\": []/"   | kubectl replace --raw /api/v1/namespaces/direktiv-services-direktiv/finalize -f - || true
	helm uninstall direktiv

.PHONY: cluster-direktiv
cluster-direktiv: ## Installs direktiv in cluster
	kubectl wait -n ingress-nginx --for=condition=ready pod --selector=app.kubernetes.io/component=controller --timeout=120s
	kubectl wait -n ingress-nginx --for=condition=complete job --selector=app.kubernetes.io/component=admission-webhook --timeout=120s
	helm install --set database.host=postgres.default.svc \
	--set database.port=5432 \
	--set database.user=admin \
	--set database.password=password \
	--set database.name=direktiv \
	--set database.sslmode=disable \
	--set ingress-nginx.install=false \
	--set image=direktiv \
	--set registry=localhost:5001 \
	--set tag=dev \
	--set pullPolicy=IfNotPresent \
	--set flow.sidecar=localhost:5001/direktiv:dev \
	direktiv charts/direktiv

	kubectl wait --for=condition=ready pod -l app=direktiv-flow --timeout=60s

	@echo "Waiting for API endpoint to return 200..."
	@until curl -s -o /dev/null -w "%{http_code}" http://127.0.0.1:9090/api/v2/status | grep -q 200; do \
		echo "Waiting..."; \
		sleep 2; \
	done
	@echo "Endpoint is ready!"

.PHONY: cluster-image-cache-stop
cluster-image-cache-stop:  
	@docker kill kind-registry proxy-docker-hub proxy-quay proxy-gcr proxy-k8s-gcr proxy-registry-k8s-io proxy-cr-fluentbit-io 2>/dev/null || true
	@docker rm -f kind-registry proxy-docker-hub proxy-quay proxy-gcr proxy-k8s-gcr proxy-registry-k8s-io proxy-cr-fluentbit-io 2>/dev/null || true

.PHONY: cluster-direktiv-run
cluster-direktiv-run: cluster-build
	kubectl delete pod -l app.kubernetes.io/name=direktiv,app.kubernetes.io/instance=direktiv 
	kubectl wait --for=condition=ready pod -l "app=direktiv-flow"
	kubectl logs -f -l "app=direktiv-flow"