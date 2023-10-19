# mkfile_path_main := $(abspath $(lastword $(MAKEFILE_LIST)))
# mkfile_dir_main := $(dir $(mkfile_path_main))
docker_repo = $(if $(DOCKER_REPO),$(DOCKER_REPO),localhost:5000)
docker_image = $(if $(DOCKER_IMAGE),$(DOCKER_IMAGE),frontend)
docker_tag = $(if $(DOCKER_TAG),$(DOCKER_TAG),dev)
enterprise = $(if $(DOCKER_TAG),$(DOCKER_TAG),FALSE)
# GIT_HASH := $(shell git rev-parse --short HEAD)
# GIT_DIRTY := $(shell git diff --quiet || echo '-dirty')
# RV := ""
# RELEASE_TAG = $(shell v='$${RV:+:}$${RV}'; echo "$${v%.*}")
# FULL_VERSION := $(shell v='$${RV}$${RV:+-}${GIT_HASH}${GIT_DIRTY}'; echo "$${v%.*}")   


DOCKERFILE_REACT=Dockerfile.base
DOCKERFILE_SERVER=Dockerfile.frontend

# .SECONDARY:

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

# this builds the ui files and copies it from the container to dist/
# used for cross-compilation but can be used locally as well
.PHONY: react
react:
	rm -Rf app.tar
	rm -Rf dist/
	docker build -t uibase --build-arg IS_ENTERPRISE=${enterprise}  -f ${DOCKERFILE_REACT} .
	container_id=$$(docker create "uibase"); \
	docker cp $$container_id:/app/dist - > app.tar; \
	docker rm -v $$container_id
	tar -xvf app.tar
	rm -Rf app.tar

# local container build
.PHONY: local
local:
	docker build -t ${docker_repo}/${docker_image}:${docker_tag} -f ${DOCKERFILE_SERVER} .
	docker tag ${docker_repo}/${docker_image}:${docker_tag} ${docker_repo}/${docker_image}
	docker push ${docker_repo}/${docker_image}:${docker_tag}
	docker push ${docker_repo}/${docker_image}

.PHONY: cross
cross:
	@docker buildx create --use --name=direktiv --node=direktiv
	docker buildx build --platform linux/amd64,linux/arm64 -f ${DOCKERFILE_SERVER} \
		-t ${docker_repo}/${docker_image}:${docker_tag} --push .


.PHONY: forward-api
forward-api: 
	kubectl port-forward svc/direktiv-api 7755:1644

# requires forward-api to run in different console
.PHONY: run-container
run-container: 
	docker run --network host -e DIREKTIV_SERVER_APIKEY=helloworld -e DIREKTIV_SERVER_BACKEND=http://127.0.0.1:7755 -p 2304:2304 localhost:5000/frontend:dev