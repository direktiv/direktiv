FROM alpine:3.13.2

RUN wget -O /bin/grpc_health_probe https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/v0.3.6/grpc_health_probe-linux-amd64
RUN chmod 755 /bin/grpc_health_probe

COPY docker/isolate/install.sh /install.sh
RUN chmod 755 /install.sh && /install.sh && rm /install.sh

COPY direktiv /bin/direktiv
RUN chmod 755 /bin/direktiv

EXPOSE 8888

CMD /bin/direktiv -t i
