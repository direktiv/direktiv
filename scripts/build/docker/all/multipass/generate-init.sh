#!/bin/sh

# call example: ./generate-init.sh direktiv/direktiv direktiv/frontend v0.8.0

if [ $# -lt 3 ]
  then
    echo "requires images <direktiv> <direktiv-ui> <tag>"
fi

arg1=`echo $1 | sed 's/\//\\\\\\//g'`
arg2=`echo $2 | sed 's/\//\\\\\\//g'`


b64=$(base64 -w 0 install.sh)
sed "s/SCRIPT/${b64}/g" init-template.yaml > init.yaml

sed -i "s/ARG1/${arg1}/g" init.yaml
sed -i "s/ARG2/${arg2}/g" init.yaml
sed -i "s/ARG3/$3/g" init.yaml
