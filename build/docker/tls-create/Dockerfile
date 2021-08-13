FROM ubuntu

RUN apt-get update && apt-get install -y curl

RUN curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"

RUN install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl

ADD docker/tls-create/issuer.yaml /
ADD docker/tls-create/script.sh /

RUN chmod 755 script.sh

RUN echo 1

CMD ["/script.sh"]
