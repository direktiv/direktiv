FROM golang:1.17.7-buster as build

COPY go.mod src/go.mod
COPY go.sum src/go.sum
RUN cd src/ && go mod download

COPY pkg src/pkg/
COPY cmd src/cmd/
COPY .git .git

RUN --mount=type=cache,target=/root/.cache/go-build cd src && \
    export GIT_HASH=`git rev-parse --short HEAD` && \
    export GIT_DIRTY=`git diff --quiet || echo '-dirty'` && \
    export CGO_LDFLAGS="-static -w -s" && \
    export FULL_VERSION="${RELEASE}${RELEASE:+-}${GIT_HASH}${GIT_DIRTY}"; echo "${v%.*}" && \
    go build -ldflags "-X github.com/direktiv/direktiv/pkg/version.Version=$FULL_VERSION" -tags osusergo,netgo -o /flow-dbinit cmd/flow-dbinit/*.go; 

FROM gcr.io/distroless/static

USER nonroot:nonroot

COPY --from=build /flow-dbinit /bin/flow-dbinit

EXPOSE 6666
EXPOSE 7777

CMD ["/bin/flow-dbinit"]