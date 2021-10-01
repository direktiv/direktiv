# knative

knative for direktiv

![Version: 0.2.0](https://img.shields.io/badge/Version-0.2.0-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: 0.25.1](https://img.shields.io/badge/AppVersion-0.25.1-informational?style=flat-square)

## Additional Information

This chart installs Knative for Direktiv. It configures Knative with correct values in Direktiv's context and it adds
 support to provide proxy values for corporate proxies.

## Installing the Chart

To install the chart with the release name `knative`:

```console
$ helm repo add direktiv https://charts.direktiv.io
$ helm install knative direktiv/knative
```

## Requirements

| Repository | Name | Version |
|------------|------|---------|
| https://charts.konghq.com | kong-external(kong) | 2.3.0 |
| https://charts.konghq.com | kong-internal(kong) | 2.3.0 |

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| http_proxy | string | `""` | HTTP proxy information for knative |
| https_proxy | string | `""` | HTTPS proxy information for knative |
| kong-external | object | `{"env":{"plugins":"key-auth,request-transformer","prefix":"/kong_prefix/"}}` | Kong for Direktiv's UI / API. Based on Kong Helm chart. |
| kong-internal | object | `{"ingressController":{"ingressClass":"kong-internal"},"proxy":{"type":"ClusterIP"}}` | Kong for internal services / direktiv functions. Based on Kong Helm chart. |
| no_proxy | string | `"localhost,127.0.0.1,10.0.0.0/8,.svc,.cluster.local"` | No proxy information for knative |

