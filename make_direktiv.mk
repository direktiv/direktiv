.PHONY: direktiv-build
direktiv-build:
	@echo "building direktiv ${RELEASE_VERSION}"
	DOCKER_BUILDKIT=1 docker build --build-arg RELEASE_VERSION=${RELEASE_VERSION} -t ${DOCKER_REPO}/direktiv:${RELEASE} . 

.PHONY: direktiv
direktiv:
	@echo "building and pushing direktiv ${RELEASE_VERSION}"
	DOCKER_BUILDKIT=1 docker build --build-arg RELEASE_VERSION=${RELEASE_VERSION} -t ${DOCKER_REPO}/direktiv:${RELEASE} . --push

.PHONY: direktiv-build-cross
direktiv-build-cross:
	@echo "building cross direktiv ${RELEASE_VERSION}"
	@docker buildx create --use --name=direktiv --node=direktiv
	docker buildx build --build-arg RELEASE_VERSION=${RELEASE_VERSION} --platform linux/amd64,linux/arm64 \
		-t ${DOCKER_REPO}/direktiv:${RELEASE} --push . 


CGO_LDFLAGS := "CGO_LDFLAGS=-static -w -s"
GO_BUILD_TAGS := "osusergo,netgo"

.PHONY: direktiv-cli
direktiv-cli:
	@echo "Building linux cli binary...";
	@export ${CGO_LDFLAGS} && go build -tags ${GO_BUILD_TAGS}  -o direktivctl cmd/exec/main.go
	@tar -czf direktivctl_amd64.tar.gz direktivctl 
	@rm direktivctl

	@echo "Building mac cli binary...";
	@export ${CGO_LDFLAGS} && GOOS=darwin go build -tags ${GO_BUILD_TAGS} -o direktivctl cmd/exec/main.go
	@tar -czf direktivctl_darwin.tar.gz direktivctl 
	@rm direktivctl

	@echo "Building mac cli arm64 binary...";
	@export ${CGO_LDFLAGS} && GOOS=darwin GOARCH=arm64 go build -tags ${GO_BUILD_TAGS} -o direktivctl cmd/exec/main.go
	@tar -czf direktivctl_darwin_arm64.tar.gz direktivctl
	@rm direktivctl

	@echo "Building windows cli binary...";
	@export ${CGO_LDFLAGS} && GOOS=windows go build -tags ${GO_BUILD_TAGS} -o direktivctl.exe cmd/exec/main.go
	@tar -czf direktivctl_windows.tar.gz direktivctl.exe
	@rm direktivctl.exe

.PHONY: direktiv-helm-docs
direktiv-helm-docs: 
	go install github.com/norwoodj/helm-docs/cmd/helm-docs@latest
	helm-docs -c charts