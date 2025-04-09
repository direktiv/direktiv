#!/bin/sh

docker pull gcr.io/knative-releases/knative.dev/serving/cmd/activator:v1.16.3
docker pull gcr.io/knative-releases/knative.dev/serving/cmd/autoscaler:v1.16.3
docker pull gcr.io/knative-releases/knative.dev/serving/cmd/controller:v1.16.3
docker pull gcr.io/knative-releases/knative.dev/serving/cmd/webhook:v1.16.3
docker pull gcr.io/knative-releases/knative.dev/serving/cmd/queue:v1.16.3
docker pull gcr.io/knative-releases/knative.dev/serving/cmd/autoscaler-hpa:v1.16.3
docker pull gcr.io/knative-releases/knative.dev/serving/pkg/cleanup/cmd/cleanup:v1.16.3
docker pull gcr.io/knative-releases/knative.dev/pkg/apiextensions/storageversion/cmd/migrate:v1.17.4


docker tag gcr.io/knative-releases/knative.dev/serving/cmd/activator:v1.16.3 direktiv/activator:v1.16.3
docker tag gcr.io/knative-releases/knative.dev/serving/cmd/autoscaler:v1.16.3 direktiv/autoscaler:v1.16.3
docker tag gcr.io/knative-releases/knative.dev/serving/cmd/controller:v1.16.3 direktiv/controller:v1.16.3
docker tag gcr.io/knative-releases/knative.dev/serving/cmd/webhook:v1.16.3 direktiv/webhook:v1.16.3
docker tag gcr.io/knative-releases/knative.dev/serving/cmd/queue:v1.16.3 direktiv/queue:v1.16.3
docker tag gcr.io/knative-releases/knative.dev/serving/cmd/autoscaler-hpa:v1.16.3 direktiv/autoscaler-hpa:v1.16.3
docker tag gcr.io/knative-releases/knative.dev/serving/pkg/cleanup/cmd/cleanup:v1.16.3 direktiv/cleanup:v1.16.3
docker tag gcr.io/knative-releases/knative.dev/pkg/apiextensions/storageversion/cmd/migrate:v1.17.4 direktiv/migrate:v1.16.3

docker push direktiv/activator:v1.16.3 
docker push direktiv/autoscaler:v1.16.3 
docker push direktiv/controller:v1.16.3
docker push direktiv/webhook:v1.16.3 
docker push direktiv/queue:v1.16.3 
docker push direktiv/autoscaler-hpa:v1.16.3 
docker push direktiv/cleanup:v1.16.3
docker push direktiv/migrate:v1.16.3



