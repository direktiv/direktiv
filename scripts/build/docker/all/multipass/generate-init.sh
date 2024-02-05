#!/bin/sh

b64=$(base64 -w 0 install.sh)
sed "s/SCRIPT/${b64}/g" init-template.yaml > init.yaml
