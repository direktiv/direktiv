FROM golang:1.19 as builder

ARG RELEASE_VERSION
ENV RELEASE=$RELEASE_VERSION

COPY go.mod src/go.mod
COPY go.sum src/go.sum
RUN cd src/ && go mod download

COPY pkg src/pkg/
COPY cmd src/cmd/
COPY .git .git

RUN --mount=type=cache,target=/root/.cache/go-build cd src && \
    export GIT_HASH=`git rev-parse --short HEAD` && \
    export FULL_VERSION="${RELEASE:-$GIT_HASH}"; echo "${v%.*}" && \
    go build -ldflags "-X github.com/direktiv/direktiv/pkg/version.Version=$FULL_VERSION" -o /direktiv cmd/direktiv/*.go;

FROM gcr.io/distroless/static

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /direktiv /bin/direktiv

ENV DIREKTIV_APP=flow
EXPOSE 6666
EXPOSE 7777
CMD ["/bin/direktiv"]