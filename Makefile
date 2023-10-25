# #
# # Makefile to build direktiv
# #

# LD_LIBRARY_PATH := "/usr/local/lib"
DOCKER_REPO := "localhost:5000"
CGO_LDFLAGS := "CGO_LDFLAGS=-static -w -s"
GO_BUILD_TAGS := "osusergo,netgo"
GIT_HASH := $(shell git rev-parse --short HEAD)
GIT_DIRTY := $(shell git diff --quiet || echo '-dirty')
RELEASE := ""
RELEASE_TAG = $(shell v='$${RELEASE:+:}$${RELEASE}'; echo "$${v%.*}")
FULL_VERSION := $(shell v='$${RELEASE}$${RELEASE:+-}${GIT_HASH}${GIT_DIRTY}'; echo "$${v%.*}")
GIT_TAG = $(shell git describe --tags --abbrev=0)
DOCKER_CLONE_REPO = "docker.io/direktiv"

# Set HELM_CONFIG value if environment variable is not set.
HELM_CONFIG ?= "scripts/dev.yaml"


.SECONDARY:

# Clones direktiv image from DOCKER_CLONE_REPO and pushes them to DOCKER_REPO
.PHONY: clone
clone:
	@docker pull ${DOCKER_CLONE_REPO}/direktiv:${GIT_TAG}
	@docker tag ${DOCKER_CLONE_REPO}/direktiv:${GIT_TAG} ${DOCKER_REPO}/direktiv${RELEASE_TAG}
	@docker push ${DOCKER_REPO}/direktiv${RELEASE_TAG}
	@echo "Clone $@${RELEASE_TAG}: SUCCESS"

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

.PHONY: binary
binary: ## Builds all Direktiv binaries. Useful only to check that code compiles.
	go build -o /dev/null cmd/direktiv/*.go

.PHONY: clean
clean: ## Deletes all build artifacts and tears down existing cluster.
	rm -f build/*.md5
	rm -f build/*.checksum
	if helm status direktiv; then helm uninstall direktiv; fi
	kubectl wait --for=delete namespace/direktiv-services-direktiv --timeout=60s
	kubectl delete --all ksvc -n direktiv-services-direktiv
	kubectl delete --all jobs -n direktiv-services-direktiv

.PHONY: helm-reinstall
helm-reinstall: ## Re-installes direktiv without pushing images
	if helm status direktiv; then helm uninstall direktiv; fi
	kubectl wait --for=delete namespace/direktiv-services-direktiv --timeout=60s
	helm install -f ${HELM_CONFIG} direktiv kubernetes/charts/direktiv/

.PHONY: cluster
cluster: ## Updates images at $DOCKER_REPO, then uses $HELM_CONFIG to build the cluster.
cluster: push
	if [ ! -d scripts/direktiv-charts ]; then \
		git clone https://github.com/direktiv/direktiv-charts.git scripts/direktiv-charts; \
		helm dependency build scripts/direktiv-charts/charts/direktiv; \
		helm dependency update scripts/direktiv-charts/charts/direktiv; \
	fi
	if helm status direktiv; then helm uninstall direktiv; fi
	kubectl wait --for=delete namespace/direktiv-services-direktiv --timeout=60s
	kubectl delete -l direktiv.io/scope=w  ksvc -n direktiv-services-direktiv || true
	helm install -f ${HELM_CONFIG} direktiv scripts/direktiv-charts/charts/direktiv/

.PHONY: teardown
teardown: ## Brings down an existing cluster.
	if helm status direktiv; then helm uninstall direktiv --wait; fi
	kubectl delete -l direktiv.io/scope=w ksvc -n direktiv-services-direktiv
	kubectl delete --all jobs -n direktiv-services-direktiv
	kubectl wait --for=delete namespace/direktiv-services-direktiv --timeout=60s

GO_SOURCE_FILES = $(shell find . -type f -name '*.go' -not -name '*_test.go')

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

# # multi-arch build
# .PHONY: cross-prepare
# cross-prepare:
# 	docker buildx create --use      
# 	docker run --privileged --rm docker/binfmt:a7996909642ee92942dcd6cff44b9b95f08dad64
# 	docker run --rm --privileged multiarch/qemu-user-static --reset -p yes

.PHONY: cross
cross:
	@if [ "${RELEASE}" = "" ]; then\
		echo "setting release to dev"; \
		$(eval RELEASE=dev) \
    fi
	@docker buildx create --use --name=direktiv --node=direktiv
	docker buildx build --build-arg RELEASE_VERSION=${FULL_VERSION} -f Dockerfile --platform linux/amd64,linux/arm64 \
		-t ${DOCKER_REPO}/direktiv:${RELEASE} --push .


.PHONY: grpc-clean
grpc-clean: ## Clean all generated grpc files.
grpc-clean:
	rm -rf pkg/*.pb.go
	rm -rf pkg/*/*.pb.go
	rm -rf pkg/*/*/*.pb.go

BUF_VERSION:=1.18.0
.PHONY: grpc-build
grpc-build: ## Manually regenerates Go packages built from protobuf.
grpc-build: grpc-clean
	docker run -v $$(pwd):/app -w /app bufbuild/buf:$(BUF_VERSION) generate

# Patterns

.PHONY: scan
scan: ## Builds and scans all Docker images
scan: push
	trivy image --exit-code 1 localhost:5000/direktiv

.PHONY: image
image:
	DOCKER_BUILDKIT=1 docker build --build-arg RELEASE_VERSION=${FULL_VERSION} -t direktiv -f Dockerfile .
	@echo "Make $@: SUCCESS"

.PHONY: push
push: ## Builds all Docker images and pushes them to $DOCKER_REPO.
push: image
	@docker tag direktiv ${DOCKER_REPO}/direktiv${RELEASE_TAG}
	@docker push ${DOCKER_REPO}/direktiv${RELEASE_TAG}
	@echo "Make $@${RELEASE_TAG}: SUCCESS"

# UI

.PHONY: docker-ui
docker-ui: ## Manually clone and build the latest UI.
	if [ ! -d direktiv-ui ]; then \
		git clone -b develop https://github.com/direktiv/direktiv-ui.git; \
		cd direktiv-ui && make react && make local; \
	fi
	
# Misc

.PHONY: docker-all
docker-all: ## Build the all-in-one image.
docker-all:
	cp -Rf kubernetes build/docker/all
	docker build --no-cache -t direktiv-kube build/docker/all/docker
#cd build/docker/all/multipass && ./generate-init.sh direktiv/direktiv direktiv/ui dev

.PHONY: template-configmaps
template-configmaps:
	scripts/misc/generate-api-configmaps.sh


.PHONY: cli
cli:
	@echo "Building linux cli binary...";
	@export ${CGO_LDFLAGS} && go build -tags ${GO_BUILD_TAGS} -o direktivctl cmd/exec/main.go
	@echo "Building mac cli binary...";
	@export ${CGO_LDFLAGS} && GOOS=darwin go build -tags ${GO_BUILD_TAGS} -o direktivctl-darwin cmd/exec/main.go
	@echo "Building mac cli arm64 binary...";
	@export ${CGO_LDFLAGS} && GOOS=darwin GOARCH=arm64 go build -tags ${GO_BUILD_TAGS} -o direktivctl-darwin-arm64 cmd/exec/main.go
	@echo "Building linux cli binary...";
	@export ${CGO_LDFLAGS} && GOOS=windows go build -tags ${GO_BUILD_TAGS} -o direktivctl-windows.exe cmd/exec/main.go

# Utility Rules

REGEX := "localhost:5000.*"

.PHONY: purge-images
purge-images: ## Purge images from knative cache by matching $REGEX.
	$(eval IMAGES := $(shell sudo k3s crictl img -o json | jq '.images[] | select (.repoDigests[] | test(${REGEX})) | .id'))
	kubectl delete -l direktiv.io/scope=w  ksvc -n direktiv-services-direktiv
	sudo k3s crictl rmi ${IMAGES}

.PHONY: tail-flow
tail-flow: ## Tail logs for currently active 'flow' container.
	$(eval FLOW_RS := $(shell kubectl -n direktiv get rs -o json | jq '.items[] | select(.metadata.labels."app.kubernetes.io/name" == "direktiv") | .metadata.name'))
	$(eval FLOW_POD := $(shell kubectl -n direktiv get pods -o json | jq '.items[] | select(.metadata.ownerReferences[0].name == ${FLOW_RS}) | .metadata.name'))
	kubectl -n direktiv logs -f ${FLOW_POD} flow

.PHONY: fwd-flow
fwd-flow: ## Tail logs for currently active 'flow' container.
	$(eval FLOW_RS := $(shell kubectl -n direktiv get rs -o json | jq '.items[] | select(.metadata.labels."app.kubernetes.io/name" == "direktiv") | .metadata.name'))
	$(eval FLOW_POD := $(shell kubectl -n direktiv get pods -o json | jq '.items[] | select(.metadata.ownerReferences[0].name == ${FLOW_RS}) | .metadata.name'))
	kubectl -n direktiv port-forward ${FLOW_POD} 6666:6666 --address 0.0.0.0

.PHONY: reboot-api
reboot-api: ## delete currently active api pod
	kubectl -n direktiv delete pod -l app.kubernetes.io/instance=direktiv-api

.PHONY: reboot-flow
reboot-flow: ## delete currently active flow pod
	kubectl -n direktiv delete pod -l app.kubernetes.io/name=direktiv,app.kubernetes.io/instance=direktiv 

.PHONY: reboot-functions
reboot-functions: ## delete currently active functions pod
	kubectl -n direktiv delete pod -l app.kubernetes.io/instance=direktiv-functions

.PHONY: wait-flow
wait-flow: ## Wait for 'flow' pod to be ready.
	$(eval FLOW_RS := $(shell kubectl -n direktiv get rs -o json | jq '.items[] | select(.metadata.labels."app.kubernetes.io/name" == "direktiv") | .metadata.name'))
	$(eval FLOW_POD := $(shell kubectl -n direktiv get pods -o json | jq '.items[] | select(.metadata.ownerReferences[0].name == ${FLOW_RS}) | .metadata.name'))
	kubectl -n direktiv wait --for=condition=ready pod ${FLOW_POD}

.PHONY: upgrade-%
upgrade-%: push-% ## Pushes new image deletes, reboots and tail new pod
	@echo "Upgrading $* pod"
	@$(MAKE) reboot-$*
	@$(MAKE) wait-$*
	@$(MAKE) tail-$*

.PHONY: upgrade
upgrade: push ## Pushes all images and reboots flow, function, and api pods
	@$(MAKE) reboot-flow

.PHONY: dependencies
dependencies: ## installs tools 
	go install github.com/google/go-licenses@latest


.PHONY: license-check 
license-check: ## Scans dependencies looking for licenses.
	go-licenses check --ignore=github.com/bbuck/go-lexer,github.com/xi2/xz,modernc.org/mathutil ./... --disallowed_types forbidden,unknown,restricted

TEST_PACKAGES := $(shell find . -type f -name '*_test.go' | sed -e 's/^\.\///g' | sed -r 's|/[^/]+$$||'  |sort |uniq)
UNITTEST_PACKAGES = $(shell echo ${TEST_PACKAGES} | sed 's/ /\n/g' | awk '{print "github.com/direktiv/direktiv/" $$0}')

.PHONY: unittest
unittest: ## Runs all Go unit tests. Or, you can run a specific set of unit tests by defining TEST_PACKAGES relative to the root directory.
	go test -cover -timeout 60s ${UNITTEST_PACKAGES}

.PHONY: lint 
lint: VERSION="v1.54"
lint: ## Runs very strict linting on the project.
	docker run \
	--rm \
	--name golangci-lint-${VERSION}-direktiv \
	-v `pwd`:/app \
	-w /app \
	golangci/golangci-lint:${VERSION} golangci-lint run

.PHONY: test
test: ## Runs end-to-end tests. DIREKTIV_HOST=128.0.0.1 make test [JEST_PREFIX=/tests/namespaces]
	docker run -it --rm \
	-v `pwd`/tests:/tests \
	-v `pwd`/direktivctl:/bin/direktivctl \
	-e 'DIREKTIV_HOST=${DIREKTIV_HOST}' \
	-e 'NODE_TLS_REJECT_UNAUTHORIZED=0' \
	node:alpine npm --prefix "/tests" run all -- ${JEST_PREFIX}

server-godoc:
	go install golang.org/x/tools/cmd/godoc@latest
	godoc -http=:6060


env-stop:
	DIREKTIV_IMAGE=direktiv-dev docker compose down --remove-orphans -v

env-start: env-stop
	rm -rf direktiv-ui
	git clone -b develop https://github.com/direktiv/direktiv-ui.git
	cd direktiv-ui && docker build -t direktiv-ui-dev .

	docker build -t direktiv-dev .
	DIREKTIV_IMAGE=direktiv-dev  docker compose up -d --scale e2e-tests=0
	DIREKTIV_IMAGE=direktiv-dev  docker compose logs -f

env-start-headless: env-stop
	docker build -t direktiv-dev .
	DIREKTIV_IMAGE=direktiv-dev  docker compose up -d --scale ui=0 --scale e2e-tests=0
	DIREKTIV_IMAGE=direktiv-dev  docker compose logs -f

e2e-tests: env-stop
	DOCKER_BUILDKIT=1 docker build -t direktiv-dev .
	DOCKER_BUILDKIT=1 DIREKTIV_IMAGE=direktiv-dev  docker compose run e2e-tests
