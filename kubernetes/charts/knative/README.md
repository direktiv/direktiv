# knative

knative for direktiv

![Version: 0.2.0](https://img.shields.io/badge/Version-0.2.0-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: 0.25.1](https://img.shields.io/badge/AppVersion-0.25.1-informational?style=flat-square)

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
$ helm install my-release foo-bar/knative
```

## Requirements

| Repository | Name | Version |
|------------|------|---------|
| https://charts.konghq.com | kong-external(kong) | 2.3.0 |
| https://charts.konghq.com | kong-internal(kong) | 2.3.0 |

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| http_proxy | string | `""` |  |
| https_proxy | string | `""` |  |
| kong-external.env.plugins | string | `"grpc-gateway,grpc-stream"` |  |
| kong-external.env.prefix | string | `"/kong_prefix/"` |  |
| kong-internal.ingressController.ingressClass | string | `"kong-internal"` |  |
| kong-internal.proxy.type | string | `"ClusterIP"` |  |
| no_proxy | string | `""` |  |

