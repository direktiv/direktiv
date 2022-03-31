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
