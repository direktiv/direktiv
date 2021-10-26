FROM alpine/git as build-env

#RUN git clone https://github.com/vorteil/direktiv.git /direktiv

RUN mkdir /direktiv

ADD kubernetes /direktiv/kubernetes

RUN apk add curl

RUN apk add go

ADD app /app

RUN cd /app && go mod download

RUN cd /app && CGO_LDFLAGS="-static -w -s"  go build  -tags osusergo,netgo  -o /dlapp /app/main.go

RUN curl -fsSL -o /helm.tar.gz https://get.helm.sh/helm-v3.5.3-linux-amd64.tar.gz && tar -C / -xvzf /helm.tar.gz

FROM rancher/k3s:v1.21.5-k3s1-amd64

COPY --from=build-env /linux-amd64/helm /helm
COPY --from=build-env /direktiv /direktiv
COPY --from=build-env /dlapp /dlapp

RUN chmod 755 /dlapp

COPY broker.yaml /
COPY profile /etc
COPY debug.yaml /
COPY registry.yaml /

COPY images.tar.gz /

COPY pg /pg
COPY config-deployment.yaml /config-deployment.yaml

ENTRYPOINT ["/dlapp"]
