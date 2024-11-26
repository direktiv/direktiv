FROM --platform=$BUILDPLATFORM docker.io/library/golang:1.23.0 as builder

ARG VERSION=dev

COPY go.mod src/go.mod
COPY go.sum src/go.sum
RUN cd src/ && go mod download

COPY pkg src/pkg/
COPY cmd src/cmd/

RUN --mount=type=cache,target=/root/.cache/go-build cd src && \
    CGO_ENABLED=false GOOS=linux GOARCH=$TARGETARCH go build -tags osusergo,netgo -ldflags "-X github.com/direktiv/direktiv/pkg/version.Version=$VERSION" -o /direktiv cmd/direktiv/*.go;

RUN --mount=type=cache,target=/root/.cache/go-build cd src && \
    CGO_ENABLED=false GOOS=linux GOARCH=$TARGETARCH go build -tags osusergo,netgo -o /direktiv-cmd cmd/cmd-exec/*.go;

#########################################################################################
FROM --platform=$BUILDPLATFORM node:18.18.1 as ui-builder
WORKDIR /app
ENV PATH /app/node_modules/.bin:$PATH

COPY ui/yarn.lock .
COPY ui/package.json .
RUN yarn install

COPY ui/assets assets
COPY ui/public public
COPY ui/src src

COPY ui/test test
COPY ui/.env.example .
COPY ui/.eslintrc.js .
COPY ui/.nvmrc .
COPY ui/index.html .
COPY ui/postcss.config.cjs .
COPY ui/tailwind.config.cjs .
COPY ui/tsconfig.json .
COPY ui/vite.config.ts .

RUN yarn build

########################################################################################
FROM  gcr.io/distroless/static
USER nonroot:nonroot

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=builder /direktiv /app/direktiv
COPY --from=builder /direktiv-cmd /app/direktiv-cmd
COPY --from=ui-builder /app/dist /app/ui

CMD ["/app/direktiv"]