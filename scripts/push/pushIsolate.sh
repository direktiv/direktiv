#!/bin/bash

make docker-isolates && docker tag direktiv-isolates localhost:5000/isolates

docker push localhost:5000/isolates
