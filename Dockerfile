FROM --platform=$BUILDPLATFORM docker.io/library/golang:1.25.0 as builder

ARG IS_ENTERPRISE=false
ARG VERSION
ARG GIT_SHA

COPY go.mod src/go.mod
COPY go.sum src/go.sum
RUN cd src/ && go mod download

COPY internal src/internal/
COPY pkg src/pkg/
COPY cmd src/cmd/
COPY direktiv-ee*/internal src/direktiv-ee/internal

RUN if [ "$IS_ENTERPRISE" = "true" ]; then \
    echo "/direktiv direktiv-ee/internal/*.go" > BUILD_PATH.txt; \
    else \
    echo "/direktiv cmd/*.go" > BUILD_PATH.txt; \
    fi

RUN --mount=type=cache,target=/root/.cache/go-build cd src &&  \
    CGO_ENABLED=false GOOS=linux GOARCH=$TARGETARCH go build \
    -tags osusergo,netgo \
    -ldflags "-X github.com/direktiv/direktiv/internal/version.Version=$VERSION -X github.com/direktiv/direktiv/internal/version.GitSha=$GIT_SHA" \
    -o $(cat ../BUILD_PATH.txt);

#########################################################################################
FROM --platform=$BUILDPLATFORM node:20.18.1-slim as ui-builder

ENV PNPM_HOME="/pnpm"
ENV PATH="$PNPM_HOME:$PATH"
RUN corepack enable
RUN corepack prepare pnpm@9.15.4 --activate

WORKDIR /app


# install dependencies (only necessary files are copied to be able to cache this layer)
COPY ui/package.json .
COPY ui/pnpm-lock.yaml .
RUN pnpm install --frozen-lockfile

COPY ui/ ./

RUN pnpm run build

########################################################################################
FROM  gcr.io/distroless/static
USER nonroot:nonroot

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /direktiv /app/direktiv
COPY --from=ui-builder /app/dist /app/ui

CMD ["/app/direktiv"]