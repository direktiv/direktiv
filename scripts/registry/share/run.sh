#!/bin/bash

HOSTNAME=$1

echo "certificates for $1"

rm -Rf /share/out/
mkdir -p /share/out/

sed "s/SERVER_NAME/$HOSTNAME/g;" /share/csr.conf > /share/out/csr_out.conf 

echo "generating ca key"
openssl genrsa -des3 -passout pass:changeme -out /share/out/ca.key 4096

echo "generating ca cert"
openssl req -new -x509 -days 3650 -key /share/out/ca.key \
-out /share/out/ca.cert.pem -passin pass:changeme \
-subj "/CN=${HOSTNAME}"

echo "generating server key & csr"
openssl req -new \
-out /share/out/${HOSTNAME}.csr \
-config /share/out/csr_out.conf \
-newkey rsa:4096 \
-sha256 \
-nodes \
-keyout /share/out/${HOSTNAME}.key 

echo "signing csr"
openssl x509 -req -days 3650 \
-in /share/out/${HOSTNAME}.csr \
-CA /share/out/ca.cert.pem \
-CAkey /share/out/ca.key -CAcreateserial \
-extfile /share/out/csr_out.conf -extensions v3_req \
-out /share/out/${HOSTNAME}.crt \
-passin pass:changeme

chown 1000:1000 /share/out/*

