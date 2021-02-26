FROM postgres:13.2-alpine

EXPOSE 6666

COPY install.sh /
RUN chmod 755 /install.sh
RUN /install.sh

ENV DIREKTIV_MINIO_SECURE=0
ENV DIREKTIV_SECURE=0
ENV MINIO_ACCESS_KEY=vorteil
ENV MINIO_SECRET_KEY=vorteilvorteil
ENV DIREKTIV_DB="host=127.0.0.1 port=5432 user=direktiv dbname=direktiv password=direktiv sslmode=disable"
ENV DIREKTIV_SECRETS_DB="host=127.0.0.1 port=5432 user=direktiv dbname=direktiv password=direktiv sslmode=disable"
#ENV DIREKTIV_DB="host=127.0.0.1 port=5432 user=sisatech dbname=postgres password=sisatech sslmode=disable"
#ENV DIREKTIV_SECRETS_DB="host=127.0.0.1 port=5432 user=sisatech dbname=postgres password=sisatech sslmode=disable"

ENV POSTGRES_USER=direktiv
ENV POSTGRES_PASSWORD=direktiv
ENV POSTGRES_DB=direktiv

COPY conf.toml /etc

COPY direktiv /bin/direktiv
RUN chmod 755 /bin/direktiv

COPY run.sh /
RUN chmod 755 /run.sh
RUN chmod 755 /usr/local/bin/docker-entrypoint.sh

RUN mkdir -p /data && chmod 755 /bin/minio

RUN apk --no-cache add sudo

CMD /run.sh
