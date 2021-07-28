#!/bin/bash
set -e
make docker-flow
docker tag direktiv-flow localhost:5000/flow
docker push localhost:5000/flow
set +e
