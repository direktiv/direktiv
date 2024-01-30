# direktiv

direktiv helm chart

![Version: 0.1.22-rc1](https://img.shields.io/badge/Version-0.1.22--rc1-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: v0.8.1-rc1](https://img.shields.io/badge/AppVersion-v0.8.1--rc1-informational?style=flat-square)

## Additional Information

This chart installs direktiv.

### Changes in 0.1.21
* Changed ui image name to frontend
* Changed ingress configuration for API and frontend
* Version upgrade

### Changes in 0.1.20

* Direktiv version upgrade
* Resources CPU/Memory configurable

### Changes in 0.1.19

* Fixed image commands

### Changes in 0.1.18

* Fixed image references

### Changes in 0.1.17

*Direktiv version upgrade*

### Changes in 0.1.16

*Direktiv version upgrade*

### Changes in 0.1.15

* Deprecated unused function configuration items
* Added disk size to small, medium, large function settings
* Added permission to watch configmaps to function deployment
* Updated dependencies (Prometheus, Nginx)

### Changes in 0.1.14

* Added helm labels to function namespace
* Updated prometheus, nginx dependencies

### Changes in 0.1.13

* Added function namespace generation automatically
* Added update strategy for deployments
* Ingresses are optional, default true

### Changes in 0.1.12

*Version upgrade*

### Changes in 0.1.11

*Version upgrade*

### Changes in 0.1.10

*Version upgrade*
*Added API secret*

### Changes in 0.1.9
*Version fix*

### Changes in 0.1.8
*Version upgrade*

### Changes in 0.1.7
*Version upgrade*

### Changes in 0.1.6
*Flow filesystem is writable for git integration*

### Changes in 0.1.3

*creating of service accounts is optional*
*added `additional` for additional attribuites for db connections*
*make the cpu/mem limits for knative containers configurable*
*multiple replicas have now requiredDuringSchedulingIgnoredDuringExecution podAntiAffinity*

### Changes in 0.1.2

*Removed unnecessary environment variables in UI deployment*
*Fixed typo in opentelemetry config*

## Installing the Chart

To install the chart with the release name `direktiv`:

```console
$ kubectl create ns direktiv-services-direktiv
$ helm repo add direktiv https://chart.direktiv.io
$ helm install direktiv direktiv/direktiv
```

## Requirements

| Repository | Name | Version |
|------------|------|---------|
| https://kubernetes.github.io/ingress-nginx | ingress-nginx | 4.9.0 |
| https://prometheus-community.github.io/helm-charts | prometheus | 25.8.1 |

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| affinity | object | `{}` |  |
| apikey | string | `"none"` | enabled api key for API authentication with the `direktiv-token` header |
| database.additional | string | `""` | additional connection attributes, e.g. target_session_attrs |
| database.host | string | `"postgres-postgresql-ha-pgpool.postgres"` | database host |
| database.name | string | `"direktiv"` | database name, has to be created before installation |
| database.password | string | `"direktivdirektiv"` | database password |
| database.port | int | `5432` | database port |
| database.sslmode | string | `"require"` | sslmode for database |
| database.user | string | `"direktiv"` | database user |
| eventing | object | `{"enabled":true}` | knative eventing enabled, requires knative setup and configuration |
| flow.affinity | object | `{}` | affinity for flow pods |
| flow.containers.secrets.resources.limits.memory | string | `"512Mi"` |  |
| flow.containers.secrets.resources.requests.memory | string | `"128Mi"` |  |
| flow.debug | bool | `false` | Output debug-level logs. |
| flow.encryptionKey | string | `nil` | Set to define an encryption key to be used for secrets. If set to empty, one will be generated on install. |
| flow.extraContainers | list | `[]` | extra container in flow pod |
| flow.extraVariables | list | `[]` | extra environment variables in flow pod  |
| flow.extraVolumeMounts | string | `nil` | extra volume mounts in flow pod |
| flow.extraVolumes | string | `nil` | extra volumes in flow pod |
| flow.logging | string | `"json"` | Set to 'json' or 'console' to change log output format. |
| flow.replicas | int | `1` | number of flow replicas |
| fluentbit.extraConfig | string | `""` | postgres for direktiv services Append extra output to fluentbit configuration. There are two log types: application (system), functions (workflows) these can be matched to new outputs. |
| frontend | object | `{"additionalAnnotations":{},"additionalLabels":{},"additionalSecEnvs":{},"backend":{"skip-verify":false,"url":null},"certificate":null,"command":"","extraConfig":null,"extraVariables":[],"image":"direktiv/frontend","logging":{"debug":true,"json":true},"nginx":{"config":"","resolver":"kube-dns.kube-system.svc.cluster.local valid=10s;"},"replicas":1,"resources":{"limits":{"memory":"512Mi"},"requests":{"memory":"128Mi"}},"tag":""}` | Frontend configuration |
| frontend.additionalAnnotations | object | `{}` | Additional Annotations for frontend |
| frontend.additionalLabels | object | `{}` | Additional Labels for frontend |
| frontend.additionalSecEnvs | object | `{}` | Additional secret environment variables |
| frontend.backend | object | `{"skip-verify":false,"url":null}` | Backend configuration for the workflow engine |
| frontend.backend.skip-verify | bool | `false` | Skip verifing TLS certificate if TLS is configured |
| frontend.backend.url | string | `nil` | Defaults to engine in the same namespace |
| frontend.certificate | string | `nil` | certificate secret for frontend |
| frontend.logging | object | `{"debug":true,"json":true}` | Logging setting for the UI |
| frontend.logging.debug | bool | `true` | Enable/Disable debug mode |
| frontend.logging.json | bool | `true` | Logging in JSON or console format |
| functions.affinity | object | `{}` |  |
| functions.containers.functionscontroller.resources.limits.memory | string | `"1024Mi"` |  |
| functions.containers.functionscontroller.resources.requests.memory | string | `"128Mi"` |  |
| functions.extraContainers | list | `[]` | extra containers for tasks and knative pods |
| functions.extraContainersPod | list | `[]` | extra containers for function controller, e.g. database containers for google cloud or logging |
| functions.extraVolumes | list | `[]` | extra volumes for tasks and knative pods |
| functions.ingressClass | string | `"contour.ingress.networking.knative.dev"` |  |
| functions.limits | object | `{"cpu":{"large":1,"medium":"500m","small":"250m"},"disk":{"large":4096,"medium":1024,"small":256},"memory":{"large":2048,"medium":1024,"small":512}}` | knative service limits  |
| functions.namespace | string | `"direktiv-services-direktiv"` |  |
| functions.netShape | string | `nil` | Egress/Ingress network limit for functions if supported by network |
| functions.podCleaner | bool | `true` | Cleaning up tasks, Kubernetes < 1.20 does not clean finished tasks |
| functions.replicas | int | `1` | number of controller replicas |
| functions.runtime | string | `"default"` | runtime to use, e.g. gvisor on GCP |
| http_proxy | string | `""` | http proxy settings |
| https_proxy | string | `""` | https proxy settings |
| image | string | `"direktiv/direktiv"` | image for main direktiv binary |
| imagePullSecrets | list | `[]` | Container registry secrets. |
| ingress-nginx | object | `{"controller":{"admissionWebhooks":{"patch":{"podAnnotations":{"linkerd.io/inject":"disabled"}}},"config":{"proxy-buffer-size":"16k"},"podAnnotations":{"linkerd.io/inject":"disabled"},"replicaCount":1},"install":true}` | nginx ingress controller configuration |
| ingress.additionalAnnotations | object | `{}` | Additional Annotations |
| ingress.additionalLabels | object | `{}` | Additional Labels |
| ingress.certificate | string | `nil` | TLS secret |
| ingress.class | string | `"nginx"` | Ingress class |
| ingress.enabled | bool | `true` |  |
| ingress.host | string | `nil` | Host for external services, only required for TLS |
| no_proxy | string | `""` | no proxy proxy settings |
| nodeSelector | object | `{}` |  |
| opentelemetry.agentconfig | string | `"receivers:\n  otlp:\n    protocols:\n      grpc:\n      http:\nexporters:\n  otlp:\n    endpoint: \"tempo.grafana.svc:4317\" # otel receivers grpc port for expl. tempo \n    insecure: true\n    sending_queue:\n      num_consumers: 4\n      queue_size: 100\n    retry_on_failure:\n      enabled: true\n  logging:\n    loglevel: debug\nprocessors:\n  batch:\n  memory_limiter:\n    # Same as --mem-ballast-size-mib CLI argument\n    ballast_size_mib: 165\n    # 80% of maximum memory up to 2G\n    limit_mib: 400\n    # 25% of limit up to 2G\n    spike_limit_mib: 100\n    check_interval: 5s\nextensions:\n  zpages: {}\nservice:\n  extensions: [zpages]\n  pipelines:\n    traces:\n      receivers: [otlp]\n      processors: [memory_limiter, batch]\n      exporters: [logging, otlp]\n"` | config for sidecar agent |
| opentelemetry.enabled | bool | `false` | installs opentelemtry agent as sidecar in flow |
| opentelemetry.image | string | `"otel/opentelemetry-collector-dev:latest"` |  |
| prometheus.alertmanager.enabled | bool | `false` |  |
| prometheus.backendName | string | `"prom-backend-server"` |  |
| prometheus.global.evaluation_interval | string | `"1m"` |  |
| prometheus.global.scrape_interval | string | `"1m"` |  |
| prometheus.install | bool | `true` |  |
| prometheus.kube-state-metrics.enabled | bool | `false` |  |
| prometheus.prometheus-node-exporter.enabled | bool | `false` |  |
| prometheus.prometheus-pushgateway.enabled | bool | `false` |  |
| prometheus.server.persistentVolume.enabled | bool | `false` |  |
| prometheus.server.retention | string | `"96h"` |  |
| prometheus.serviceAccounts.server.create | bool | `true` |  |
| pullPolicy | string | `"Always"` | Container pull policy. |
| registry | string | `"docker.io"` | Container registry from which to pull direktiv. |
| requestTimeout | int | `7200` | max request timeouts in seconds. Used in Knative and the ingress controller if enabled. |
| serviceAccount | object | `{"annotations":{},"create":true,"name":""}` | service account for components. If preconfigured serviceaccounts are used the name ise the base  and two additional service accounts are needed, e.g. service account name is myaccount then another two  acounts are needed: myaccount-functions and myaccount-functions-pod |
| tag | string | `""` | image tag for main direktiv binary pod |
| tolerations | list | `[]` |  |

