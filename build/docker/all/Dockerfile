# This layer preloads the images needed by Direktiv and other components.
# It will reduce the startup time for k3s
FROM ubuntu:21.10 as images

# RUN apt-get update && apt-get install wget curl jq  skopeo -y
RUN apt-get update && apt-get install wget ca-certificates skopeo -y
RUN mkdir images

COPY download-image.sh /

ARG DIREKTIV_VERSION=v0.6.6
ARG KNATIVE_VERSION=v1.5.0

RUN ./download-image.sh \   
    direktiv/flow:${DIREKTIV_VERSION} \
    direktiv/secrets:${DIREKTIV_VERSION} \
    direktiv/functions:${DIREKTIV_VERSION} \
    direktiv/api:${DIREKTIV_VERSION} \
    direktiv/ui:${DIREKTIV_VERSION}

# # download knative serving images
RUN wget https://github.com/knative/serving/releases/download/knative-${KNATIVE_VERSION}/serving-core.yaml
RUN wget https://github.com/knative/net-contour/releases/download/knative-${KNATIVE_VERSION}/net-contour.yaml

RUN cat serving-core.yaml | sed -n -e '/gcr.io\/knative-releases/p' | cut -d : -f 2-3 | xargs ./download-image.sh
RUN cat net-contour.yaml | sed -n -e '/gcr.io\/knative-releases/p' | cut -d : -f 2-3 | xargs ./download-image.sh

RUN ls -la images

# tar up all images to use later in final image
RUN cd images && tar -czf /images.tar.gz *

# This layer builds the installation app
FROM ubuntu:21.10 as build-env

RUN apt-get update && apt-get install -y ca-certificates curl golang

# RUN curl -fsSL -o /helm.tar.gz https://get.helm.sh/helm-v3.9.0-linux-amd64.tar.gz && tar -C / -xvzf /helm.tar.gz

RUN mkdir /app
COPY app/go.mod /app 
COPY app/go.sum /app
RUN cd /app && go mod download

COPY app/main.go /app

RUN cd /app && CGO_LDFLAGS="-static -w -s"  go build  -tags osusergo,netgo  -o /dlapp /app/main.go


FROM ubuntu:21.10 as base

ARG K3S_VERSION=v1.24.2%2Bk3s2

RUN apt-get update && apt-get install wget curl netcat -y

RUN wget https://github.com/k3s-io/k3s/releases/download/${K3S_VERSION}/k3s-airgap-images-amd64.tar
RUN mkdir -p /var/lib/rancher/k3s/agent/images/
RUN mv k3s-airgap-images-amd64.tar /var/lib/rancher/k3s/agent/images/

RUN wget https://github.com/k3s-io/k3s/releases/download/${K3S_VERSION}/k3s
RUN chmod 755 /k3s
RUN mkdir -p /usr/local/bin
RUN mv k3s /usr/local/bin

RUN curl -fsSL -o /helm.tar.gz https://get.helm.sh/helm-v3.9.0-linux-amd64.tar.gz && tar -C / -xvzf /helm.tar.gz
RUN cp /linux-amd64/helm /helm && rm -Rf /linux-amd64

# RUN curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
# RUN install -o 0 -g 0 -m 0755 kubectl /usr/local/bin/kubectl

RUN git clone https://github.com/direktiv/direktiv-charts.git

FROM base

# COPY --from=build-env /linux-amd64/helm /helm

COPY --from=images images.tar.gz /
RUN tar -C  /var/lib/rancher/k3s/agent/images/ -xvzf images.tar.gz

# k3s stuff
VOLUME /var/lib/kubelet
VOLUME /var/lib/rancher/k3s
VOLUME /var/lib/cni
VOLUME /var/log

COPY run.sh /

# ARG CACHEBUST=0

# # CACHEBUST="$(shell date +%s)$(shell echo $$PPID)"

# # RUN rm -Rf /etc/rancher/k3s/k3s.yaml
# RUN nohup bash -c "/k3s server --kube-proxy-arg=conntrack-max-per-core=0 --disable traefik --write-kubeconfig-mode=644 &" && sleep 5



# # # wait for file to be created
# RUN while [ ! -f /etc/rancher/k3s/k3s.yaml ]; do  sleep 2; done
# # RUN while ! nc -z 127.0.0.1 6443; do sleep 1; done

# RUN echo START!!!!!!!!!!!!!!!!!!

# RUN ls -la
# RUN /k3s ctr image pull docker.io/calico/node:v3.11.2



# RUN ls -la /etc/rancher/k3s/k3s.yaml

# RUN while ! nc -z 127.0.0.1 6443; do sleep 1; done

# RUN /usr/local/bin/kubectl get pods -A

# start k3s

	# cmd := exec.Command("/k3s", "server", "--kube-proxy-arg=conntrack-max-per-core=0",
	# 	"--disable", "traefik", "--write-kubeconfig-mode=644", "--kube-apiserver-arg",
	# 	"feature-gates=TTLAfterFinished=true")

# RUN apt-get update &&  DEBIAN_FRONTEND=noninteractive TZ=Etc/UTC apt-get -y install tzdata && apt-get install -y ca-certificates curl iproute2 dnsutils wget bash-completion git vim

# install kubectl


# # add k3s
# COPY k3s /
# RUN chmod 755 /k3s

# COPY --from=build-env /linux-amd64/helm /helm
# RUN git clone https://github.com/direktiv/direktiv-charts.git 
# RUN cd direktiv-charts/charts/direktiv && /helm dependencies update

# COPY broker.yaml /
COPY profile /
RUN cat profile >> /root/.bashrc

# COPY debug.yaml /
# COPY registry.yaml /

# COPY pg /pg
# # COPY config-deployment.yaml /config-deployment.yaml

# # k3s stuff
# VOLUME /var/lib/kubelet
# VOLUME /var/lib/rancher/k3s
# VOLUME /var/lib/cni
# VOLUME /var/log

# ENV PATH="$PATH:/bin/aux"
# ENV CRI_CONFIG_FILE="/var/lib/rancher/k3s/agent/etc/crictl.yaml"
# ENV KUBECONFIG=/etc/rancher/k3s/k3s.yaml

# # get containerd stuff
# RUN wget https://github.com/containerd/containerd/releases/download/v1.5.7/containerd-1.5.7-linux-amd64.tar.gz
# RUN tar xvf containerd-1.5.7-linux-amd64.tar.gz bin/ctr

# COPY images.tar.gz /
COPY --from=build-env /dlapp /dlapp
RUN chmod 755 /dlapp
CMD ["/dlapp"]

# CMD ["/usr/local/bin/k3s", "server", "--kube-proxy-arg=conntrack-max-per-core=0", "--disable", "traefik", "--write-kubeconfig-mode=644"]

# ENTRYPOINT ["/usr/local/bin/k3s", "server", "--kube-proxy-arg=conntrack-max-per-core=0", "--disable", "traefik", "--write-kubeconfig-mode=644"]

 	# cmd := exec.Command("/k3s", "server", "--kube-proxy-arg=conntrack-max-per-core=0",
	# 	"--disable", "traefik", "--write-kubeconfig-mode=644", "--kube-apiserver-arg",
	# 	"feature-gates=TTLAfterFinished=true")