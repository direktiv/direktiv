FROM ubuntu:21.10 as build-env

RUN apt-get update && apt-get install -y ca-certificates curl golang

RUN curl -fsSL -o /helm.tar.gz https://get.helm.sh/helm-v3.5.3-linux-amd64.tar.gz && tar -C / -xvzf /helm.tar.gz

RUN mkdir /app
COPY app/go.mod /app 
COPY app/go.sum /app
RUN cd /app && go mod download

COPY app/main.go /app

RUN cd /app && CGO_LDFLAGS="-static -w -s"  go build  -tags osusergo,netgo  -o /dlapp /app/main.go


# FROM rancher/k3s:v1.22.3-rc2-k3s1
FROM ubuntu:21.10

RUN apt-get update &&  DEBIAN_FRONTEND=noninteractive TZ=Etc/UTC apt-get -y install tzdata && apt-get install -y ca-certificates curl iproute2 dnsutils wget bash-completion git vim

# install kubectl
RUN curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
RUN install -o 0 -g 0 -m 0755 kubectl /usr/local/bin/kubectl

# add k3s
COPY k3s /
RUN chmod 755 /k3s

COPY --from=build-env /linux-amd64/helm /helm
RUN git clone https://github.com/direktiv/direktiv-charts.git 
RUN cd direktiv-charts/charts/direktiv && /helm dependencies update

COPY broker.yaml /
COPY profile /
RUN cat profile >> /root/.bashrc

COPY debug.yaml /
COPY registry.yaml /

COPY pg /pg
# COPY config-deployment.yaml /config-deployment.yaml

# k3s stuff
VOLUME /var/lib/kubelet
VOLUME /var/lib/rancher/k3s
VOLUME /var/lib/cni
VOLUME /var/log

ENV PATH="$PATH:/bin/aux"
ENV CRI_CONFIG_FILE="/var/lib/rancher/k3s/agent/etc/crictl.yaml"
ENV KUBECONFIG=/etc/rancher/k3s/k3s.yaml

# get containerd stuff
RUN wget https://github.com/containerd/containerd/releases/download/v1.5.7/containerd-1.5.7-linux-amd64.tar.gz
RUN tar xvf containerd-1.5.7-linux-amd64.tar.gz bin/ctr

COPY images.tar.gz /
COPY --from=build-env /dlapp /dlapp
RUN chmod 755 /dlapp


ENTRYPOINT ["/dlapp"]
