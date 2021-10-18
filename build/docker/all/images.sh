#!/bin/bash

# knative
VERSION=0.25.1

docker pull gcr.io/knative-releases/knative.dev/serving/cmd/queue@sha256:5af97fa5e19c6fd1e04642e7cf6eb01e6982b4cfee19d8fd16649423277c2eb9
docker tag gcr.io/knative-releases/knative.dev/serving/cmd/queue@sha256:5af97fa5e19c6fd1e04642e7cf6eb01e6982b4cfee19d8fd16649423277c2eb9 gcr.io/knative-releases/knative.dev/serving/cmd/queue:$VERSION
docker save --output=queue.tar gcr.io/knative-releases/knative.dev/serving/cmd/queue:$VERSION

docker pull gcr.io/knative-releases/knative.dev/serving/cmd/activator@sha256:00aeb9267dc6445bbb11cb90636e9cde44404c6303d55b81aa074381b4989eef
docker tag gcr.io/knative-releases/knative.dev/serving/cmd/activator@sha256:00aeb9267dc6445bbb11cb90636e9cde44404c6303d55b81aa074381b4989eef gcr.io/knative-releases/knative.dev/serving/cmd/activator:$VERSION
docker save --output=activator.tar gcr.io/knative-releases/knative.dev/serving/cmd/activator:$VERSION

docker pull gcr.io/knative-releases/knative.dev/serving/cmd/autoscaler@sha256:ca3bad368a2ac40f33ad9e47c1075cc2b833301b4bc772fb84c51f52cc1c0a35
docker tag gcr.io/knative-releases/knative.dev/serving/cmd/autoscaler@sha256:ca3bad368a2ac40f33ad9e47c1075cc2b833301b4bc772fb84c51f52cc1c0a35 gcr.io/knative-releases/knative.dev/serving/cmd/autoscaler:$VERSION
docker save --output=autoscaler.tar gcr.io/knative-releases/knative.dev/serving/cmd/autoscaler:$VERSION

docker pull gcr.io/knative-releases/knative.dev/serving/cmd/controller@sha256:50fcbbc79b1078991280bf423e590c8904882dc8750c7f7d61bc06d944a052f2
docker tag gcr.io/knative-releases/knative.dev/serving/cmd/controller@sha256:50fcbbc79b1078991280bf423e590c8904882dc8750c7f7d61bc06d944a052f2 gcr.io/knative-releases/knative.dev/serving/cmd/controller:$VERSION
docker save --output=controller.tar gcr.io/knative-releases/knative.dev/serving/cmd/controller:$VERSION

docker pull gcr.io/knative-releases/knative.dev/serving/cmd/domain-mapping@sha256:3b7da888fafca8cc5ba5e2aa62d6f97751d50890ed9d0b01aabce66a7d26351e
docker tag gcr.io/knative-releases/knative.dev/serving/cmd/domain-mapping@sha256:3b7da888fafca8cc5ba5e2aa62d6f97751d50890ed9d0b01aabce66a7d26351e gcr.io/knative-releases/knative.dev/serving/cmd/domain-mapping:$VERSION
docker save --output=domain-mapping.tar gcr.io/knative-releases/knative.dev/serving/cmd/domain-mapping:$VERSION

docker pull gcr.io/knative-releases/knative.dev/serving/cmd/domain-mapping-webhook@sha256:a6529f0625483c81741c92895e4d54be8a103ecc5801e7c3aa049d3b3ea7cc90
docker tag gcr.io/knative-releases/knative.dev/serving/cmd/domain-mapping-webhook@sha256:a6529f0625483c81741c92895e4d54be8a103ecc5801e7c3aa049d3b3ea7cc90 gcr.io/knative-releases/knative.dev/serving/cmd/domain-mapping-webhook:$VERSION
docker save --output=domain-mapping-webhook.tar gcr.io/knative-releases/knative.dev/serving/cmd/domain-mapping-webhook:$VERSION

docker pull gcr.io/knative-releases/knative.dev/serving/cmd/webhook@sha256:382a1d64ea0686da2b973c95be96fcae29ac25b256b90f5735b1479a93d19c7a
docker tag gcr.io/knative-releases/knative.dev/serving/cmd/webhook@sha256:382a1d64ea0686da2b973c95be96fcae29ac25b256b90f5735b1479a93d19c7a gcr.io/knative-releases/knative.dev/serving/cmd/webhook:$VERSION
docker save --output=webhook.tar gcr.io/knative-releases/knative.dev/serving/cmd/webhook:$VERSION

# kong

docker pull docker.io/kong/kubernetes-ingress-controller:1.3
docker save --output=kongig.tar docker.io/kong/kubernetes-ingress-controller:1.3

docker pull docker.io/library/kong:2.5
docker save --output=konglib.tar docker.io/library/kong:2.5

# docker registry

docker pull registry:2.7.1
docker save --output=registry.tar registry:2.7.1

# database

docker pull postgres:13.4
docker save --output=postgres.tar postgres:13.4

# direktiv

PREFIX="localhost:5000"
VERSION="v0.5.4"

docker pull $PREFIX/flow
docker tag $PREFIX/flow vorteil/flow:$VERSION
docker save --output=flow.tar vorteil/flow

docker pull $PREFIX/init-pod
docker tag $PREFIX/init-pod vorteil/init-pod:$VERSION
docker save --output=init-pod.tar vorteil/init-pod

docker pull $PREFIX/secrets
docker tag $PREFIX/secrets vorteil/secrets:$VERSION
docker save --output=secrets.tar vorteil/secrets

docker pull $PREFIX/sidecar
docker tag $PREFIX/sidecar vorteil/sidecar:$VERSION
docker save --output=sidecar.tar vorteil/sidecar

docker pull $PREFIX/functions
docker tag $PREFIX/functions vorteil/functions:$VERSION
docker save --output=functions.tar vorteil/functions

docker pull $PREFIX/api
docker tag $PREFIX/api vorteil/api:$VERSION
docker save --output=api.tar vorteil/api

docker pull $PREFIX/ui
docker tag $PREFIX/ui vorteil/ui:$VERSION
docker save --output=ui.tar vorteil/ui

tar -cvzf images.tar.gz *.tar
