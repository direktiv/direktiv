#
# Makefile to build direktiv
#

flow_generated_files := $(shell find pkg/flow/ -type f -name '*.proto' -exec sh -c 'echo "{}" | sed "s/\.proto/\.pb.go/"' \;)
health_generated_files := $(shell find pkg/health/ -type f -name '*.proto' -exec sh -c 'echo "{}" | sed "s/\.proto/\.pb.go/"' \;)
ingress_generated_files := $(shell find pkg/ingress/ -type f -name '*.proto' -exec sh -c 'echo "{}" | sed "s/\.proto/\.pb.go/"' \;)
isolate_generated_files := $(shell find pkg/isolate/ -type f -name '*.proto' -exec sh -c 'echo "{}" | sed "s/\.proto/\.pb.go/"' \;)
secrets_generated_files := $(shell find pkg/secrets/ -type f -name '*.proto' -exec sh -c 'echo "{}" | sed "s/\.proto/\.pb.go/"' \;)

.SILENT:

mkfile_path_main := $(abspath $(lastword $(MAKEFILE_LIST)))
mkfile_dir_main := $(dir $(mkfile_path_main))

include build/build.mk

# run minio on vorteil
.PHONY: build-tc-redirect-tap
build-tc-redirect-tap:
	wget https://github.com/awslabs/tc-redirect-tap/archive/master.zip;
	unzip -o master.zip;
	cd tc-redirect-tap-master && GOPATH=/tmp GOCACHE=/tmp make && cp tc-redirect-tap /opt/cni/bin && cd ..
	rm -Rf master.zip && rm -Rf tc-redirect-tap-master

# run minio on vorteil
.PHONY: run-minio
run-minio:
	if [ ! -f ${mkfile_dir_main}/minio ]; then \
		wget https://apps.vorteil.io/file/vorteil/minio; \
	fi
	vorteil run --vm.disk-size="+2048 MiB" ${mkfile_dir_main}/minio

.PHONY: run-minio-docker
run-minio-docker:
	docker run -p 9000:9000 -e MINIO_ACCESS_KEY=vorteil -e MINIO_SECRET_KEY=vorteilvorteil minio/minio server data

# run postgres on vorteil
.PHONY: run-postgres
run-postgres:
	if [ ! -f ${mkfile_dir_main}/postgres ]; then \
		wget https://apps.vorteil.io/file/vorteil/postgres; \
	fi
	vorteil run --vm.ram="2048MiB" --vm.disk-size="+1024MiB" ${mkfile_dir_main}/postgres

.PHONY: run-postgres-docker
run-postgres-docker:
	docker run -p 5432:5432 -e POSTGRES_USER=sisatech -e POSTGRES_PASSWORD=sisatech -e POSTGRES_DB=postgres postgres

# protoc generation
.PHONY: protoc
protoc: $(flow_generated_files) $(health_generated_files) $(ingress_generated_files) $(isolate_generated_files) $(secrets_generated_files)

.PHONY: docker-all
docker-all:
docker-all: build
	cp ${mkfile_dir_main}/direktiv  ${mkfile_dir_main}/build/
	cd build && sudo docker build -t direktiv -f docker/all/Dockerfile .

.PHONY: docker-isolate
docker-isolate:
docker-isolate: build
	cp ${mkfile_dir_main}/direktiv  ${mkfile_dir_main}/build/
	cd build && docker build -t direktiv-isolate -f docker/isolate/Dockerfile .

.PHONY: build
build:
	go get entgo.io/ent
	go generate ./ent
	go generate ./pkg/secrets/ent/schema
	export CGO_LDFLAGS="-static -w -s" && go build -tags osusergo,netgo -o ${mkfile_dir_main}/direktiv cmd/direktiv/main.go

.PHONY: build-cli
build-cli:
	go build -o direkcli cmd/direkcli/main.go

# run e.g. IP=192.168.0.120 make run-isolate-docker
# add -e DIREKTIV_ISOLATION=container for container isolation
.PHONY: run-isolate-docker
run-isolate-docker:
	docker run -p 8888:8888 -e DIREKTIV_ISOLATE_BIND="0.0.0.0:8888" \
	-e DIREKTIV_MINIO_ENDPOINT="$(IP):9000" \
	-e DIREKTIV_DB="host=$(IP) port=5432 user=sisatech dbname=postgres password=sisatech sslmode=disable" \
	--privileged \
	-e DIREKTIV_ISOLATION=container \
	direktiv-isolate /bin/direktiv -t i -d

# run as sudo because networking needs root privileges
.PHONY: run
run:
	DIREKTIV_DB="host=$(DB) port=5432 user=sisatech dbname=postgres password=sisatech sslmode=disable" \
	DIREKTIV_SECRETS_DB="host=$(DB) port=5432 user=sisatech dbname=postgres password=sisatech sslmode=disable" \
	go run cmd/direktiv/main.go -d -t wis -c ${mkfile_dir_main}/build/conf.toml

pkg/secrets/%.pb.go: pkg/secrets/%.proto
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative --experimental_allow_proto3_optional $<

pkg/flow/%.pb.go: pkg/flow/%.proto
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative --experimental_allow_proto3_optional $<

pkg/health/%.pb.go: pkg/health/%.proto
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative --experimental_allow_proto3_optional $<

pkg/ingress/%.pb.go: pkg/ingress/%.proto
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative --experimental_allow_proto3_optional $<

pkg/isolate/%.pb.go: pkg/isolate/%.proto
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative --experimental_allow_proto3_optional $<
