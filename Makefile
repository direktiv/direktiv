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
.PHONY: run-minio
run-minio:
	if [ ! -f ${mkfile_dir_main}/minio ]; then \
		wget https://apps.vorteil.io/file/vorteil/minio; \
	fi
	vorteil run --vm.disk-size="+2048 MiB" ${mkfile_dir_main}/minio

# run postgres on vorteil
.PHONY: run-postgres
run-postgres:
	if [ ! -f ${mkfile_dir_main}/postgres ]; then \
		wget https://apps.vorteil.io/file/vorteil/postgres; \
	fi
	vorteil run ${mkfile_dir_main}/postgres

# protoc generation
.PHONY: protoc
protoc: $(flow_generated_files) $(health_generated_files) $(ingress_generated_files) $(isolate_generated_files) $(secrets_generated_files)

.PHONY: build
build:
	go get entgo.io/ent
	go generate ./ent
	go generate ./pkg/secrets/ent/schema
	export CGO_LDFLAGS="-static -w -s" && go build -tags osusergo,netgo -o ${mkfile_dir_main}/direktiv cmd/direktiv/main.go

# run as sudo because networking needs root privileges
.PHONY: run
run:
	DIREKTIV_DB="host=192.168.1.10 port=5432 user=sisatech dbname=postgres password=sisatech sslmode=disable" \
	DIREKTIV_SECRETS_DB="host=192.168.1.10 port=5432 user=sisatech dbname=postgres password=sisatech sslmode=disable" go run cmd/direktiv/main.go -d -t wf -c ${mkfile_dir_main}/build/conf.toml

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
