# direktiv

direktiv helm chart

![Version: 0.1.0](https://img.shields.io/badge/Version-0.1.0-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: v0.4.1](https://img.shields.io/badge/AppVersion-v0.4.1-informational?style=flat-square)

## Additional Information

.Values.functions.namespace

Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore
et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut
aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse
cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in
culpa qui officia deserunt mollit anim id est laborum.

## Installing the Chart

To install the chart with the release name `my-release`:

```console
$ helm repo add foo-bar http://charts.foo-bar.com
$ helm install my-release foo-bar/direktiv
```

## Requirements

| Repository | Name | Version |
|------------|------|---------|
| https://charts.bitnami.com/bitnami | kube-prometheus | 6.1.8 |
| https://operator.min.io/ | minio-operator | 4.2.7 |

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| affinity | object | `{}` |  |
| database.host | string | `"yb-tservers.yugabyte"` | database host |
| database.name | string | `"direktiv"` | database name, auto created if it does not exist |
| database.password | string | `"direktivdirektiv"` | database password |
| database.port | int | `5433` | database port |
| database.sslmode | string | `"require"` | sslmode for database |
| database.user | string | `"direktiv"` | database user |
| debug | bool | `false` | enable debug across all direktiv components |
| flow.extraContainers | list | `[]` | extra container in flow pod |
| flow.extraVolumeMounts | string | `nil` | extra volume mounts in flow pod |
| flow.extraVolumes | string | `nil` | extra volumes in flow pod |
| flow.image | string | `"vorteil/flow"` | image for flow pod |
| flow.replicas | int | `1` | number of flow replicas |
| flow.tag | string | `""` | image tag for flow pod |
| fluentbit.extraConfig | string | `""` | postgres for direktiv services Append extra output to fluentbit configuration. There are two log types: application (system), functions (workflows) these can be matched to new outputs. |
| functions.certificate | string | `"none"` |  |
| functions.db | string | `""` |  |
| functions.extraContainers | list | `[]` |  |
| functions.extraContainersPod | list | `[]` |  e.g. database containers for google cloud or logging |
| functions.image | string | `"vorteil/functions"` |  |
| functions.initPodCertificate | string | `"none"` |  |
| functions.initPodImage | string | `"vorteil/direktiv-init-pod"` |  |
| functions.mtls | string | `"none"` |  |
| functions.namespace | string | `"direktiv-services-direktiv"` |  |
| functions.netShape | string | `"10M"` |  |
| functions.podCleaner | bool | `true` |  |
| functions.replicaCount | int | `1` |  |
| functions.runtime | string | `"default"` |  |
| functions.sidecar | string | `"vorteil/sidecar"` |  |
| functions.tag | string | `""` |  |
| http_proxy | string | `""` | http proxy settings |
| https_proxy | string | `""` | https proxy settings |
| imagePullSecrets | list | `[]` |  |
| ingress.certificate | string | `"none"` |  |
| ingress.class | string | `"kong"` |  |
| ingress.host | string | `""` |  |
| ingress.timeout | string | `"900s"` |  |
| kube-prometheus.alertmanager.enabled | bool | `false` |  |
| kube-prometheus.exporters.coreDns.enabled | bool | `false` |  |
| kube-prometheus.exporters.kube-state-metrics.enabled | bool | `false` |  |
| kube-prometheus.exporters.kubeApiServer.enabled | bool | `false` |  |
| kube-prometheus.exporters.kubeControllerManager.enabled | bool | `false` |  |
| kube-prometheus.exporters.kubeProxy.enabled | bool | `false` |  |
| kube-prometheus.exporters.kubeScheduler.enabled | bool | `false` |  |
| kube-prometheus.exporters.kubelet.enabled | bool | `false` |  |
| kube-prometheus.exporters.node-exporter.enabled | bool | `false` |  |
| kube-prometheus.operator.enabled | bool | `true` |  |
| kube-prometheus.prometheus.enabled | bool | `true` |  |
| kube-prometheus.prometheus.replicaCount | int | `1` |  |
| kube-prometheus.prometheus.thanos.create | bool | `false` |  |
| minio-operator.tenants[0].name | string | `"direktiv-tenant"` |  |
| minio-operator.tenants[0].pools[0] | object | `{"servers":1,"size":"1Gi","storageClassName":"local-path","volumesPerServer":4}` | set to 4 for HA |
| minio-operator.tenants[0].pools[0].storageClassName | string | `"local-path"` | storage class to use. k3s uses local-path |
| minio-operator.tenants[0].secrets.accessKey | string | `"minio"` |  |
| minio-operator.tenants[0].secrets.enabled | bool | `true` |  |
| minio-operator.tenants[0].secrets.name | string | `"minio-secret"` |  |
| minio-operator.tenants[0].secrets.secretKey | string | `"minio123"` |  |
| networkPolicies.db | string | `"0.0.0.0/0"` |  |
| networkPolicies.enabled | bool | `false` |  |
| networkPolicies.podCidr | string | `"0.0.0.0/0"` |  |
| networkPolicies.serviceCidr | string | `"0.0.0.0/0"` |  |
| no_proxy | string | `""` | no proxy proxy settings |
| nodeSelector | object | `{}` |  |
| prometheus.enabled | bool | `true` |  |
| prometheus.enabled | bool | `true` |  |
| prometheus.image | string | `"prom/prometheus"` |  |
| prometheus.queryAddress | string | `"http://direktiv-prometheus-service.default:9090"` |  |
| prometheus.targetAnnotations."prometheus.io/path" | string | `"/metrics"` |  |
| prometheus.targetAnnotations."prometheus.io/port" | string | `"2112"` |  |
| prometheus.targetAnnotations."prometheus.io/scrape" | string | `"true"` |  |
| pullPolicy | string | `"Always"` |  |
| registry | string | `"docker.io"` |  |
| replicaCount | int | `1` |  |
| secrets.db | string | `""` |  |
| secrets.extraVolumeMounts | list | `[]` |  |
| secrets.image | string | `"vorteil/secrets"` |  |
| secrets.tag | string | `""` |  |
| serviceAccount | object | `{"annotations":{},"name":""}` | service account for flow component |
| thanos.enabled | bool | `true` |  |
| tolerations | list | `[]` |  |
| ui.certificate | string | `"none"` |  |
| ui.extraContainers | list | `[]` |  |
| ui.image | string | `"vorteil/direktiv-ui"` |  |
| ui.tag | string | `""` |  |
| withUI | bool | `true` |    enabled: false   minReplicas: 1   maxReplicas: 10   targetCPUUtilizationPercentage: 80   targetMemoryUtilizationPercentage: 80 support services |

