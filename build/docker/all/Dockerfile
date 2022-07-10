# This stage preloads the images needed by Direktiv and other components.
# It will reduce the startup time for k3s with preloaded images.
FROM alpine:3.16 as images

RUN apk update \
	&& apk upgrade \
	&& apk add --no-cache \
	ca-certificates \
	skopeo coreutils wget curl \
	&& rm -rf /var/cache/apk/*

RUN update-ca-certificates   /var/lib/rancher/k3s/agent/images/tificates skopeo -y
RUN mkdir images

COPY download-image.sh /


RUN ls -la /bin
ARG DIREKTIV_VERSION=v0.6.6
ARG KNATIVE_VERSION=v1.5.0

RUN ./download-image.sh \   
    direktiv/flow:${DIREKTIV_VERSION} \
    direktiv/secrets:${DIREKTIV_VERSION} \
    direktiv/functions:${DIREKTIV_VERSION} \
    direktiv/api:${DIREKTIV_VERSION} \
    direktiv/ui:${DIREKTIV_VERSION}

RUN ./download-image.sh \   
    docker.io/envoyproxy/envoy:v1.21.1 \
    ghcr.io/projectcontour/contour:v1.20.1 \
	postgres:13.4 \
	registry:2.7.1

# download knative slayerrving and contour yaml. Extract the images to preload them for k3s.
RUN wget https://github.com/knative/serving/releases/download/knative-${KNATIVE_VERSION}/serving-core.yaml
RUN wget https://github.com/knative/net-contour/releases/download/knative-${KNATIVE_VERSION}/net-contour.yaml

RUN cat serving-core.yaml | sed -n -e '/gcr.io\/knative-releases/p' | cut -d : -f 2-3 | xargs ./download-image.sh
RUN cat net-contour.yaml | sed -n -e '/gcr.io\/knative-releases/p' | cut -d : -f 2-3 | xargs ./download-image.sh

# tar up all images to use later in final image
RUN cd images && tar -czf /images.tar.gz *

RUN ls -la /images

# just in case prepare eventing
RUN mkdir /eventing
RUN wget -P /eventing https://github.com/knative/eventing/releases/download/knative-${KNATIVE_VERSION}/eventing-crds.yaml
RUN wget -P /eventing https://github.com/knative/eventing/releases/download/knative-${KNATIVE_VERSION}/eventing-core.yaml
RUN wget -P /eventing https://github.com/knative/eventing/releases/download/knative-${KNATIVE_VERSION}/in-memory-channel.yaml
RUN wget -P /eventing https://github.com/knative/eventing/releases/download/knative-${KNATIVE_VERSION}/mt-channel-broker.yaml


# This stage builds the installation app
FROM ubuntu:21.10 as build-env

RUN apt-get update && apt-get install -y ca-certificates curl golang

RUN mkdir /app
COPY app/go.mod /app 
COPY app/go.sum /app
RUN cd /app && go mod download

COPY app/main.go /app

RUN cd /app && CGO_LDFLAGS="-static -w -s"  go build  -tags osusergo,netgo  -o /dlapp /app/main.go

# This stage is just a base image for the actual final stage
FROM ubuntu:21.10 as base

ARG K3S_VERSION=v1.24.2%2Bk3s2

RUN apt-get update && apt-get install wget curl netcat git bash-completion -y

RUN wget https://github.com/k3s-io/k3s/releases/download/${K3S_VERSION}/k3s-airgap-images-amd64.tar
RUN mkdir -p /var/lib/rancher/k3s/agent/images/
RUN mv k3s-airgap-images-amd64.tar /var/lib/rancher/k3s/agent/images/

RUN wget https://github.com/k3s-io/k3s/releases/download/${K3S_VERSION}/k3s
RUN chmod 755 /k3s
RUN mkdir -p /usr/local/bin
RUN mv k3s /usr/local/bin

RUN curl -fsSL -o /helm.tar.gz https://get.helm.sh/helm-v3.9.0-linux-amd64.tar.gz && tar -C / -xvzf /helm.tar.gz
RUN cp /linux-amd64/helm /helm && rm -Rf /linux-amd64

RUN curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
RUN install -o 0 -g 0 -m 0755 kubectl /usr/local/bin/kubectl

RUN git clone https://github.com/direktiv/direktiv-charts.git
RUN /helm repo add nginx https://kubernetes.github.io/ingress-nginx
RUN /helm repo add prometheus https://prometheus-community.github.io/helm-charts
RUN /helm dependency build /direktiv-charts/charts/direktiv

# This/
FROM base

COPY --from=images /eventing /eventing
COPY --from=images images.tar.gz /
RUN tar -C  /var/lib/rancher/k3s/agent/images/ -xvzf images.tar.gz

# k3s stuff
VOLUME /var/lib/kubelet
VOLUME /var/lib/rancher/k3s
VOLUME /var/lib/cni
VOLUME /var/log

# COPY broker.yaml /
COPY profile /
RUN cat profile >> /root/.bashrc

COPY debug.yaml /
COPY registry.yaml /
COPY pg /pg

ENV KUBECONFIG=/etc/rancher/k3s/k3s.yaml

COPY --from=build-env /dlapp /dlapp
RUN chmod 755 /dlapp
CMD ["/dlapp"]