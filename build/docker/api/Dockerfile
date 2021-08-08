FROM alpine:3.13.2

COPY api /bin/server
RUN chmod 755 /bin/server

EXPOSE 8080

RUN apk add shadow
RUN /usr/sbin/groupadd -g 22222 direktivg && /usr/sbin/useradd -s /bin/sh -g 22222 -u 33333 direktivu

CMD ["/bin/server", "-c", "/config/api-config.yaml"]
