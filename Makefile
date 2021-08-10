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
binaries: cmd/direkcli/*.go build/api-binary build/flow-binary build/init-pod-binary build/secrets-binary build/sidecar-binary build/isolates-binary

.PHONY: clean
clean: ## Deletes all build artifacts and tears down existing cluster.
	rm -f build/*.md5
	rm -f build/*.checksum
	rm -f build/*-binary
	rm -f build/api
	rm -f build/flow
	rm -f build/init-pod
	rm -f build/secrets
	rm -f build/sidecar
	if helm status direktiv; then helm uninstall direktiv; fi
	kubectl delete --all ksvc -n direktiv-services-direktiv
	kubectl delete --all jobs -n direktiv-services-direktiv

.PHONY: images
images: image-api image-flow image-init-pod image-secrets image-sidecar image-isolates

.PHONY: push
push: ## Builds all Docker images and pushes them to $DOCKER_REPO.
push: push-api push-flow push-init-pod push-secrets push-sidecar push-isolates push-tls-create

HELM_CONFIG := "scripts/dev.yaml"

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
	go generate ./ent
	go generate ./pkg/secrets/ent/schema

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
	@if [ -d "cmd/$*" ]; then \
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
	@if ! cmp --silent build/$*.md5 build/$*-docker.checksum; then echo "Building docker image for $* binary..." && cd build && docker build -t direktiv-$* -f docker/$*/Dockerfile . ; else echo "Skipping docker build due to unchanged $* binary." && touch build/$*-docker.checksum; fi
	@cp build/$*.md5 build/$*-docker.checksum

.PHONY: image-%
image-%: cmd/direkcli/*.go build/%-docker.checksum
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

.PHONY: template-configmaps
template-configmaps:
	scripts/misc/generate-api-configmaps.sh

cmd/direkcli/*.go:
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

.PHONY: tail-flow
tail-flow: ## Tail logs for currently active 'flow' container.
	$(eval FLOW_RS := $(shell kubectl get rs -o json | jq '.items[] | select(.metadata.labels."app.kubernetes.io/instance" == "direktiv") | .metadata.name'))
	$(eval FLOW_POD := $(shell kubectl get pods -o json | jq '.items[] | select(.metadata.ownerReferences[0].name == ${FLOW_RS}) | .metadata.name'))
	kubectl logs -f ${FLOW_POD} ingress

.PHONY: tail-secrets
tail-secrets: ## Tail logs for currently active 'secrets' container.
	$(eval FLOW_RS := $(shell kubectl get rs -o json | jq '.items[] | select(.metadata.labels."app.kubernetes.io/instance" == "direktiv") | .metadata.name'))
	$(eval FLOW_POD := $(shell kubectl get pods -o json | jq '.items[] | select(.metadata.ownerReferences[0].name == ${FLOW_RS}) | .metadata.name'))
	kubectl logs -f ${FLOW_POD} secrets

.PHONY: tail-api
tail-api: ## Tail logs for currently active 'api' container.
	$(eval API_RS := $(shell kubectl get rs -o json | jq '.items[] | select(.metadata.labels."app.kubernetes.io/instance" == "direktiv-api") | .metadata.name'))
	$(eval API_POD := $(shell kubectl get pods -o json | jq '.items[] | select(.metadata.ownerReferences[0].name == ${API_RS}) | .metadata.name'))
	kubectl logs -f ${API_POD} api

.PHONY: tail-isolates
tail-isolates: ## Tail logs for currently active 'api' container.
	$(eval ISOLATES_RS := $(shell kubectl get rs -o json | jq '.items[] | select(.metadata.labels."app.kubernetes.io/instance" == "direktiv-isolates") | .metadata.name'))
	$(eval ISOLATES_POD := $(shell kubectl get pods -o json | jq '.items[] | select(.metadata.ownerReferences[0].name == ${ISOLATES_RS}) | .metadata.name'))
	kubectl logs -f ${ISOLATES_POD} isolate-controller
