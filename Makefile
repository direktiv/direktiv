#
# Makefile to build direktiv
#
# Help information is automatically scraped from comments beginning with 
# double-hash. Please add these lines to commands that deserve documentation.

DOCKER_REPO := "localhost:5000"
GIT_HASH := $(shell git rev-parse --short HEAD)
GIT_DIRTY := $(shell git diff --quiet || echo '-dirty')
RELEASE := ""
RELEASE_TAG = $(shell v='$${RELEASE:+:}$${RELEASE}'; echo "$${v%.*}")
FULL_VERSION := $(shell v='$${RELEASE}$${RELEASE:+-}${GIT_HASH}${GIT_DIRTY}'; echo "$${v%.*}")

# used by binary build
CGO_LDFLAGS := "CGO_LDFLAGS=-static -w -s"
GO_BUILD_TAGS := "osusergo,netgo"

.SECONDARY:

.PHONY: help
help: ## Prints usage information.
	@echo "\033[36mMakefile Help\033[0m"
	@echo ""
	@echo "\033[36mVariables\033[0m"
	@printf "  %-16s %s\n" '$$DOCKER_REPO' "${DOCKER_REPO}"
	@printf "  %-16s %s\n" '$$RELEASE' "${RELEASE}"
	@echo ""
	@echo "\033[36mTargets\033[0m"
	@grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-16s %s\n", $$1, $$2}'

#
# Targets for compiling code.
#

.PHONY: clean-protobuf
clean-protobuf:
	find . -name "*.pb.go" -type f -delete

BUF_VERSION:=1.18.0
.PHONY: protobuf
protobuf: ## Manually regenerates Go packages built from protobuf.
protobuf: clean-protobuf
	docker run -v $$(pwd):/app -w /app bufbuild/buf:$(BUF_VERSION) generate

.PHONY: direktiv
direktiv: ## Builds Docker image and pushes it to $DOCKER_REPO.
direktiv:
	DOCKER_BUILDKIT=1 docker build --build-arg RELEASE_VERSION=${FULL_VERSION} -t direktiv -f Dockerfile .
	@docker tag direktiv ${DOCKER_REPO}/direktiv${RELEASE_TAG}
	@docker push ${DOCKER_REPO}/direktiv${RELEASE_TAG}
	@echo "Make $@${RELEASE_TAG}: SUCCESS"

.PHONY: direktivctl
direktivctl: ## Builds the commandline tool.
	@export ${CGO_LDFLAGS} && go build -tags ${GO_BUILD_TAGS} -o direktivctl cmd/exec/main.go

#
# Helper targets for devs and CI/CD tools.
#

.PHONY: godoc
godoc: ## Hosts a godoc server for the project on http port 6060.
	go install golang.org/x/tools/cmd/godoc@latest
	godoc -http=:6060

.PHONY: lint 
lint: VERSION="v1.56"
lint: ## Runs very strict linting on the project.
	docker run \
	--rm \
	-v `pwd`:/app \
	-w /app \
	-e GOLANGCI_LINT_CACHE=/app/.cache/golangci-lint \
	golangci/golangci-lint:${VERSION} golangci-lint run --verbose


.PHONY: license-check 
license-check: ## Scans dependencies looking for blacklisted licenses.
	go install github.com/google/go-licenses@latest
	go-licenses check --ignore=github.com/bbuck/go-lexer,github.com/xi2/xz,modernc.org/mathutil ./... --disallowed_types forbidden,unknown,restricted

TEST_PACKAGES := $(shell find . -type f -name '*_test.go' | sed -e 's/^\.\///g' | sed -r 's|/[^/]+$$||'  |sort |uniq)
UNITTEST_PACKAGES := $(shell echo ${TEST_PACKAGES} | sed 's/ /\n/g' | awk '{print "github.com/direktiv/direktiv/" $$0}')

.PHONY: unittest
unittest: ## Runs all Go unit tests. Or, you can run a specific set of unit tests by defining UNITTEST_PACKAGES relative to the root directory.	
	go test -cover -timeout 60s ${UNITTEST_PACKAGES}

# 
# Targets for running a simple local deployment using docker compose.
#

.PHONY: docker-build
docker-build:
	docker build -t direktiv-dev .

.PHONY: docker-start
docker-start: docker-build docker-stop
docker-start: ## Create a local docker deployment.
	cd ui && docker build -t direktiv-ui-dev .
	DIREKTIV_UI_IMAGE=direktiv-ui-dev DIREKTIV_IMAGE=direktiv-dev  docker compose up -d --scale e2e-api=0

.PHONY: docker-headless
docker-headless: docker-stop docker-stop
docker-headless: ## Create a local docker deployment without an included UI container.
	DIREKTIV_UI_IMAGE=direktiv-ui-dev DIREKTIV_IMAGE=direktiv-dev  docker compose up -d --scale ui=0 --scale e2e-api=0

.PHONY: docker-stop 
docker-stop: ## Stop an existing docker deployment.
	docker rm -f $$(docker ps -q -f "label=direktiv.io/object-type=container") || true
	DIREKTIV_UI_IMAGE=direktiv-ui-dev DIREKTIV_IMAGE=direktiv-dev docker compose down --remove-orphans -v

.PHONY: docker-tail
docker-tail: ## Tail the logs for the direktiv container in the docker deployment.
	DIREKTIV_UI_IMAGE=direktiv-ui-dev DIREKTIV_IMAGE=direktiv-dev  docker compose logs -f

.PHONY: docker-e2e-api
docker-e2e-api: docker-stop docker-build
docker-e2e-api: ## Perform backend end-to-end tests against the docker deployment.
	DIREKTIV_UI_IMAGE=direktiv-ui-dev DIREKTIV_IMAGE=direktiv-dev  docker compose run e2e-api

.PHONY: docker-e2e-playwright
docker-e2e-playwright: docker-build docker-stop
docker-e2e-playwright: ## Create a local docker deployment.
	cd ui && docker build -t direktiv-ui-dev .
	DIREKTIV_UI_IMAGE=direktiv-ui-dev DIREKTIV_IMAGE=direktiv-dev  docker compose run e2e-playwright


# 
# Targets for running a complete k3s local development deployment.
#

.PHONY: k3s-wait 
k3s-wait: 
	kubectl -n direktiv wait --for=condition=ready pod -l "app=direktiv-flow"

.PHONY: k3s-uninstall
k3s-uninstall: ## Uninstall the local development k3s environment.
	./scripts/installer.sh uninstall

.PHONY: k3s-install
k3s-install: k3s-uninstall
k3s-install: ## Install the local development k3s environment.
	DEV=true ./scripts/installer.sh all
	@$(MAKE) k3s-wait

.PHONY: k3s-monitoring-install
k3s-monitoring-install: k3s-uninstall
k3s-monitoring-install: ## Install the local development k3s environment.
	WITH_MONITORING=true DEV=true ./scripts/installer.sh all
	@$(MAKE) k3s-wait

.PHONY: k3s-redeploy
k3s-redeploy: ## Upgrade the local deployment.
	DEV=true ./scripts/installer.sh all
	@$(MAKE) k3s-wait

.PHONY: k3s-reboot 
k3s-reboot: direktiv 
k3s-reboot: ## Recompile the server image and delete the existing pod in the k3s deployment to force an update.
	kubectl -n direktiv delete pod -l app.kubernetes.io/name=direktiv,app.kubernetes.io/instance=direktiv 
	@$(MAKE) k3s-wait

.PHONY: k3s-tail 
k3s-tail: k3s-wait
k3s-tail: ## Tail the logs of the direktiv server running in the local k3s environment.
	kubectl -n direktiv logs -f -l "app=direktiv-flow"

.PHONY: k3s-tests
k3s-tests: k3s-wait
k3s-tests: ## Runs end-to-end tests. DIREKTIV_HOST=128.0.0.1 make test [JEST_PREFIX=/tests/namespaces]
	docker run -it --rm \
	-v `pwd`/tests:/tests \
	-v `pwd`/direktivctl:/bin/direktivctl \
	-e 'DIREKTIV_HOST=${DIREKTIV_HOST}' \
	-e 'NODE_TLS_REJECT_UNAUTHORIZED=0' \
	node:lts-alpine3.18 npm --prefix "/tests" run all -- ${JEST_PREFIX}

# TODO: do we need "make binary"?

# TODO: move this elsewhere
.PHONY: cross
cross:
	@if [ "${RELEASE}" = "" ]; then\
		echo "setting release to dev"; \
		$(eval RELEASE=dev) \
    fi
	@docker buildx create --use --name=direktiv --node=direktiv
	docker buildx build --build-arg RELEASE_VERSION=${FULL_VERSION} -f Dockerfile --platform linux/amd64,linux/arm64 \
		-t ${DOCKER_REPO}/direktiv:${RELEASE} --push .

# TODO: move this elsewhere
.PHONY: scan
scan: 
scan: push
	trivy image --exit-code 1 localhost:5000/direktiv

# TODO: move this elsewhere
.PHONY: cli
cli:
	@echo "Building linux cli binary...";
	@export ${CGO_LDFLAGS} && go build -tags ${GO_BUILD_TAGS}  -o direktivctl cmd/exec/main.go
	@echo "Building mac cli binary...";
	@export ${CGO_LDFLAGS} && GOOS=darwin go build -tags ${GO_BUILD_TAGS} -o direktivctl-darwin cmd/exec/main.go
	@echo "Building mac cli arm64 binary...";
	@export ${CGO_LDFLAGS} && GOOS=darwin GOARCH=arm64 go build -tags ${GO_BUILD_TAGS} -o direktivctl-darwin-arm64 cmd/exec/main.go
	@echo "Building windows cli binary...";
	@export ${CGO_LDFLAGS} && GOOS=windows go build -tags ${GO_BUILD_TAGS} -o direktivctl-windows.exe cmd/exec/main.go


.PHONY: binary
binary: ## Useful only to check that code compiles properly.
	go build -o /dev/null cmd/cmd-exec/*.go
	go build -o /dev/null cmd/direktiv/*.go

.PHONY: helm-docs
helm-docs: 
	go install github.com/norwoodj/helm-docs/cmd/helm-docs@latest
	helm-docs -c charts

.PHONY: openapi
openapi: ## Build and run openapi doc
	make -C openapi run