FROM golang:1.19-buster as builder

WORKDIR /dist

COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
COPY .git .git

ARG RELEASE_VERSION
ENV RELEASE=$RELEASE_VERSION

RUN --mount=type=cache,target=/root/.cache/go-build cd src && \
    export GIT_HASH=`git rev-parse --short HEAD` && \
    export GIT_DIRTY=`git diff --quiet || echo '-dirty'` && \
    export CGO_LDFLAGS="-static -w -s" && \
    export FULL_VERSION="${RELEASE:-$GIT_HASH}"; echo "${v%.*}" && \
    go build -ldflags "-X github.com/direktiv/direktiv/pkg/version.Version=$FULL_VERSION" -tags osusergo,netgo -o binary cmd/apps/*.go;

FROM gcr.io/distroless/static

USER nonroot:nonroot

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /dist/binary /dist/binary

ENTRYPOINT ["/dist/binary"]