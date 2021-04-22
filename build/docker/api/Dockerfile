FROM alpine:3.13.2

COPY api /bin/server
RUN chmod 755 /bin/server

EXPOSE 8080

CMD ["/bin/server", "-c", "/config/default.toml"]
