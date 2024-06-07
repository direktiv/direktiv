FROM --platform=$BUILDPLATFORM docker.io/library/golang:1.22 as builder

ARG VERSION=dev

COPY go.mod src/go.mod
COPY go.sum src/go.sum
RUN cd src/ && go mod download

COPY pkg src/pkg/
COPY cmd src/cmd/

RUN --mount=type=cache,target=/root/.cache/go-build cd src && \
    CGO_ENABLED=false GOOS=linux GOARCH=$TARGETARCH go build -tags osusergo,netgo -ldflags "-X github.com/direktiv/direktiv/pkg/version.Version=$VERSION" -o /direktiv cmd/direktiv/*.go;


# Remove pkg folder so that the direktiv-cmd binary doesn't include logic.
RUN rm -rf pkg
RUN --mount=type=cache,target=/root/.cache/go-build cd src && \
    CGO_ENABLED=false GOOS=linux GOARCH=$TARGETARCH go build -tags osusergo,netgo -o /direktiv-cmd cmd/cmd-exec/*.go;


FROM  gcr.io/distroless/static
USER nonroot:nonroot

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /direktiv /bin/direktiv
COPY --from=builder /direktiv-cmd /bin/direktiv-cmd

CMD ["/bin/direktiv"]