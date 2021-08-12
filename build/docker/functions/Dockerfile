FROM alpine:3.13.2

COPY functions /bin/functions
RUN chmod 755 /bin/functions

EXPOSE 5555

RUN apk add shadow
RUN /usr/sbin/groupadd -g 22222 direktivg && /usr/sbin/useradd -s /bin/sh -g 22222 -u 33333 direktivu

CMD /bin/functions
