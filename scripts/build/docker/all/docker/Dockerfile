# This stage builds the installation app
FROM ubuntu:22.04 as build-env

RUN apt-get update && apt-get install -y ca-certificates curl golang

RUN mkdir /app
COPY installer/go.mod /app 
COPY installer/go.sum /app
RUN cd /app && go mod download

COPY installer/main.go /app

RUN cd /app && CGO_LDFLAGS="-static -w -s"  go build  -tags osusergo,netgo  -o /dlapp /app/main.go

# This stage is just a base image for the actual final stage
FROM ubuntu:22.04 as base

ARG K3S_VERSION=v1.26.3%2Bk3s1

RUN apt-get update && apt-get install wget curl netcat git bash-completion -y

RUN wget https://github.com/k3s-io/k3s/releases/download/${K3S_VERSION}/k3s-airgap-images-amd64.tar
RUN mkdir -p /var/lib/rancher/k3s/agent/images/
RUN mv k3s-airgap-images-amd64.tar /var/lib/rancher/k3s/agent/images/

RUN wget https://github.com/k3s-io/k3s/releases/download/${K3S_VERSION}/k3s
RUN chmod 755 /k3s
RUN mkdir -p /usr/local/bin
RUN mv k3s /usr/local/bin

RUN curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash

RUN curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
RUN install -o 0 -g 0 -m 0755 kubectl /usr/local/bin/kubectl

RUN helm repo add direktiv https://charts.direktiv.io
RUN helm repo add nginx https://kubernetes.github.io/ingress-nginx
RUN helm repo add prometheus https://prometheus-community.github.io/helm-charts
RUN helm repo add fluent-bit https://fluent.github.io/helm-charts
RUN helm repo update

FROM base

# k3s stuff
VOLUME /var/lib/kubelet
VOLUME /var/lib/rancher/k3s
VOLUME /var/lib/cni
VOLUME /var/log

COPY eventing.yaml /
COPY broker.yaml /
COPY profile /
RUN cat profile >> /root/.bashrc

COPY debug.yaml /
COPY registry.yaml /
COPY pg /pg

ENV KUBECONFIG=/etc/rancher/k3s/k3s.yaml

COPY --from=build-env /dlapp /dlapp
RUN chmod 755 /dlapp
CMD ["/dlapp"]