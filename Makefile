# #
# # Makefile to build direktiv
# #

DOCKER_REPO := "localhost:5000"
CGO_LDFLAGS := "CGO_LDFLAGS=\"-static -w -s\""
GO_BUILD_TAGS := "osusergo,netgo"

.SECONDARY:

.PHONY: help
help: ## Prints usage information.
	@echo "\033[36mMakefile Help\033[0m"
	@echo ""
	@echo "Everything should work out-of-the-box. Just use 'make cluster'."
	@echo ""
	@echo 'If you need to tweak things, make a copy of scripts/dev.yaml and set your $$HELM_CONFIG environment variable to point to it. Ensure that $$DOCKER_REPO matches the registry in your $$HELM_CONFIG file, and that each 'image' in the config file references that same registry.'
	@echo ""
	@echo "\033[36mVariables\033[0m"
	@printf "  %-16s %s\n" '$$DOCKER_REPO' "${DOCKER_REPO}"
	@printf "  %-16s %s\n" '$$HELM_CONFIG' "${HELM_CONFIG}"
	@printf "  %-16s %s\n" '$$REGEX' "${REGEX}"
	@printf "  %-16s %s\n" '$$RELEASE' "${RELEASE}"
	@echo ""
	@echo "\033[36mTargets\033[0m"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-16s %s\n", $$1, $$2}'

.PHONY: binaries
binaries: ## Builds all Direktiv binaries. Useful only to check that code compiles.
binaries: build/flow-binary build/api-binary build/init-pod-binary build/secrets-binary build/sidecar-binary build/functions-binary

.PHONY: clean
clean: ## Deletes all build artifacts and tears down existing cluster.
	rm -f build/*.md5
	rm -f build/*.checksum
	rm -f build/*-binary
	rm -f build/flow
	rm -f build/api
	rm -f build/init-pod
	rm -f build/secrets
	rm -f build/sidecar
	if helm status direktiv; then helm uninstall direktiv; fi
	kubectl delete --all ksvc -n direktiv-services-direktiv
	kubectl delete --all jobs -n direktiv-services-direktiv

.PHONY: images
images: image-api image-flow image-init-pod image-secrets image-sidecar image-functions

.PHONY: push
push: ## Builds all Docker images and pushes them to $DOCKER_REPO.
push: push-api push-flow push-init-pod push-secrets push-sidecar push-functions

HELM_CONFIG := "scripts/dev.yaml"

.PHONY: helm-reinstall
helm-reinstall: ## Re-installes direktiv without pushing images
	if helm status direktiv; then helm uninstall direktiv; fi
	helm install -f ${HELM_CONFIG} direktiv kubernetes/charts/direktiv/

.PHONY: cluster
cluster: ## Updates images at $DOCKER_REPO, then uses $HELM_CONFIG to build the cluster.
cluster: push
	$(eval X := $(shell kubectl get namespaces | grep -c direktiv-services-direktiv))
	if [ ${X} -eq 0 ]; then kubectl create namespace direktiv-services-direktiv; fi
	if helm status direktiv; then helm uninstall direktiv; fi
	kubectl delete -l direktiv.io/scope=w  ksvc -n direktiv-services-direktiv
	kubectl delete --all jobs -n direktiv-services-direktiv
	helm install -f ${HELM_CONFIG} direktiv kubernetes/charts/direktiv/

.PHONY: teardown
teardown: ## Brings down an existing cluster.
	if helm status direktiv; then helm uninstall direktiv; fi
	kubectl delete -l direktiv.io/scope=w ksvc -n direktiv-services-direktiv
	kubectl delete --all jobs -n direktiv-services-direktiv

GO_SOURCE_FILES = $(shell find . -type f -name '*.go' -not -name '*_test.go')
DOCKER_FILES = $(shell find build/docker/ -type f)

# ENT

.PHONY: ent
ent: ## Manually regenerates ent database packages.
	go get entgo.io/ent
	go generate ./pkg/flow/ent
	go generate ./pkg/secrets/ent
	go generate ./pkg/functions/ent


# API docs

.PHONY: api-docs
api-docs: ## Generates API documentation
api-docs:
	go get -u github.com/go-swagger/go-swagger/cmd/swagger
	cd pkg/api
	swagger generate spec -o ./swagger.json
	swagger generate markdown --output ./api.md

# Helm docs

.PHONY: helm-docs
helm-docs: ## Generates helm documentation
helm-docs:
	GO111MODULE=on go get github.com/norwoodj/helm-docs/cmd/helm-docs
	helm-docs kubernetes/charts

# PROTOC

PROTOBUF_SOURCE_FILES := $(shell find . -type f -name '*.proto' -exec sh -c 'echo "{}"' \;)

.PHONY: protoc
protoc: ## Manually regenerates Go packages built from protobuf.
protoc:
	for val in ${PROTOBUF_SOURCE_FILES}; do \
		echo "Generating protobuf file $$val..."; protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative --experimental_allow_proto3_optional $$val; \
	done

# Patterns

build/%-binary: Makefile ${GO_SOURCE_FILES}
	@set -e ; if [ -d "cmd/$*" ]; then \
		echo "Building $* binary..."; \
		export ${CGO_LDFLAGS} && go build -tags ${GO_BUILD_TAGS} -o $@ cmd/$*/*.go; \
		cp build/$*-binary build/$*; \
	else \
   	touch $@; \
	fi

build/%.md5: build/%-binary
	@echo "Calculating md5 checkum of $<..."
	@md5sum $< build/docker/$*/Dockerfile > $@

build/%-docker.checksum: build/%.md5 ${DOCKER_FILES}
	@set -e ; if ! cmp --silent build/$*.md5 build/$*-docker.checksum; then echo "Building docker image for $* binary..." && cd build && docker build -t direktiv-$* -f docker/$*/Dockerfile . ; else echo "Skipping docker build due to unchanged $* binary." && touch build/$*-docker.checksum; fi
	@cp build/$*.md5 build/$*-docker.checksum

.PHONY: image-%
image-%: build/%-docker.checksum
	@echo "Make $@: SUCCESS"

RELEASE := ""
RELEASE_TAG = $(shell v='$${RELEASE:+:}$${RELEASE}'; echo "$${v%.*}")

.PHONY: push-%
push-%: image-%
	@docker tag direktiv-$* ${DOCKER_REPO}/$*${RELEASE_TAG}
	@docker push ${DOCKER_REPO}/$*${RELEASE_TAG}
	@echo "Make $@${RELEASE_TAG}: SUCCESS"

# UI

.PHONY: docker-ui
docker-ui: ## Manually clone and build the latest UI.
	if [ ! -d direktiv-ui ]; then \
		git clone https://github.com/vorteil/direktiv-ui.git; \
	fi
	if [ -z "${RELEASE}" ]; then \
		cd direktiv-ui && DOCKER_REPO=${DOCKER_REPO} DOCKER_IMAGE=direktiv-ui make server; \
	else \
		cd direktiv-ui && make update-containers RV=${RELEASE}; \
	fi

# Misc

.PHONY: docker-all
docker-all: ## Build the all-in-one image.
docker-all:
	cp -Rf kubernetes build/docker/all
	docker build --no-cache -t direktiv-kube build/docker/all

.PHONY: template-configmaps
template-configmaps:
	scripts/misc/generate-api-configmaps.sh

.PHONY: cli
cli:
	@echo "Building linux cli binary...";
	@export ${CGO_LDFLAGS} && go build -tags ${GO_BUILD_TAGS} -o direkcli cmd/direkcli/main.go
	@echo "Building mac cli binary...";
	@export ${CGO_LDFLAGS} && GOOS=darwin go build -tags ${GO_BUILD_TAGS} -o direkcli-darwin cmd/direkcli/main.go
	@echo "Building linux cli binary...";
	@export ${CGO_LDFLAGS} && GOOS=windows go build -tags ${GO_BUILD_TAGS} -o direkcli-windows.exe cmd/direkcli/main.go

# Utility Rules

REGEX := "localhost:5000.*"

.PHONY: purge-images
purge-images: ## Purge images from knative cache by matching $REGEX.
	$(eval IMAGES := $(shell sudo k3s crictl img -o json | jq '.images[] | select (.repoDigests[] | test(${REGEX})) | .id'))
	kubectl delete -l direktiv.io/scope=w  ksvc -n direktiv-services-direktiv
	sudo k3s crictl rmi ${IMAGES}

.PHONY: tail-api
tail-api: ## Tail logs for currently active 'api' container.
	$(eval FLOW_RS := $(shell kubectl get rs -o json | jq '.items[] | select(.metadata.labels."app.kubernetes.io/name" == "direktiv-api") | .metadata.name'))
	$(eval FLOW_POD := $(shell kubectl get pods -o json | jq '.items[] | select(.metadata.ownerReferences[0].name == ${FLOW_RS}) | .metadata.name'))
	kubectl logs -f ${FLOW_POD} api

.PHONY: tail-flow
tail-flow: ## Tail logs for currently active 'flow' container.
	$(eval FLOW_RS := $(shell kubectl get rs -o json | jq '.items[] | select(.metadata.labels."app.kubernetes.io/name" == "direktiv") | .metadata.name'))
	$(eval FLOW_POD := $(shell kubectl get pods -o json | jq '.items[] | select(.metadata.ownerReferences[0].name == ${FLOW_RS}) | .metadata.name'))
	kubectl logs -f ${FLOW_POD} flow

.PHONY: fwd-flow
fwd-flow: ## Tail logs for currently active 'flow' container.
	$(eval FLOW_RS := $(shell kubectl get rs -o json | jq '.items[] | select(.metadata.labels."app.kubernetes.io/name" == "direktiv") | .metadata.name'))
	$(eval FLOW_POD := $(shell kubectl get pods -o json | jq '.items[] | select(.metadata.ownerReferences[0].name == ${FLOW_RS}) | .metadata.name'))
	kubectl port-forward ${FLOW_POD} 8080:6666 --address 0.0.0.0

.PHONY: tail-secrets
tail-secrets: ## Tail logs for currently active 'secrets' container.
	$(eval FLOW_RS := $(shell kubectl get rs -o json | jq '.items[] | select(.metadata.labels."app.kubernetes.io/name" == "direktiv") | .metadata.name'))
	$(eval FLOW_POD := $(shell kubectl get pods -o json | jq '.items[] | select(.metadata.ownerReferences[0].name == ${FLOW_RS}) | .metadata.name'))
	kubectl logs -f ${FLOW_POD} secrets

.PHONY: tail-functions
tail-functions: ## Tail logs for currently active 'functions' container.
	$(eval FUNCTIONS_RS := $(shell kubectl get rs -o json | jq '.items[] | select(.metadata.labels."app.kubernetes.io/instance" == "direktiv-functions") | .metadata.name'))
	$(eval FUNCTIONS_POD := $(shell kubectl get pods -o json | jq '.items[] | select(.metadata.ownerReferences[0].name == ${FUNCTIONS_RS}) | .metadata.name'))
	kubectl logs -f ${FUNCTIONS_POD} functions-controller
