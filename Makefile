# #
# # Makefile to build direktiv
# #

DOCKER_REPO := "localhost:5000"
CGO_LDFLAGS := "CGO_LDFLAGS=\"-static -w -s\""
GO_BUILD_TAGS := "osusergo,netgo"

.SECONDARY:

.PHONY: binaries
binaries: build/api-binary build/flow-binary build/init-pod-binary build/secrets-binary build/sidecar-binary

.PHONY: clean 
clean:
	rm -f build/*.md5
	rm -f build/*.checksum 
	rm -f build/*-binary 
	rm -f build/api
	rm -f build/flow 
	rm -f build/init-pod 
	rm -f build/secrets 
	rm -f build/sidecar
	if helm status direktiv; then helm uninstall direktiv; fi
	kubectl delete --all ksvc
	kubectl delete --all jobs

.PHONY: images 
images: image-api image-flow image-init-pod image-secrets image-sidecar

.PHONY: push 
push: push-api push-flow push-init-pod push-secrets push-sidecar

.PHONE: cluster 
cluster: push 
	if helm status direktiv; then helm uninstall direktiv; fi
	kubectl delete --all ksvc
	kubectl delete --all jobs
	helm install -f ${HELM_CONFIG} direktiv kubernetes/charts/direktiv/

GO_SOURCE_FILES = $(shell find . -type f -name '*.go' -not -name '*_test.go') 
DOCKER_FILES = $(shell find build/docker/ -type f)

# ENT 

.PHONY: ent
ent:
	go get entgo.io/ent
	go generate ./ent
	go generate ./pkg/secrets/ent/schema


# PROTOC 

PROTOBUF_SOURCE_FILES := $(shell find . -type f -name '*.proto' -exec sh -c 'echo "{}" | sed "s/\.proto/\.pb.go/"' \;)

pkg/%.pb.go: pkg/%.proto
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative --experimental_allow_proto3_optional $<

.PHONY: protoc
protoc: ${PROTOBUF_SOURCE_FILES}

# Patterns 

build/%-binary: Makefile ${GO_SOURCE_FILES}
	@echo "Building $* binary..."
	@export ${CGO_LDFLAGS} && go build -tags ${GO_BUILD_TAGS} -o $@ cmd/$*/*.go
	@cp build/$*-binary build/$*

build/%.md5: build/%-binary
	@echo "Calculating md5 checkum of $<..."
	@md5sum $< | cut -d" " -f1 > $@

build/%-docker.checksum: build/%.md5 ${DOCKER_FILES}
	@if ! cmp --silent build/$*.md5 build/$*-docker.checksum; then echo "Building docker image for $* binary..." && cd build && docker build -t direktiv-$* -f docker/$*/Dockerfile . ; else echo "Skipping docker build due to unchanged $* binary." && touch build/$*-docker.checksum; fi
	@cp build/$*.md5 build/$*-docker.checksum

.PHONY: image-%
image-%: build/%-docker.checksum
	@echo "Make $@: SUCCESS"

.PHONY: push-% 
push-%: image-%
	@docker tag direktiv-$* ${DOCKER_REPO}/$*
	@docker push ${DOCKER_REPO}/$*
	@echo "Make $@: SUCCESS"

# UI  

.PHONY: docker-ui
docker-ui:
	if [ ! -d ${mkfile_dir_main}direktiv-ui ]; then \
		git clone https://github.com/vorteil/direktiv-ui.git; \
	fi
	cd direktiv-ui && make update-containers

# Utility Rules 

REGEX := "localhost:5000.*"

.PHONY: purge-images
purge-images:
	$(eval IMAGES := $(shell sudo k3s crictl img -o json | jq '.images[] | select (.repoDigests[] | test(${REGEX})) | .id'))
	kubectl delete --all ksvc
	sudo k3s crictl rmi ${IMAGES}

.PHONY: logs-flow
logs-flow:
	$(eval FLOW_RS := $(shell kubectl get rs -o json | jq '.items[] | select(.metadata.labels."app.kubernetes.io/instance" == "direktiv") | .metadata.name'))
	$(eval FLOW_POD := $(shell kubectl get pods -o json | jq '.items[] | select(.metadata.ownerReferences[0].name == ${FLOW_RS}) | .metadata.name'))
	kubectl logs -f ${FLOW_POD} ingress

.PHONY: logs-secrets
logs-secrets:
	$(eval FLOW_RS := $(shell kubectl get rs -o json | jq '.items[] | select(.metadata.labels."app.kubernetes.io/instance" == "direktiv") | .metadata.name'))
	$(eval FLOW_POD := $(shell kubectl get pods -o json | jq '.items[] | select(.metadata.ownerReferences[0].name == ${FLOW_RS}) | .metadata.name'))
	kubectl logs -f ${FLOW_POD} secrets

.PHONY: logs-api
logs-api:
	$(eval API_RS := $(shell kubectl get rs -o json | jq '.items[] | select(.metadata.labels."app.kubernetes.io/instance" == "direktiv-api") | .metadata.name'))
	$(eval API_POD := $(shell kubectl get pods -o json | jq '.items[] | select(.metadata.ownerReferences[0].name == ${API_RS}) | .metadata.name'))
	kubectl logs -f ${API_POD} api
