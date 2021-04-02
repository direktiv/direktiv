FROM golang:1.16 as build-env

WORKDIR /go/src/app
ADD . /go/src/app

RUN go build -o /go/bin/sidecar cmd/sidecar/main.go

FROM gcr.io/distroless/base
COPY --from=build-env /go/bin/sidecar /

EXPOSE 8889

CMD ["/sidecar"]
