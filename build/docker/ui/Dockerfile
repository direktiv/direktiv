FROM alpine

COPY docker/ui/web /srv
COPY docker/ui/direktiv-ui /
RUN chmod 755 /direktiv-ui

EXPOSE 1605

CMD /direktiv-ui
