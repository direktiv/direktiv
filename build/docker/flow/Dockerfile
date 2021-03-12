FROM alpine:3.13.2

COPY direktiv /bin/direktiv
RUN chmod 755 /bin/direktiv

EXPOSE 6666
EXPOSE 7777

CMD /bin/direktiv -t w
