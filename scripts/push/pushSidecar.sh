#!/bin/bash

# call from direktiv as scripts/pushSidecar.sh

set -e 
make docker-sidecar 
docker tag sidecar localhost:5000/sidecar
docker push localhost:5000/sidecar
set +e
