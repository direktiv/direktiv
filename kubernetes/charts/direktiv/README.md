# direktiv

direktiv helm chart

![Version: 0.1.0](https://img.shields.io/badge/Version-0.1.0-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: v0.5.8](https://img.shields.io/badge/AppVersion-v0.5.8-informational?style=flat-square)

## Additional Information

This chart installs direktiv.

## Installing the Chart

To install the chart with the release name `direktiv`:

```console
$ kubectl create ns direktiv-services-direktiv
$ helm repo add direktiv https://charts.direktiv.io
$ helm install direktiv direktiv/direktiv
```

## Requirements

| Repository | Name | Version |
|------------|------|---------|
| https://charts.bitnami.com/bitnami | thanos | 6.0.1 |
| https://operator.min.io/ | minio-operator | 4.2.7 |
| https://prometheus-community.github.io/helm-charts | prometheus | 14.7.1 |

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| affinity | object | `{}` |  |
| api.extraContainers | list | `[]` | extra container in api pod |
| api.extraContainers | list | `[]` |  |
| api.extraVolumeMounts | string | `nil` | extra volume mounts in api pod |
| api.extraVolumes | string | `nil` | extra volumes in api pod |
| api.image | string | `"direktiv/api"` | image for api pod |
| api.kongPlugins | string | `"none"` | Kong plugins to enable |
| api.replicas | int | `1` |  |
| api.tag | string | `""` | image tag for api pod |
| apikey | string | `""` | api key, value 'apikey' required in header |
| database.host | string | `"postgres-postgresql-ha-pgpool.postgres"` | database host |
| database.name | string | `"direktiv"` | database name, auto created if it does not exist |
| database.password | string | `"direktivdirektiv"` | database password |
| database.port | int | `5432` | database port |
| database.sslmode | string | `"require"` | sslmode for database |
| database.user | string | `"direktiv"` | database user |
| debug | bool | `false` | enable debug across all direktiv components |
| encryptionKey | string | `""` |  if set to empty, one will be generated on install |
| eventing | object | `{"enabled":false}` | knative eventing enabled, requires knative setup and configuration |
| flow.extraContainers | list | `[]` | extra container in flow pod |
| flow.extraVolumeMounts | string | `nil` | extra volume mounts in flow pod |
| flow.extraVolumes | string | `nil` | extra volumes in flow pod |
| flow.image | string | `"direktiv/flow"` | image for flow pod |
| flow.replicas | int | `1` | number of flow replicas |
| flow.tag | string | `""` | image tag for flow pod |
| fluentbit.extraConfig | string | `""` | postgres for direktiv services Append extra output to fluentbit configuration. There are two log types: application (system), functions (workflows) these can be matched to new outputs. |
| functions.extraContainers | list | `[]` | extra containers for tasks and knative pods |
| functions.extraContainersPod | list | `[]` | extra containers for function controller, e.g. database containers for google cloud or logging |
| functions.extraVolumes | list | `[]` | extra volumes for tasks and knative pods |
| functions.http_proxy | string | `""` | http_proxy injected as environment variable in functions |
| functions.https_proxy | string | `""` | https_proxy injected as environment variable in functions |
| functions.image | string | `"direktiv/functions"` |  |
| functions.ingressClass | string | `"kong-internal"` |  |
| functions.initPodImage | string | `"direktiv/direktiv-init-pod"` |  |
| functions.namespace | string | `"direktiv-services-direktiv"` |  |
| functions.netShape | string | `"10M"` | Egress/Ingress network limit for functions if supported by network |
| functions.no_proxy | string | `""` | no_proxy injected as environment variable in functions |
| functions.podCleaner | bool | `true` | Cleaning up tasks, Kubernetes < 1.20 does not clean finished tasks |
| functions.replicaCount | int | `1` | number of controller replicas |
| functions.runtime | string | `"default"` | runtime to use, e.g. gvisor on GCP |
| functions.sidecar | string | `"direktiv/sidecar"` |  |
| functions.tag | string | `""` |  |
| functions.timeout | int | `900000` |  |
| http_proxy | string | `""` | http proxy settings |
| https_proxy | string | `""` | https proxy settings |
| imagePullSecrets | list | `[]` |  |
| ingress.certificate | string | `"none"` | TLS secret |
| ingress.class | string | `"kong"` | ingress class |
| ingress.host | string | `""` | host for external services, only required for TLS |
| ingress.timeout | int | `900000` | timeout for /api route |
| logging | string | `"json"` | json or console logger |
| minio-operator.enabled | bool | `false` |  |
| minio-operator.tenants[0].name | string | `"direktiv-tenant"` |  |
| minio-operator.tenants[0].pools[0] | object | `{"servers":1,"size":"1Gi","storageClassName":"local-path","volumesPerServer":4}` | set to 4 for HA |
| minio-operator.tenants[0].pools[0].storageClassName | string | `"local-path"` | storage class to use. k3s uses local-path |
| minio-operator.tenants[0].secrets.accessKey | string | `"minio"` |  |
| minio-operator.tenants[0].secrets.enabled | bool | `true` |  |
| minio-operator.tenants[0].secrets.name | string | `"minio-secret"` |  |
| minio-operator.tenants[0].secrets.secretKey | string | `"minio123"` |  |
| networkPolicies.db | string | `"0.0.0.0/0"` | CIDR for database, excempt from policies |
| networkPolicies.enabled | bool | `false` | adds network policies |
| networkPolicies.podCidr | string | `"0.0.0.0/0"` | CIDR for pods, excempt from policies |
| networkPolicies.serviceCidr | string | `"0.0.0.0/0"` | CIDR for services, excempt from policies |
| no_proxy | string | `""` | no proxy proxy settings |
| nodeSelector | object | `{}` |  |
| opentelemetry.address | string | `"localhost:4317"` | opentelemetry address where Direktiv is sending data to |
| opentelemetry.agentconfig | string | `"receivers:\n  otlp:\n    protocols:\n      grpc:\n      http:\nexporters:\n  otlp:\n    endpoint: \"192.168.1.113:14250\"\n    insecure: true\n    sending_queue:\n      num_consumers: 4\n      queue_size: 100\n    retry_on_failure:\n      enabled: true\n  logging:\n    loglevel: debug\nprocessors:\n  batch:\n  memory_limiter:\n    # Same as --mem-ballast-size-mib CLI argument\n    ballast_size_mib: 165\n    # 80% of maximum memory up to 2G\n    limit_mib: 400\n    # 25% of limit up to 2G\n    spike_limit_mib: 100\n    check_interval: 5s\nextensions:\n  zpages: {}\nservice:\n  extensions: [zpages]\n  pipelines:\n    traces:\n      receivers: [otlp]\n      processors: [memory_limiter, batch]\n      exporters: [logging, otlp]\n"` | config for sidecar agent |
| opentelemetry.enabled | bool | `false` | installs opentelemtry agent as sidecar in flow |
| prometheus.alertmanager.enabled | bool | `false` |  |
| prometheus.global.evaluation_interval | string | `"1m"` |  |
| prometheus.global.scrape_interval | string | `"1m"` |  |
| prometheus.kubeStateMetrics.enabled | bool | `false` |  |
| prometheus.nodeExporter.enabled | bool | `false` |  |
| prometheus.pushgateway.enabled | bool | `false` |  |
| prometheus.server.persistentVolume.enabled | bool | `false` |  |
| prometheus.server.retention | string | `"96h"` |  |
| prometheus.serviceAccounts.alertmanager.create | bool | `false` |  |
| prometheus.serviceAccounts.nodeExporter.create | bool | `false` |  |
| prometheus.serviceAccounts.pushgateway.create | bool | `false` |  |
| prometheus.serviceAccounts.server.create | bool | `true` |  |
| pullPolicy | string | `"Always"` |  |
| registry | string | `"docker.io"` |  |
| secrets | object | `{"db":"","extraVolumeMounts":[],"image":"direktiv/secrets","tag":""}` | secrets sidecar in flow pod |
| serviceAccount | object | `{"annotations":{},"name":""}` | service account for flow component |
| thanos.bucketweb.enabled | bool | `true` |  |
| thanos.compactor.enabled | bool | `true` |  |
| thanos.compactor.persistence.storageClass | string | `"local-path"` |  |
| thanos.enabled | bool | `false` | install Thanos |
| thanos.enabled | bool | `false` |  |
| thanos.global.storageClass | string | `"local-path"` |  |
| thanos.objstoreConfig | string | `"type: s3\nconfig:\n  bucket: thanos\n  endpoint: direktiv-tenant-console.{{ .Release.Namespace }}.svc.cluster.local:9000\n  access_key: minio\n  secret_key: minio123\n  insecure: true"` |  |
| thanos.query.dnsDiscovery.sidecarsNamespace | string | `"{{ .Release.Namespace }}"` |  |
| thanos.query.dnsDiscovery.sidecarsService | string | `"{{ .Release.Namespace }}-kube-prometheus-prometheus"` |  |
| thanos.ruler.enabled | bool | `false` |  |
| thanos.storegateway.enabled | bool | `true` |  |
| thanos.storegateway.persistence.storageClass | string | `"local-path"` |  |
| timeout | int | `900000` | api timeouts |
| tolerations | list | `[]` |  |
| ui | object | `{"certificate":"none","extraContainers":[],"image":"direktiv/ui","kongPlugins":"none","tag":""}` | UI configuration |

