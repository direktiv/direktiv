enterprise = $(if $(ENTERPRISE),$(ENTERPRISE),FALSE)
frontend = $(if $(FRONTEND),$(FRONTEND),frontend)

.PHONY: direktiv-ui-build
direktiv-ui-build:
	DOCKER_BUILDKIT=1 docker build --build-arg IS_ENTERPRISE=${enterprise} --build-arg RELEASE_VERSION=${RELEASE_VERSION} -t ${DOCKER_REPO}/${frontend}:${RELEASE} ui/ 


.PHONY: direktiv-ui
direktiv-ui:
	DOCKER_BUILDKIT=1 docker build --build-arg IS_ENTERPRISE=${enterprise} --build-arg RELEASE_VERSION=${RELEASE_VERSION} -t ${DOCKER_REPO}/${frontend}:${RELEASE} ui/ --push

.PHONY: direktiv-ui-build-cross
direktiv-ui-build-cross:
	@docker buildx create --use --name=direktiv --node=direktiv
	docker buildx build --build-arg IS_ENTERPRISE=${enterprise} --platform linux/amd64,linux/arm64 -t ${DOCKER_REPO}/${frontend}:${RELEASE} ui/ --push
