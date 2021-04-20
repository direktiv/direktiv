#!/bin/bash

make docker-secrets && docker tag direktiv-secrets localhost:5000/secrets

docker push localhost:5000/secrets
