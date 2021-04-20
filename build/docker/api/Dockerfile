FROM alpine:3.13.2

RUN wget -O /bin/grpc_health_probe https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/v0.3.6/grpc_health_probe-linux-amd64
RUN chmod 755 /bin/grpc_health_probe

COPY apiserver /bin/server
RUN chmod 755 /bin/server

EXPOSE 8080

CMD ["/bin/server", "-c", "/config/default.toml"]
