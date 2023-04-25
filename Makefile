mkfile_path_main := $(abspath $(lastword $(MAKEFILE_LIST)))
mkfile_dir_main := $(dir $(mkfile_path_main))
docker_repo = $(if $(DOCKER_REPO),$(DOCKER_REPO),localhost:5000)
docker_image = $(if $(DOCKER_IMAGE),$(DOCKER_IMAGE),ui)
docker_tag = $(if $(DOCKER_TAG),:$(DOCKER_TAG),)
GIT_HASH := $(shell git rev-parse --short HEAD)
GIT_DIRTY := $(shell git diff --quiet || echo '-dirty')
RV := ""
RELEASE_TAG = $(shell v='$${RV:+:}$${RV}'; echo "$${v%.*}")
FULL_VERSION := $(shell v='$${RV}$${RV:+-}${GIT_HASH}${GIT_DIRTY}'; echo "$${v%.*}")   


.SECONDARY:

# Build the new server on docker
.PHONY: server
server:
	echo ${RELEASE_TAG}
	docker build . --tag ${docker_repo}/${docker_image}${RELEASE_TAG} --build-arg FULL_VERSION=${FULL_VERSION}
	docker push ${docker_repo}/${docker_image}${RELEASE_TAG}

# Updates remote containers
.PHONY: update-containers
update-containers:
	docker build . --tag direktiv/ui --build-arg FULL_VERSION=${FULL_VERSION}
	docker tag direktiv/ui:latest direktiv/ui${RELEASE_TAG}
	docker push direktiv/ui
	docker push direktiv/ui${RELEASE_TAG}


.PHONY: cross-prepare
cross-prepare:
	docker buildx create --use	
	docker run --privileged --rm docker/binfmt:a7996909642ee92942dcd6cff44b9b95f08dad64
	docker run --rm --privileged multiarch/qemu-user-static --reset -p yes

.PHONY: cross-build
cross-build:
	@if [ "${RELEASE_TAG}" = "" ]; then\
		echo "setting release to dev"; \
		$(eval RELEASE_TAG=dev) \
    fi
	echo "building ${RELEASE}:${RELEASE_TAG}, full version ${FULL_VERSION}"
	rm -Rf app.tar
	docker build -t uibase -f Dockerfile.base .
	container_id=$$(docker create "uibase"); \
	docker cp $$container_id:/app - > app.tar; \
	docker rm -v $$container_id
	tar -xvf app.tar
	docker buildx build --build-arg RELEASE_VERSION=${FULL_VERSION} -f Dockerfile.cross --platform=linux/amd64,linux/arm64 --push -t direktiv/ui:${RELEASE_TAG} .