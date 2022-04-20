FROM gcr.io/distroless/static

USER nonroot:nonroot

COPY build/api /bin/server

EXPOSE 8080

CMD ["/bin/server", "-c", "/config/api-config.yaml"]
