FROM kong:2.5.1

COPY proto /usr/local/kong/proto

ADD myplugin  /usr/local/share/lua/5.1/kong/plugins/grpc-stream

# USER 1000:1000

# RUN rm -f /usr/local/share/lua/5.1/kong/plugins/grpc-gateway/deco.lua

# COPY --chown=1000:1000 grpc-gateway/deco.lua  /usr/local/share/lua/5.1/kong/plugins/grpc-gateway/deco.lua

