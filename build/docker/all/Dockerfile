FROM alpine/git as build-env

RUN git clone https://github.com/vorteil/direktiv.git /direktiv

RUN apk add curl

RUN curl -fsSL -o /helm.tar.gz https://get.helm.sh/helm-v3.5.3-linux-amd64.tar.gz && tar -C / -xvzf /helm.tar.gz

FROM rancher/k3s
COPY --from=build-env /direktiv /direktiv

COPY --from=build-env /linux-amd64/helm /helm

RUN chmod -R 755 direktiv

ADD run.sh /
RUN chmod 755 /run.sh

ADD debug.yaml /

ENTRYPOINT ["/run.sh"]
