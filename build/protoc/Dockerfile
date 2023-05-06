FROM ubuntu:22.04

RUN apt-get update && apt-get install unzip curl wget -y

ENV PB_REL="https://github.com/protocolbuffers/protobuf/releases"
RUN curl -LO $PB_REL/download/v3.15.8/protoc-3.15.8-linux-x86_64.zip

RUN unzip protoc-3.15.8-linux-x86_64.zip -d /usr/local

RUN wget https://go.dev/dl/go1.20.0.linux-amd64.tar.gz

RUN rm -rf /usr/local/go && tar -C /usr/local -xzf go1.20.0.linux-amd64.tar.gz

ENV PATH="${PATH}:/usr/local/go/bin:/root/go/bin"

RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2

ENTRYPOINT ["protoc"]