#!/usr/bin/env bash
dir="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
kubectl apply -f $dir/install-monitoring.yaml
helm repo add grafana https://grafana.github.io/helm-charts
helm repo add fluent https://fluent.github.io/helm-charts
helm repo update
helm upgrade --install tempo grafana/tempo
helm upgrade --install fluent-bit fluent/fluent-bit --values $dir/fluentbit.values.yaml

echo "opentelemetry:
  # -- opentelemetry address where Direktiv is sending data to
  address: "localhost:4317"
  # -- installs opentelemtry agent as sidecar in flow
  enabled: true
  # -- config for sidecar agent
  agentconfig: |
    receivers:
      otlp:
        protocols:
          grpc:
          http:
    exporters:
      otlp:
        endpoint: "tempo:4317"
        insecure: true
        sending_queue:
          num_consumers: 4
          queue_size: 100
        retry_on_failure:
          enabled: true
      logging:
        loglevel: debug
    processors:
      batch:
      memory_limiter:
        # Same as --mem-ballast-size-mib CLI argument
        ballast_size_mib: 165
        # 80% of maximum memory up to 2G
        limit_mib: 400
        # 25% of limit up to 2G
        spike_limit_mib: 100
        check_interval: 5s
    extensions:
      zpages: {}
    service:
      extensions: [zpages]
      pipelines:
        traces:
          receivers: [otlp]
          processors: [memory_limiter, batch]
          exporters: [logging, otlp]" >> $dir/dev.yaml
helm upgrade -f $dir/dev.yaml direktiv $dir/direktiv-charts/charts/direktiv/
kubectl rollout restart deployment direktiv-flow
