# direktiv

direktiv helm chart

![Version: 0.1.0](https://img.shields.io/badge/Version-0.1.0-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: v0.4.1](https://img.shields.io/badge/AppVersion-v0.4.1-informational?style=flat-square)

## Additional Information

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

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| affinity | object | `{}` |  |
| api.certificate | string | `"none"` |  |
| api.extraContainers | list | `[]` |  |
| api.image | string | `"vorteil/api"` |  |
| api.key | string | `""` |  |
| api.tag | string | `""` |  |
| autoscaling.enabled | bool | `false` |  |
| autoscaling.maxReplicas | int | `10` |  |
| autoscaling.minReplicas | int | `1` |  |
| autoscaling.targetCPUUtilizationPercentage | int | `80` |  |
| autoscaling.targetMemoryUtilizationPercentage | int | `80` |  |
| database.host | string | `"yb-tservers.yugabyte"` | database host |
| database.name | string | `"direktiv"` | database name, auto created if it does not exist |
| database.password | string | `"direktiv"` | database password |
| database.port | int | `5433` | database port |
| database.sslmode | string | `"require"` | sslmode for database |
| database.user | string | `"direktiv"` | database user |
| debug | bool | `false` | enable debug across all direktiv components |
| flow.certificates.flow | string | `"none"` |  |
| flow.certificates.ingress | string | `"none"` |  |
| flow.certificates.mtlsFlow | string | `"none"` |  |
| flow.certificates.mtlsIngress | string | `"none"` |  |
| flow.db | string | `""` |  |
| flow.extraContainers | list | `[]` |  |
| flow.extraVolumeMounts | string | `nil` |  |
| flow.extraVolumes | string | `nil` |    mountPath: /etc/config |
| flow.functionsCA | string | `"none"` |  |
| flow.functionsProtocol | string | `"http"` |  |
| flow.image | string | `"vorteil/flow"` |  |
| flow.tag | string | `""` |  |
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
| grpc.client.maxRecvSize | string | `"4194304"` |  |
| grpc.client.maxSendSize | string | `"4194304"` |  |
| grpc.server.maxRecvSize | string | `"4194304"` |  |
| grpc.server.maxSendSize | string | `"4194304"` |  |
| http_proxy | string | `""` |  |
| https_proxy | string | `""` |  |
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
| kube-prometheus.operator.enabled | bool | `false` |  |
| minio.enabled | bool | `false` |  |
| networkPolicies.db | string | `"0.0.0.0/0"` |  |
| networkPolicies.enabled | bool | `false` |  |
| networkPolicies.podCidr | string | `"0.0.0.0/0"` |  |
| networkPolicies.serviceCidr | string | `"0.0.0.0/0"` |  |
| no_proxy | string | `""` |  |
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
| secrets.key | string | `"01234567890123456789012345678912"` |  |
| secrets.tag | string | `""` |  |
| serviceAccount | object | `{"annotations":{},"name":""}` | service account for flow component |
| supportPersist | bool | `false` |  |
| thanos.enabled | bool | `true` |  |
| tolerations | list | `[]` |  |
| ui.certificate | string | `"none"` |  |
| ui.extraContainers | list | `[]` |  |
| ui.image | string | `"vorteil/direktiv-ui"` |  |
| ui.tag | string | `""` |  |
| withAPI | bool | `true` |  |
| withUI | bool | `true` |  |

