# FROM alpine:3.13.2
#
# RUN wget -O /bin/grpc_health_probe https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/v0.3.6/grpc_health_probe-linux-amd64
# RUN chmod 755 /bin/grpc_health_probe
#
# COPY secrets /bin/secrets
# RUN chmod 755 /bin/secrets
#
# EXPOSE 2610
#
# CMD /bin/secrets
#

FROM gcr.io/distroless/static

USER nonroot:nonroot

COPY secrets /bin/secrets

EXPOSE 2610

CMD ["/bin/secrets"]
