#
# Makefile to build direktiv
#
flow_generated_files := $(shell find pkg/flow/ -type f -name '*.proto' -exec sh -c 'echo "{}" | sed "s/\.proto/\.pb.go/"' \;)
health_generated_files := $(shell find pkg/health/ -type f -name '*.proto' -exec sh -c 'echo "{}" | sed "s/\.proto/\.pb.go/"' \;)
ingress_generated_files := $(shell find pkg/ingress/ -type f -name '*.proto' -exec sh -c 'echo "{}" | sed "s/\.proto/\.pb.go/"' \;)
secrets_generated_files := $(shell find pkg/secrets/grpc -type f -name '*.proto' -exec sh -c 'echo "{}" | sed "s/\.proto/\.pb.go/"' \;)
hasYarn := $(shell which yarn)

.SILENT:

mkfile_path_main := $(abspath $(lastword $(MAKEFILE_LIST)))
mkfile_dir_main := $(dir $(mkfile_path_main))

# run postgres on vorteil
.PHONY: run-postgres
run-postgres:
	if [ ! -f ${mkfile_dir_main}/postgres ]; then \
		wget https://apps.vorteil.io/file/vorteil/postgres; \
	fi
	vorteil run --vm.ram="2048MiB" --vm.disk-size="+1024MiB" ${mkfile_dir_main}/postgres

# protoc generation
.PHONY: protoc
protoc: $(flow_generated_files) $(health_generated_files) $(ingress_generated_files) $(secrets_generated_files)


.PHONY: docker-secrets
docker-secrets:
docker-secrets: build
	cp ${mkfile_dir_main}/secrets  ${mkfile_dir_main}/build/
	cd build && docker build -t direktiv-secrets -f docker/secrets/Dockerfile .

.PHONY: docker-all
docker-all:
	docker build -t direktiv-kube ${mkfile_dir_main}/build/docker/all

.PHONY: docker-api
docker-api:
docker-api: build
	cp ${mkfile_dir_main}/api  ${mkfile_dir_main}/build/
	cd build && docker build -t direktiv-api -f docker/api/Dockerfile .

.PHONY: docker-flow
docker-flow:
docker-flow: build
	cp ${mkfile_dir_main}/direktiv  ${mkfile_dir_main}/build/
	cd build && docker build -t direktiv-flow -f docker/flow/Dockerfile .

.PHONY: docker-cli
docker-cli:
docker-cli: build
		cp ${mkfile_dir_main}/direkcli  ${mkfile_dir_main}/build/
		cd build && docker build -t direktiv-cli -f docker/cli/Dockerfile .

.PHONY: docker-sidecar
docker-sidecar:
	export CGO_LDFLAGS="-static -w -s" && go build -tags osusergo,netgo -o ${mkfile_dir_main}/build/docker/sidecar/sidecar cmd/sidecar/main.go
	docker build -t sidecar  ${mkfile_dir_main}/build/docker/sidecar/

.PHONY: build
build:
	go get entgo.io/ent
	go generate ./ent
	go generate ./pkg/secrets/ent/schema
	export CGO_LDFLAGS="-static -w -s" && go build -tags osusergo,netgo -o ${mkfile_dir_main}/direktiv cmd/direktiv/main.go
	export CGO_LDFLAGS="-static -w -s" && go build -tags osusergo,netgo -o ${mkfile_dir_main}/secrets cmd/secrets/main.go
	export CGO_LDFLAGS="-static -w -s" && go build -tags osusergo,netgo -o ${mkfile_dir_main}/api cmd/api/main.go
	export CGO_LDFLAGS="-static -w -s" && go build -tags osusergo,netgo -o direkcli cmd/direkcli/main.go
	cp ${mkfile_dir_main}/direktiv  ${mkfile_dir_main}/build/
	cp ${mkfile_dir_main}/api ${mkfile_dir_main}/build/

# run as sudo because networking needs root privileges
.PHONY: run
run:
	DIREKTIV_DB="host=$(DB) port=5432 user=sisatech dbname=postgres password=sisatech sslmode=disable" \
	DIREKTIV_SECRETS_DB="host=$(DB) port=5432 user=sisatech dbname=postgres password=sisatech sslmode=disable" \
	DIREKTIV_INSTANCE_LOGGING_DRIVER="database" \
	DIREKTIV_MOCKUP=1 \
	go run cmd/direktiv/main.go -d

pkg/secrets/%.pb.go: pkg/secrets/%.proto
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative --experimental_allow_proto3_optional $<

pkg/flow/%.pb.go: pkg/flow/%.proto
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative --experimental_allow_proto3_optional $<

pkg/health/%.pb.go: pkg/health/%.proto
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative --experimental_allow_proto3_optional $<

pkg/ingress/%.pb.go: pkg/ingress/%.proto
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative --experimental_allow_proto3_optional $<
