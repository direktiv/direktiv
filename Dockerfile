FROM --platform=$BUILDPLATFORM docker.io/library/golang:1.24.0 as builder

ARG VERSION=dev
ARG IS_ENTERPRISE=false

COPY go.mod src/go.mod
COPY go.sum src/go.sum
RUN cd src/ && go mod download

COPY pkg src/pkg/
COPY cmd src/cmd/
COPY direktiv-ee*/pkg src/direktiv-ee/pkg

RUN if [ "$IS_ENTERPRISE" = "true" ]; then \
    echo "/direktiv direktiv-ee/pkg/*.go" > BUILD_PATH.txt; \
    else \
    echo "/direktiv cmd/*.go" > BUILD_PATH.txt; \
    fi

RUN --mount=type=cache,target=/root/.cache/go-build cd src &&  \
    CGO_ENABLED=false GOOS=linux GOARCH=$TARGETARCH go build -tags osusergo,netgo -ldflags "-X github.com/direktiv/direktiv/pkg/version.Version=$VERSION" -o $(cat ../BUILD_PATH.txt);

#########################################################################################
FROM --platform=$BUILDPLATFORM node:20-slim as ui-builder
ENV PNPM_HOME="/pnpm"
ENV PATH="$PNPM_HOME:$PATH"
RUN corepack enable
RUN corepack prepare pnpm@9.15.4 --activate

WORKDIR /app

COPY ui/package.json .
COPY ui/pnpm-lock.yaml .

RUN pnpm install --frozen-lockfile

COPY ui/.eslintrc.js .
COPY ui/.prettierrc.mjs .
COPY ui/.prettierignore .
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