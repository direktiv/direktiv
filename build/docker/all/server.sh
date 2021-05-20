#!/bin/sh

socat TCP-LISTEN:9090,crlf,reuseaddr,fork SYSTEM:"echo HTTP/1.0 200; echo Content-Type\: text/plain; echo; cat /etc/rancher/k3s/k3s.yaml"
