# SECTION PROTOBUF 
.PHONY: clean-protobuf
clean-protobuf:
	find . -name "*.pb.go" -type f -delete

BUF_VERSION:=1.18.0
.PHONY: protobuf
protobuf: ## Manually regenerates Go packages built from protobuf.
protobuf: clean-protobuf
	docker run -v $$(pwd):/app -w /app bufbuild/buf:$(BUF_VERSION) generate