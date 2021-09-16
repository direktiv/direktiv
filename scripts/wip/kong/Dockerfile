FROM kong:2.5.1

COPY proto /usr/local/kong/proto

ADD myplugin  /usr/local/share/lua/5.1/kong/plugins/grpc-stream
