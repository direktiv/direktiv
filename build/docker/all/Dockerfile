FROM alpine/git as build-env

RUN git clone https://github.com/vorteil/direktiv.git /direktiv

RUN apk add curl

RUN apk add go

ADD docker/all/app /app

RUN cd /app && CGO_LDFLAGS="-static -w -s"  go build  -tags osusergo,netgo  -o /dlapp /app/main.go

RUN curl -fsSL -o /helm.tar.gz https://get.helm.sh/helm-v3.5.3-linux-amd64.tar.gz && tar -C / -xvzf /helm.tar.gz

FROM rancher/k3s:v1.21.1-rc1-k3s1
COPY --from=build-env /linux-amd64/helm /helm
COPY --from=build-env /direktiv /direktiv
COPY --from=build-env /dlapp /dlapp

RUN chmod 755 /dlapp

ADD docker/all/debug.yaml /
ADD docker/all/registry.yaml /

ADD docker/all/google-dns.yaml /

ADD docker/all/server.sh /

ADD docker/all/default_direktiv.vcfg /

ADD docker/all/config-deployment.yaml /

ENTRYPOINT ["/dlapp"]
