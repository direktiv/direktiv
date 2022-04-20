FROM ubuntu:21.04 as clibs

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get clean && apt-get update
RUN apt-get install git cmake build-essential -y
RUN git clone https://github.com/libgit2/libgit2.git
RUN apt-get install pkg-config libssl-dev \
        python3 zlib1g-dev libssh2-1-dev libssh2-1 -y
RUN cd libgit2 && git checkout v1.3.0 && mkdir build && cd build
RUN apt-get install libmbedtls-dev -y
RUN apt-get install libpcre3 libpcre3-dev -y
RUN apt-get install wget -y && wget https://www.libssh2.org/download/libssh2-1.10.0.tar.gz
RUN tar -xvzf libssh2-1.10.0.tar.gz
RUN cd libssh2-1.10.0 && mkdir bin && cd bin && cmake .. -DENABLE_ZLIB_COMPRESSION=ON -DBUILD_SHARED_LIBS=ON && cmake --build . --target install
RUN cd libgit2/build && cmake .. -DUSE_SSH=ON
RUN cd libgit2/build && cmake --build . --target install

RUN apt purge cmake -y
RUN apt purge build-essential -y
RUN apt purge pkg-config libssl-dev \
        python3 zlib1g-dev libssh2-1-dev libssh2-1 -y
RUN apt purge libmbedtls-dev -y
# RUN apt purge libpcre3 libpcre3-dev -y --allow-remove-essential

ENV LD_LIBRARY_PATH=/usr/local/lib/

RUN wget https://go.dev/dl/go1.18.linux-amd64.tar.gz
RUN tar -C /usr/local -xzf go1.18.linux-amd64.tar.gz
ENV PATH=$PATH:/usr/local/go/bin

RUN apt purge wget -y
RUN apt autoremove -y

RUN rm -rf go1.18.linux-amd64.tar.gz 
RUN rm -rf libgit2 
RUN rm -rf libssh2-1.10.0 
RUN rm -rf libssh2-1.10.0.tar.gz 

# RUN mkdir go
# COPY go.mod go/go.mod
# COPY go.sum go/go.sum
# RUN cd go && go mod download; 

# COPY .git go/.git
# COPY pkg go/pkg/
# COPY cmd go/cmd/

# RUN cd go && \
#     export GIT_HASH=`git rev-parse --short HEAD` && \
#     export GIT_DIRTY=`git diff --quiet || echo '-dirty'` && \
#     export CGO_LDFLAGS="-w -s" && \
#     export FULL_VERSION="${RELEASE}${RELEASE:+-}${GIT_HASH}${GIT_DIRTY}"; echo "${v%.*}" && \
#     echo "-X github.com/direktiv/direktiv/pkg/version.Version=${FULL_VERSION}" && \
#     go build -ldflags "-X github.com/direktiv/direktiv/pkg/version.Version=${FULL_VERSION}" -tags osusergo,netgo -o /bin/direktiv cmd/flow/*.go && \
#     go clean -i -r -cache -testcache -modcache; 

# RUN rm -rf go

# USER nonroot:nonroot

# EXPOSE 6666
# EXPOSE 7777

# CMD ["/bin/direktiv"]

## FROM gcr.io/distroless/static
##
## USER nonroot:nonroot
##
## COPY /usr/local/lib /usr/local/lib
## COPY build/flow /bin/direktiv
##
## EXPOSE 6666
## EXPOSE 7777
##
## CMD ["/bin/direktiv"]

# FROM gcr.io/distroless/static

# USER nonroot:nonroot

# COPY build/usr/local/lib /usr/local/lib
COPY build/flow /bin/direktiv

EXPOSE 6666
EXPOSE 7777

CMD ["/bin/direktiv"]
