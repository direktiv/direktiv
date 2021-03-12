FROM alpine:3.13.2

COPY direktiv /bin/direktiv
RUN chmod 755 /bin/direktiv

EXPOSE 2610

CMD /bin/direktiv -t s
