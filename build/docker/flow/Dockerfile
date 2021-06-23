FROM alpine:3.13.2

RUN wget -O /bin/grpc_health_probe https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/v0.3.6/grpc_health_probe-linux-amd64
RUN chmod 755 /bin/grpc_health_probe

COPY direktiv /bin/direktiv
RUN chmod 755 /bin/direktiv

EXPOSE 6666
EXPOSE 7777

RUN apk add shadow
RUN /usr/sbin/groupadd -g 22222 direktivg && /usr/sbin/useradd -s /bin/sh -g 22222 -u 33333 direktivu

CMD /bin/direktiv
