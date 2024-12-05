FROM --platform=$BUILDPLATFORM docker.io/library/golang:1.23.0 as builder

ARG VERSION=dev

COPY go.mod src/go.mod
COPY go.sum src/go.sum
RUN cd src/ && go mod download

COPY pkg src/pkg/
COPY cmd src/cmd/

RUN --mount=type=cache,target=/root/.cache/go-build cd src && \
    CGO_ENABLED=false GOOS=linux GOARCH=$TARGETARCH go build -tags osusergo,netgo -ldflags "-X github.com/direktiv/direktiv/pkg/version.Version=$VERSION" -o /direktiv cmd/*.go;

#########################################################################################
FROM --platform=$BUILDPLATFORM node:20-slim as ui-builder
ENV PNPM_HOME="/pnpm"
ENV PATH="$PNPM_HOME:$PATH"
RUN corepack enable

WORKDIR /app

COPY ui/package.json .
COPY ui/pnpm-lock.yaml .

RUN pnpm install

COPY ui/.eslintrc.js .
COPY ui/.prettierrc.mjs .
COPY ui/index.html .
COPY ui/postcss.config.cjs .
COPY ui/tailwind.config.cjs .
COPY ui/tsconfig.json .
COPY ui/vite.config.mts .
COPY ui/assets assets
COPY ui/public public
COPY ui/src src
COPY ui/test test

RUN pnpm run build

########################################################################################
FROM  gcr.io/distroless/static
USER nonroot:nonroot

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /direktiv /app/direktiv
COPY --from=ui-builder /app/dist /app/ui

CMD ["/app/direktiv"]