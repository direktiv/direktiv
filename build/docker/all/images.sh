#!/bin/bash

# k3s download

rm k3s
wget https://github.com/k3s-io/k3s/releases/download/v1.23.1-rc1%2Bk3s1/k3s

rm *.tar
rm *.tar.gz

# knative
VERSION=1.1.1

docker pull gcr.io/knative-releases/knative.dev/serving/cmd/queue@sha256:2fd485778a47e9f6c38e6198715c08ce77b49ffd70363d99ca18a00fac7bf1d8
docker tag gcr.io/knative-releases/knative.dev/serving/cmd/queue@sha256:2fd485778a47e9f6c38e6198715c08ce77b49ffd70363d99ca18a00fac7bf1d8 gcr.io/knative-releases/knative.dev/serving/cmd/queue:$VERSION
docker save --output=queue.tar gcr.io/knative-releases/knative.dev/serving/cmd/queue:$VERSION

docker pull gcr.io/knative-releases/knative.dev/serving/cmd/activator@sha256:ebc3a26353b38a0c1af1fd5fc3715acfadbdbdef13752738e55611cf9574e386
docker tag gcr.io/knative-releases/knative.dev/serving/cmd/activator@sha256:ebc3a26353b38a0c1af1fd5fc3715acfadbdbdef13752738e55611cf9574e386 gcr.io/knative-releases/knative.dev/serving/cmd/activator:$VERSION
docker save --output=activator.tar gcr.io/knative-releases/knative.dev/serving/cmd/activator:$VERSION

docker pull gcr.io/knative-releases/knative.dev/serving/cmd/autoscaler@sha256:9d1811291955d213dc9d362c5a8e80463eb40e15c3a631f1538b762d4cd331ec
docker tag gcr.io/knative-releases/knative.dev/serving/cmd/autoscaler@sha256:9d1811291955d213dc9d362c5a8e80463eb40e15c3a631f1538b762d4cd331ec gcr.io/knative-releases/knative.dev/serving/cmd/autoscaler:$VERSION
docker save --output=autoscaler.tar gcr.io/knative-releases/knative.dev/serving/cmd/autoscaler:$VERSION

docker pull gcr.io/knative-releases/knative.dev/serving/cmd/controller@sha256:13408b713d2c4ed3a2aea440f1da691626946081223ef74f3079a8c040d1c022
docker tag gcr.io/knative-releases/knative.dev/serving/cmd/controller@sha256:13408b713d2c4ed3a2aea440f1da691626946081223ef74f3079a8c040d1c022 gcr.io/knative-releases/knative.dev/serving/cmd/controller:$VERSION
docker save --output=controller.tar gcr.io/knative-releases/knative.dev/serving/cmd/controller:$VERSION

docker pull gcr.io/knative-releases/knative.dev/serving/cmd/domain-mapping@sha256:3c23df2ad0634057a8311393ba61a9ceec4f3e0898a2aa3dcd7e2ed6caf17548
docker tag gcr.io/knative-releases/knative.dev/serving/cmd/domain-mapping@sha256:3c23df2ad0634057a8311393ba61a9ceec4f3e0898a2aa3dcd7e2ed6caf17548 gcr.io/knative-releases/knative.dev/serving/cmd/domain-mapping:$VERSION
docker save --output=domain-mapping.tar gcr.io/knative-releases/knative.dev/serving/cmd/domain-mapping:$VERSION

docker pull gcr.io/knative-releases/knative.dev/serving/cmd/domain-mapping-webhook@sha256:9822a9b1c55568886eb0895c0d3c963c7dd5e30aaaf9679bc17add22e00363e0
docker tag gcr.io/knative-releases/knative.dev/serving/cmd/domain-mapping-webhook@sha256:9822a9b1c55568886eb0895c0d3c963c7dd5e30aaaf9679bc17add22e00363e0 gcr.io/knative-releases/knative.dev/serving/cmd/domain-mapping-webhook:$VERSION
docker save --output=domain-mapping-webhook.tar gcr.io/knative-releases/knative.dev/serving/cmd/domain-mapping-webhook:$VERSION

docker pull gcr.io/knative-releases/knative.dev/serving/cmd/webhook@sha256:3d46c616305acba3993cc1de9c2e35f3d680a49e2043996fa38f5ad03d5ef805
docker tag gcr.io/knative-releases/knative.dev/serving/cmd/webhook@sha256:3d46c616305acba3993cc1de9c2e35f3d680a49e2043996fa38f5ad03d5ef805 gcr.io/knative-releases/knative.dev/serving/cmd/webhook:$VERSION
docker save --output=webhook.tar gcr.io/knative-releases/knative.dev/serving/cmd/webhook:$VERSION

# contour

docker pull gcr.io/knative-releases/github.com/projectcontour/contour/cmd/contour@sha256:5f726d901a2852197447b5d0ca43d7d0b3bb0756290fbd7984371ab5a49db853
docker tag gcr.io/knative-releases/github.com/projectcontour/contour/cmd/contour@sha256:5f726d901a2852197447b5d0ca43d7d0b3bb0756290fbd7984371ab5a49db853  gcr.io/knative-releases/github.com/projectcontour/contour/cmd/contour:$VERSION
docker save --output=contour.tar gcr.io/knative-releases/github.com/projectcontour/contour/cmd/contour:$VERSION

docker pull gcr.io/knative-releases/knative.dev/net-contour/cmd/controller@sha256:922ce3f28a1dc618e4ebd62cfdf10216f06543bb70e280277107c4ec3d2e4eac
docker tag gcr.io/knative-releases/knative.dev/net-contour/cmd/controller@sha256:922ce3f28a1dc618e4ebd62cfdf10216f06543bb70e280277107c4ec3d2e4eac  gcr.io/knative-releases/github.com/projectcontour/contour/cmd/controller:$VERSION
docker save --output=contour-controller.tar gcr.io/knative-releases/github.com/projectcontour/contour/cmd/controller:$VERSION

docker pull docker.io/envoyproxy/envoy:v1.21.0
docker save --output=envoy.tar docker.io/envoyproxy/envoy:v1.21.0

# docker registry

docker pull registry:2.7.1
docker save --output=registry.tar registry:2.7.1

# database

docker pull postgres:13.4
docker save --output=postgres.tar postgres:13.4

# nginx (k3s crictl inspecti <IMAGEID>)

docker pull k8s.gcr.io/ingress-nginx/controller@sha256:f766669fdcf3dc26347ed273a55e754b427eb4411ee075a53f30718b4499076a
docker tag k8s.gcr.io/ingress-nginx/controller@sha256:f766669fdcf3dc26347ed273a55e754b427eb4411ee075a53f30718b4499076a  k8s.gcr.io/ingress-nginx/controller:$VERSION
docker save --output=nginx-controller.tar k8s.gcr.io/ingress-nginx/controller:$VERSION

docker pull k8s.gcr.io/ingress-nginx/kube-webhook-certgen@sha256:64d8c73dca984af206adf9d6d7e46aa550362b1d7a01f3a0a91b20cc67868660
docker tag k8s.gcr.io/ingress-nginx/kube-webhook-certgen@sha256:64d8c73dca984af206adf9d6d7e46aa550362b1d7a01f3a0a91b20cc67868660  k8s.gcr.io/ingress-nginx/kube-webhook-certgen:$VERSION
docker save --output=nginx-webhook.tar k8s.gcr.io/ingress-nginx/kube-webhook-certgen:$VERSION

# direktiv

PREFIX="localhost:5000"
VERSION="v0.5.10"

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
