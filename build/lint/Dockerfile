FROM ubuntu:22.04 as libgit

ENV DEBIAN_FRONTEND=noninteractive
RUN apt-get clean && apt-get update

RUN apt-get install libmbedtls-dev git cmake build-essential pkg-config libssl-dev \
         python3 zlib1g-dev libmbedtls-dev libpcre3 libpcre3-dev wget -y

# build libssh
RUN wget https://www.libssh2.org/download/libssh2-1.10.0.tar.gz && tar -xvzf libssh2-1.10.0.tar.gz
RUN cd libssh2-1.10.0 && mkdir bin && cd bin && cmake .. -DENABLE_ZLIB_COMPRESSION=ON -DBUILD_SHARED_LIBS=ON && cmake --build . --target install

RUN git clone -b v1.3.0 https://github.com/libgit2/libgit2.git 

# build libgit
RUN mkdir libgit2/build && cd libgit2/build
RUN cd libgit2/build && cmake .. -DUSE_SSH=ON
RUN cd libgit2/build && cmake --build . --target install

FROM libgit as golang

ENV GOPATH=/root

RUN apt-get update && apt-get install curl -y
RUN wget https://go.dev/dl/go1.19.3.linux-amd64.tar.gz
RUN tar -C /usr/local -xzf go1.19.3.linux-amd64.tar.gz

ENV PATH=$PATH:/usr/local/go/bin

RUN /usr/local/go/bin/go version

COPY --from=libgit /usr/local/lib/libssh2.so.1.0.1 /lib/x86_64-linux-gnu/libssh2.so.1.0.1
RUN ln -s /lib/x86_64-linux-gnu/libssh2.so.1.0.1 /lib/x86_64-linux-gnu/libssh2.so.1
RUN ln -s /lib/x86_64-linux-gnu/libssh2.so.1 /lib/x86_64-linux-gnu/libssh2.so

COPY --from=libgit /usr/local/lib/libgit2.so.1.3.0 /lib/x86_64-linux-gnu/
RUN ln -s /lib/x86_64-linux-gnu/libgit2.so.1.3.0 /lib/x86_64-linux-gnu/libgit2.so.1.3
RUN ln -s /lib/x86_64-linux-gnu/libgit2.so.1.3 /lib/x86_64-linux-gnu/libgit2.so

RUN curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.50.1
ENV PATH=$PATH:/root/bin

RUN go install github.com/google/go-licenses@latest

RUN git config --global --add safe.directory /direktiv
WORKDIR /direktiv

CMD ["make", "lint"]