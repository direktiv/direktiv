#!/bin/bash

# k3s download

rm k3s
wget https://github.com/k3s-io/k3s/releases/download/v1.23.1-rc1%2Bk3s1/k3s

rm *.tar
rm *.tar.gz

# knative
VERSION=1.1.1

docker pull gcr.io/knative-releases/knative.dev/serving/cmd/queue@sha256:48a1753de35ecbe060611aea9e95751e3e4851183c4373e65aa1b9410ea6e263
docker tag gcr.io/knative-releases/knative.dev/serving/cmd/queue@sha256:48a1753de35ecbe060611aea9e95751e3e4851183c4373e65aa1b9410ea6e263 gcr.io/knative-releases/knative.dev/serving/cmd/queue:$VERSION
docker save --output=queue.tar gcr.io/knative-releases/knative.dev/serving/cmd/queue:$VERSION

docker pull gcr.io/knative-releases/knative.dev/serving/cmd/activator@sha256:ba1485ded12049525afb9856c2fa10d613dbc2b2da90556116bf257f2128eaae
docker tag gcr.io/knative-releases/knative.dev/serving/cmd/activator@sha256:ba1485ded12049525afb9856c2fa10d613dbc2b2da90556116bf257f2128eaae gcr.io/knative-releases/knative.dev/serving/cmd/activator:$VERSION
docker save --output=activator.tar gcr.io/knative-releases/knative.dev/serving/cmd/activator:$VERSION

docker pull gcr.io/knative-releases/knative.dev/serving/cmd/autoscaler@sha256:dca8258a46dd225b8a72dfe63e8971b23876458f6f64b4ad82792c4d6e470bdc
docker tag gcr.io/knative-releases/knative.dev/serving/cmd/autoscaler@sha256:dca8258a46dd225b8a72dfe63e8971b23876458f6f64b4ad82792c4d6e470bdc gcr.io/knative-releases/knative.dev/serving/cmd/autoscaler:$VERSION
docker save --output=autoscaler.tar gcr.io/knative-releases/knative.dev/serving/cmd/autoscaler:$VERSION

docker pull gcr.io/knative-releases/knative.dev/serving/cmd/controller@sha256:2765feeaa3958827388e6f5119010ee08c0eec9ad7518bb38ac4b9a4355d87fb
docker tag gcr.io/knative-releases/knative.dev/serving/cmd/controller@sha256:2765feeaa3958827388e6f5119010ee08c0eec9ad7518bb38ac4b9a4355d87fb gcr.io/knative-releases/knative.dev/serving/cmd/controller:$VERSION
docker save --output=controller.tar gcr.io/knative-releases/knative.dev/serving/cmd/controller:$VERSION

docker pull gcr.io/knative-releases/knative.dev/serving/cmd/domain-mapping@sha256:25df5b854d28dac69c6293db4db50d8fa819c96ad2f2a30bdde6aad467de1b17
docker tag gcr.io/knative-releases/knative.dev/serving/cmd/domain-mapping@sha256:25df5b854d28dac69c6293db4db50d8fa819c96ad2f2a30bdde6aad467de1b17 gcr.io/knative-releases/knative.dev/serving/cmd/domain-mapping:$VERSION
docker save --output=domain-mapping.tar gcr.io/knative-releases/knative.dev/serving/cmd/domain-mapping:$VERSION

docker pull gcr.io/knative-releases/knative.dev/serving/cmd/domain-mapping-webhook@sha256:6ccc1f6ac07d27e97d96c502b4c6e928d5fb3abd165ae7670e94a57788416c75
docker tag gcr.io/knative-releases/knative.dev/serving/cmd/domain-mapping-webhook@sha256:6ccc1f6ac07d27e97d96c502b4c6e928d5fb3abd165ae7670e94a57788416c75 gcr.io/knative-releases/knative.dev/serving/cmd/domain-mapping-webhook:$VERSION
docker save --output=domain-mapping-webhook.tar gcr.io/knative-releases/knative.dev/serving/cmd/domain-mapping-webhook:$VERSION

docker pull gcr.io/knative-releases/knative.dev/serving/cmd/webhook@sha256:9f3c83def9d0d5de0e8e1d1f4c10f262e283fe12d21dcbb91de06b65d3bd08ad
docker tag gcr.io/knative-releases/knative.dev/serving/cmd/webhook@sha256:9f3c83def9d0d5de0e8e1d1f4c10f262e283fe12d21dcbb91de06b65d3bd08ad gcr.io/knative-releases/knative.dev/serving/cmd/webhook:$VERSION
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
VERSION="v1-all-in-one"

docker pull $PREFIX/flow
docker tag $PREFIX/flow localhost:31212/flow:$VERSION
rm -Rf flow.tar
docker save --output=flow.tar localhost:31212/flow:$VERSION

docker pull $PREFIX/flow-dbinit
docker tag $PREFIX/flow-dbinit localhost:31212/flow-dbinit:$VERSION
rm -Rf flow-dbinit.tar
docker save --output=flow-dbinit.tar localhost:31212/flow-dbinit:$VERSION

docker pull $PREFIX/init-pod
docker tag $PREFIX/init-pod localhost:31212/init-pod:$VERSION
rm -Rf init-pod.tar
docker save --output=init-pod.tar localhost:31212/init-pod:$VERSION

docker pull $PREFIX/secrets
docker tag $PREFIX/secrets localhost:31212/secrets:$VERSION
rm -Rf secrets.tar
docker save --output=secrets.tar localhost:31212/secrets:$VERSION

docker pull $PREFIX/sidecar
docker tag $PREFIX/sidecar localhost:31212/sidecar:$VERSION
rm -Rf sidecar.tar
docker save --output=sidecar.tar localhost:31212/sidecar:$VERSION

docker pull $PREFIX/functions
docker tag $PREFIX/functions localhost:31212/functions:$VERSION
rm -Rf functions.tar
docker save --output=functions.tar localhost:31212/functions:$VERSION

docker pull $PREFIX/api
docker tag $PREFIX/api localhost:31212/api:$VERSION
rm -Rf api.tar
docker save --output=api.tar localhost:31212/api:$VERSION

docker pull $PREFIX/ui
docker tag $PREFIX/ui localhost:31212/ui:$VERSION
rm -Rf ui.tar
docker save --output=ui.tar localhost:31212/ui:$VERSION

tar -cvzf images.tar.gz *.tar




# docker pull $PREFIX/flow
# docker tag $PREFIX/flow direktiv/flow:$VERSION
# rm -Rf flow.tar
# docker save --output=flow.tar direktiv/flow:$VERSION

# docker pull $PREFIX/flow-dbinit
# docker tag $PREFIX/flow-dbinit direktiv/flow-dbinit:$VERSION
# rm -Rf flow-dbinit.tar
# docker save --output=flow-dbinit.tar direktiv/flow-dbinit:$VERSION

# docker pull $PREFIX/init-pod
# docker tag $PREFIX/init-pod direktiv/init-pod:$VERSION
# rm -Rf init-pod.tar
# docker save --output=init-pod.tar direktiv/init-pod:$VERSION

# docker pull $PREFIX/secrets
# docker tag $PREFIX/secrets direktiv/secrets:$VERSION
# rm -Rf secrets.tar
# docker save --output=secrets.tar direktiv/secrets:$VERSION

# docker pull $PREFIX/sidecar
# docker tag $PREFIX/sidecar direktiv/sidecar:$VERSION
# rm -Rf sidecar.tar
# docker save --output=sidecar.tar direktiv/sidecar:$VERSION

# docker pull $PREFIX/functions
# docker tag $PREFIX/functions direktiv/functions:$VERSION
# rm -Rf functions.tar
# docker save --output=functions.tar direktiv/functions:$VERSION

# docker pull $PREFIX/api
# docker tag $PREFIX/api direktiv/api:$VERSION
# rm -Rf api.tar
# docker save --output=api.tar direktiv/api:$VERSION

# docker pull $PREFIX/ui
# docker tag $PREFIX/ui direktiv/ui:$VERSION
# rm -Rf ui.tar
# docker save --output=ui.tar direktiv/ui:$VERSION

# tar -cvzf images.tar.gz *.tar


