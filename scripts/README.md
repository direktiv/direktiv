*Pull from insecure registry (k3s dev):*

/etc/rancher/k3s/registries.yaml:

```
"localhost:5000":
  endpoint:
    - "localhost:5000"
```

/etc/docker/daemon.json:

```
{
  "insecure-registries" : ["localhost:5000"]
}
```


https://github.com/knative/serving/issues/7881
https://stackoverflow.com/questions/63671125/how-to-collect-knative-service-logs
