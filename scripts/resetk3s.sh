#!/bin/sh

service k3s stop

rm -Rf /etc/rancher/k3s
rm -Rf /var/lib/rancher/k3s

service k3s start

sleep 30
