#!/bin/bash

make docker-init-pod  && docker tag init-pod localhost:5000/init-pod

docker push localhost:5000/init-pod
