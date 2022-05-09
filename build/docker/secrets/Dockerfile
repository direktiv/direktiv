FROM gcr.io/distroless/static

USER nonroot:nonroot

COPY build/secrets /bin/secrets

EXPOSE 2610

CMD ["/bin/secrets"]
