KIND_CONFIG ?= kind-config.yaml
TELEMETRY ?= false
LOKI ?= 

.PHONY: cluster-setup
cluster-setup: cluster-create cluster-prep cluster-direktiv

.PHONY: cluster-setup-ee
cluster-setup-ee:
	make cluster-setup IS_ENTERPRISE=true

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

	DOCKER_BUILDKIT=1 docker build --build-arg IS_ENTERPRISE=${IS_ENTERPRISE} --push -t localhost:5001/direktiv:dev .

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

	@if [ "$(IS_ENTERPRISE)" != "true" ]; then \
		helm install --set database.host=postgres.default.svc \
		--set database.port=5432 \
		--set database.user=admin \
		--set database.password=password \
		--set database.name=direktiv \
		--set database.sslmode=disable \
		--set pullPolicy=Always \
		--set ingress-nginx.install=false \
		--set image=direktiv \
		--set registry=localhost:5001 \
		--set tag=dev \
		--set otel.install=${TELEMETRY} $(LOKI) \
		direktiv charts/direktiv; \
	fi

	@if [ "$(IS_ENTERPRISE)" = "true" ]; then \
	helm install --set database.host=postgres.default.svc \
	-f direktiv-ee/install/05_direktiv/keys.yaml \
	--set database.port=5432 \
	--set database.user=admin \
	--set database.password=password \
	--set database.name=direktiv \
	--set database.sslmode=disable \
	--set pullPolicy=Always \
	--set ingress-nginx.install=false \
	--set image=direktiv \
	--set registry=localhost:5001 \
	--set tag=dev \
	--set flow.additionalEnvs[0].name=DIREKTIV_OIDC_ADMIN_GROUP \
	--set flow.additionalEnvs[0].value="admin" \
	--set flow.additionalEnvs[1].name=DIREKTIV_OIDC_DEV \
	--set flow.additionalEnvs[1].value=true \
	--set otel.install=${TELEMETRY} \
	direktiv charts/direktiv; \
	fi

	kubectl wait --for=condition=ready pod -l app=direktiv-flow --timeout=60s

	@if [ "$(IS_ENTERPRISE)" = "true" ]; then \
	@echo "Installing Dex"; \
	helm repo add dex https://charts.dexidp.io; \
	helm repo update; \
	helm install dex dex/dex -f kind/dex-values.yaml; \
	fi

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

cluster-dev:
	DOCKER_BUILDKIT=1 docker build --build-arg IS_ENTERPRISE=${IS_ENTERPRISE} --push -t localhost:5001/direktiv:dev .
	kubectl delete pod -l app=direktiv-flow

cluster-setup-tracing:
	$(MAKE) cluster-setup TELEMETRY=true LOKI="-f scripts/telemetry/telemetry.yaml"	
	helm repo add grafana https://grafana.github.io/helm-charts
	helm repo update
	helm uninstall tempo || true
	helm install tempo grafana/tempo
	helm uninstall grafana || true
	helm install \
	--set ingress.enabled=false \
	--set ingress.path=/grafana \
	--set adminPassword=admin \
	--set ingress.hosts={} \
	--set ingress.ingressClassName=nginx \
	--set datasources."datasources\.yaml".apiVersion=1 \
	--set datasources."datasources\.yaml".datasources[0].name=loki \
	--set datasources."datasources\.yaml".datasources[0].type=loki \
	--set datasources."datasources\.yaml".datasources[0].access=proxy \
	--set datasources."datasources\.yaml".datasources[0].url=http://loki-gateway \
	--set datasources."datasources\.yaml".datasources[1].name=tempo \
	--set datasources."datasources\.yaml".datasources[1].type=tempo \
	--set datasources."datasources\.yaml".datasources[1].access=proxy \
	--set datasources."datasources\.yaml".datasources[1].url=http://tempo:3100 \
	--set service.nodePort=31788 \
	--set service.type=NodePort \
	grafana grafana/grafana
	helm uninstall loki || true
	helm install \
	--set loki.useTestSchema=true \
	--set loki.commonConfig.replication_factor=1 \
	--set loki.pattern_ingester.enabled=true \
	--set loki.limits_config.allow_structured_metadata=true \
	--set loki.limits_config.volume_enabled=true \
	--set deploymentMode=SingleBinary \
	--set singleBinary.replicas=1 \
	--set loki.ruler.enable_api=true \
	--set backend.replicas=0 \
	--set read.replicas=0 \
	--set write.replicas=0 \
	--set ingester.replicas=0 \
	--set querier.replicas=0 \
	--set queryFrontend.replicas=0 \
	--set queryScheduler.replicas=0 \
	--set ingdistributorester.replicas=0 \
	--set indexGateway.replicas=0 \
	--set bloomCompactor.replicas=0 \
	--set bloomGateway.replicas=0 \
	--set minio.enabled=true \
	--set loki.auth_enabled=false \
	loki grafana/loki
