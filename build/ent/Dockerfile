FROM ubuntu:22.04

RUN apt-get update && apt-get install unzip curl wget -y

RUN wget https://go.dev/dl/go1.20.0.linux-amd64.tar.gz
RUN rm -rf /usr/local/go && tar -C /usr/local -xzf go1.20.0.linux-amd64.tar.gz

ENV PATH="${PATH}:/usr/local/go/bin:/root/go/bin"

WORKDIR /ent

COPY run.sh /run.sh
RUN chmod 755 /run.sh

ENTRYPOINT ["/run.sh"]