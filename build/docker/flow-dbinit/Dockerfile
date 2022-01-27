FROM alpine:3.15.0

COPY flow-dbinit /bin/flow-dbinit
RUN chmod 755 /bin/flow-dbinit

RUN apk add shadow
RUN /usr/sbin/groupadd -g 22222 direktivg && /usr/sbin/useradd -s /bin/sh -g 22222 -u 33333 direktivu

CMD /bin/flow-dbinit
