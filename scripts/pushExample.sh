#!/bin/bash

cd examples/action && docker build -t localhost:5000/demo-action .

docker push localhost:5000/demo-action
