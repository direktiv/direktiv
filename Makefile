# #
# # Makefile to build direktiv
# #

DOCKER_REPO := "localhost:5000"
CGO_LDFLAGS := "CGO_LDFLAGS=\"-static -w -s\""
GO_BUILD_TAGS := "osusergo,netgo"
GIT_HASH := $(shell git rev-parse --short HEAD)
GIT_DIRTY := $(shell git diff --quiet || echo '-dirty')
RELEASE := ""
RELEASE_TAG = $(shell v='$${RELEASE:+:}$${RELEASE}'; echo "$${v%.*}")
FULL_VERSION := $(shell v='$${RELEASE}$${RELEASE:+-}${GIT_HASH}${GIT_DIRTY}'; echo "$${v%.*}")   

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
	@printf "  %-16s %s\n" '$$GIT_HASH' "${GIT_HASH}"
	@printf "  %-16s %s\n" '$$GIT_DIRTY' "${GIT_DIRTY}"
	@printf "  %-16s %s\n" '$$FULL_VERSION' "${FULL_VERSION}"
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
	rm -f build/functions
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


# Generate API client inside of pkg api
.PHONY: api-client
api-client: ## Generates a golang client to use based off swagger
api-client: api-docs
	swagger generate client -t pkg/api -f scripts/api/swagger.json --name direktivsdk

# API docs

.PHONY: api-docs
api-docs: ## Generates API documentation, (Also fixes markdown tables, examples & description)
api-docs: 
	# go get -u github.com/go-swagger/go-swagger/cmd/swagger
	cd pkg/api
	swagger generate spec -o scripts/api/swagger.json -m
	swagger generate markdown --output scripts/api/api.md -f scripts/api/swagger.json
	echo "Cleanup markdown tables and descriptions"
	sed -i -z 's/#### All responses\n|/#### All responses\n\n|/g' scripts/api/api.md
	sed -i -z 's/description: |//g' scripts/api/api.md
	sed -i -z 's/Example: {/\n**Example**\n!!!!{/g' scripts/api/api.md
	sed -i '/^!!!!{/ s/$$/\n```/' scripts/api/api.md
	sed -i -z 's/!!!!{/```\n{/g' scripts/api/api.md

.PHONY: api-swagger
api-swagger: ## runs swagger server. Use make host=192.168.0.1 api-swagger to change host for API.
api-swagger:
	scripts/api/swagger.sh $(host)

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
		export ${CGO_LDFLAGS} && go build -ldflags "-X github.com/direktiv/direktiv/pkg/version.Version=${FULL_VERSION}" -tags ${GO_BUILD_TAGS} -o $@ cmd/$*/*.go; \
		cp build/$*-binary build/$*; \
	else \
   	touch $@; \
	fi

.PHONY: image-%
image-%: build/%-binary
	cd build && DOCKER_BUILDKIT=1 docker build -t direktiv-$* -f docker/$*/Dockerfile .
	@echo "Make $@: SUCCESS"

.PHONY: push-%
push-%: image-%
	@docker tag direktiv-$* ${DOCKER_REPO}/$*${RELEASE_TAG}
	@docker push ${DOCKER_REPO}/$*${RELEASE_TAG}
	@echo "Make $@${RELEASE_TAG}: SUCCESS"

# UI

.PHONY: docker-ui
docker-ui: ## Manually clone and build the latest UI.
	if [ ! -d direktiv-ui ]; then \
		git clone https://github.com/direktiv/direktiv-ui.git; \
	fi
	if [ -z "${RELEASE}" ]; then \
		cd direktiv-ui && DOCKER_REPO=${DOCKER_REPO} DOCKER_IMAGE=ui make server; \
	else \
		cd direktiv-ui && make update-containers RV=${RELEASE}; \
	fi

# Misc

.PHONY: docker-all
docker-all: ## Build the all-in-one image.
docker-all:
	cp -Rf kubernetes build/docker/all
	cd build/docker/all && ./images.sh
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

.PHONY: reboot-api
reboot-api: ## delete currently active api pod
	kubectl delete pod -l app.kubernetes.io/instance=direktiv-api

.PHONY: reboot-flow
reboot-flow: ## delete currently active flow pod
	kubectl delete pod -l app.kubernetes.io/instance=direktiv

.PHONY: reboot-functions
reboot-functions: ## delete currently active functions pod
	kubectl delete pod -l app.kubernetes.io/instance=direktiv-functions

.PHONY: wait-functions
wait-functions: ## Wait for 'functions' pod to be ready.
	$(eval FUNCTIONS_RS := $(shell kubectl get rs -o json | jq '.items[] | select(.metadata.labels."app.kubernetes.io/instance" == "direktiv-functions") | .metadata.name'))
	$(eval FUNCTIONS_POD := $(shell kubectl get pods -o json | jq '.items[] | select(.metadata.ownerReferences[0].name == ${FUNCTIONS_RS}) | .metadata.name'))
	kubectl wait --for=condition=ready pod ${FUNCTIONS_POD}

.PHONY: wait-flow
wait-flow: ## Wait for 'flow' pod to be ready.
	$(eval FLOW_RS := $(shell kubectl get rs -o json | jq '.items[] | select(.metadata.labels."app.kubernetes.io/name" == "direktiv") | .metadata.name'))
	$(eval FLOW_POD := $(shell kubectl get pods -o json | jq '.items[] | select(.metadata.ownerReferences[0].name == ${FLOW_RS}) | .metadata.name'))
	kubectl wait --for=condition=ready pod ${FLOW_POD}

.PHONY: wait-api
wait-api: ## Wait for 'api' pod to be ready.
	$(eval FLOW_RS := $(shell kubectl get rs -o json | jq '.items[] | select(.metadata.labels."app.kubernetes.io/name" == "direktiv-api") | .metadata.name'))
	$(eval FLOW_POD := $(shell kubectl get pods -o json | jq '.items[] | select(.metadata.ownerReferences[0].name == ${FLOW_RS}) | .metadata.name'))
	kubectl wait --for=condition=ready pod ${FLOW_POD}

.PHONY: upgrade-%
upgrade-%: push-% # Pushes new image deletes, reboots and tail new pod
	@echo "Upgrading $* pod"
	@$(MAKE) reboot-$*
	@$(MAKE) wait-$*
	@$(MAKE) tail-$*
