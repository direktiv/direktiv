FROM alpine

ADD docker/ui/build /web
COPY docker/ui/direktiv-ui /
RUN chmod 755 /direktiv-ui

EXPOSE 8080

RUN apk add shadow
RUN /usr/sbin/groupadd -g 22222 direktivg && /usr/sbin/useradd -s /bin/sh -g 22222 -u 33333 direktivu

CMD /direktiv-ui
