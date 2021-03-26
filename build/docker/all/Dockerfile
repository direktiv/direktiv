FROM postgres:13.2-alpine

EXPOSE 6666

RUN apk --no-cache add podman --repository=http://dl-cdn.alpinelinux.org/alpine/edge/community
RUN apk --no-cache add sudo ca-certificates iptables ip6tables device-mapper udev make go openssl git

COPY docker/ui/direktiv-ui /bin
ADD docker/ui/build /web

COPY docker/all/install.sh /
RUN chmod 755 /install.sh && /install.sh

ENV MINIO_ACCESS_KEY=vorteil
ENV MINIO_SECRET_KEY=vorteilvorteil
ENV DIREKTIV_DB="host=127.0.0.1 port=5432 user=direktiv dbname=direktiv password=direktiv sslmode=disable"
ENV DIREKTIV_SECRETS_DB="host=127.0.0.1 port=5432 user=direktiv dbname=direktiv password=direktiv sslmode=disable"

ENV POSTGRES_USER=direktiv
ENV POSTGRES_PASSWORD=direktiv
ENV POSTGRES_DB=direktiv
ENV DIREKTIV_INGRESS_BIND==0.0.0.0:6666 

COPY direktiv /bin/direktiv
RUN chmod 755 /bin/direktiv

RUN mkdir -p /data && chmod 755 /bin/minio

COPY docker/all/run.sh /
RUN chmod 755 /run.sh

#ENTRYPOINT ["/run.sh"]
CMD ["/run.sh"]
