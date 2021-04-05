#!/bin/bash

make docker-ui && docker tag direktiv-ui localhost:5000/ui

docker push localhost:5000/ui
