FROM alpine

ADD docker/ui/build /web
COPY docker/ui/direktiv-ui /
RUN chmod 755 /direktiv-ui

EXPOSE 1605

CMD /direktiv-ui
