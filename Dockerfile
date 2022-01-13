FROM node:17 as gui-build

ARG FULL_VERSION

RUN echo "building $FULL_VERSION"
WORKDIR /app

ENV PATH /app/node_modules/.bin:$PATH


COPY public ./public
COPY src ./src
COPY package.json ./
COPY yarn.lock ./

RUN yarn install
# If this causes problems on github actions: A potential fix is to change the builder image to `node:alpine`
RUN NODE_OPTIONS=--openssl-legacy-provider REACT_APP_VERSION=$FULL_VERSION yarn build

FROM golang:1.16-buster as server-build

WORKDIR /go/src/app
ADD ./reactjs-embed/. /go/src/app
COPY --from=gui-build /app/build /go/src/app/build

RUN go get -d -v
RUN CGO_ENABLED=0 go build -o /server -ldflags="-s -w" main.go


FROM alpine:latest

RUN apk add shadow
RUN /usr/sbin/groupadd -g 22222 direktivg && /usr/sbin/useradd -s /bin/sh -g 22222 -u 33333 direktivu

COPY --from=server-build /server /
CMD ["/server"]