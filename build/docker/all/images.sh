#!/bin/bash

# k3s download

rm k3s
wget https://github.com/k3s-io/k3s/releases/download/v1.22.2-rc1%2Bk3s1/k3s

# knative
VERSION=1.0.0

docker pull gcr.io/knative-releases/knative.dev/serving/cmd/queue@sha256:6b642967b884ae8971c6fa1f0d3a436c06ccf90ec03da81f85e74768773eb290
docker tag gcr.io/knative-releases/knative.dev/serving/cmd/queue@sha256:6b642967b884ae8971c6fa1f0d3a436c06ccf90ec03da81f85e74768773eb290 gcr.io/knative-releases/knative.dev/serving/cmd/queue:$VERSION
docker save --output=queue.tar gcr.io/knative-releases/knative.dev/serving/cmd/queue:$VERSION

docker pull gcr.io/knative-releases/knative.dev/serving/cmd/activator@sha256:c54682f20c4758aabd9e9a45805ae382ef9de8a953026a308273fb94b71719d5
docker tag gcr.io/knative-releases/knative.dev/serving/cmd/activator@sha256:c54682f20c4758aabd9e9a45805ae382ef9de8a953026a308273fb94b71719d5 gcr.io/knative-releases/knative.dev/serving/cmd/activator:$VERSION
docker save --output=activator.tar gcr.io/knative-releases/knative.dev/serving/cmd/activator:$VERSION

docker pull gcr.io/knative-releases/knative.dev/serving/cmd/autoscaler@sha256:a2938d3c0e913b74b96f69845cdc09d4674a465a0895f71db9afe76d805db853
docker tag gcr.io/knative-releases/knative.dev/serving/cmd/autoscaler@sha256:a2938d3c0e913b74b96f69845cdc09d4674a465a0895f71db9afe76d805db853 gcr.io/knative-releases/knative.dev/serving/cmd/autoscaler:$VERSION
docker save --output=autoscaler.tar gcr.io/knative-releases/knative.dev/serving/cmd/autoscaler:$VERSION

docker pull gcr.io/knative-releases/knative.dev/serving/cmd/controller@sha256:d903707ec8c20f7a0a36852a0cf70062fbd23015d820f4ef085855de02e293ec
docker tag gcr.io/knative-releases/knative.dev/serving/cmd/controller@sha256:d903707ec8c20f7a0a36852a0cf70062fbd23015d820f4ef085855de02e293ec gcr.io/knative-releases/knative.dev/serving/cmd/controller:$VERSION
docker save --output=controller.tar gcr.io/knative-releases/knative.dev/serving/cmd/controller:$VERSION

docker pull gcr.io/knative-releases/knative.dev/serving/cmd/domain-mapping@sha256:d8754f853daefe201785ee4e3f71626bd5c010456259debe520ea0da78f04673
docker tag gcr.io/knative-releases/knative.dev/serving/cmd/domain-mapping@sha256:d8754f853daefe201785ee4e3f71626bd5c010456259debe520ea0da78f04673 gcr.io/knative-releases/knative.dev/serving/cmd/domain-mapping:$VERSION
docker save --output=domain-mapping.tar gcr.io/knative-releases/knative.dev/serving/cmd/domain-mapping:$VERSION

docker pull gcr.io/knative-releases/knative.dev/serving/cmd/domain-mapping-webhook@sha256:0b8fe2e03c4ce979d9f98df542ce48884b46cda098006fafa7db45b7f90ccfdb
docker tag gcr.io/knative-releases/knative.dev/serving/cmd/domain-mapping-webhook@sha256:0b8fe2e03c4ce979d9f98df542ce48884b46cda098006fafa7db45b7f90ccfdb gcr.io/knative-releases/knative.dev/serving/cmd/domain-mapping-webhook:$VERSION
docker save --output=domain-mapping-webhook.tar gcr.io/knative-releases/knative.dev/serving/cmd/domain-mapping-webhook:$VERSION

docker pull gcr.io/knative-releases/knative.dev/serving/cmd/webhook@sha256:86a5b8bb6cc0bd8cc9f02bf7035cff840c3543055578286013d59e8b4c313308
docker tag gcr.io/knative-releases/knative.dev/serving/cmd/webhook@sha256:86a5b8bb6cc0bd8cc9f02bf7035cff840c3543055578286013d59e8b4c313308 gcr.io/knative-releases/knative.dev/serving/cmd/webhook:$VERSION
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
VERSION="v0.6.0"

docker pull $PREFIX/flow
docker tag $PREFIX/flow direktiv/flow:$VERSION
rm -Rf flow.tar
docker save --output=flow.tar direktiv/flow:$VERSION

docker pull $PREFIX/init-pod
docker tag $PREFIX/init-pod direktiv/init-pod:$VERSION
rm -Rf init-pod.tar
docker save --output=init-pod.tar direktiv/init-pod:$VERSION

docker pull $PREFIX/secrets
docker tag $PREFIX/secrets direktiv/secrets:$VERSION
rm -Rf secrets.tar
docker save --output=secrets.tar direktiv/secrets:$VERSION

docker pull $PREFIX/sidecar
docker tag $PREFIX/sidecar direktiv/sidecar:$VERSION
rm -Rf sidecar.tar
docker save --output=sidecar.tar direktiv/sidecar:$VERSION

docker pull $PREFIX/functions
docker tag $PREFIX/functions direktiv/functions:$VERSION
rm -Rf functions.tar
docker save --output=functions.tar direktiv/functions:$VERSION

docker pull $PREFIX/api
docker tag $PREFIX/api direktiv/api:$VERSION
rm -Rf api.tar
docker save --output=api.tar direktiv/api:$VERSION

docker pull $PREFIX/ui
docker tag $PREFIX/ui direktiv/ui:$VERSION
rm -Rf ui.tar
docker save --output=ui.tar direktiv/ui:$VERSION

tar -cvzf images.tar.gz *.tar
