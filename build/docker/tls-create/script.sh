#!/bin/sh

# openssl genrsa -out servercakey.pem
# openssl req -new -x509 -key servercakey.pem -out serverca.crt

openssl req -new -x509 -sha256 -newkey rsa:2048 -nodes \
    -keyout key.pem -out cert.pem -days 3650 \
    -subj "/C=AU/ST=QLD/L=Varsity/O=Foo Corp/OU=Bar Div/CN=www.foo.com"

kubectl create secret tls direktiv-ca \
  --cert=serverca.crt \
  --key=servercakey.pem

kubectl apply -f issuer.yaml
