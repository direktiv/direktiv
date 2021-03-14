FROM alpine:3.13.2

COPY docker/isolate/install.sh /install.sh
RUN chmod 755 /install.sh && /install.sh && rm /install.sh

COPY direktiv /bin/direktiv
RUN chmod 755 /bin/direktiv

EXPOSE 8888

CMD /bin/direktiv -t i
