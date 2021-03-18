FROM alpine

COPY docker/ui/web /web
COPY docker/ui/direktiv-ui /
RUN chmod 755 /direktiv-ui

EXPOSE 1605

CMD /direktiv-ui
