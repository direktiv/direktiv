**_Ubuntu Development Installation_**

**Install k3s**

```
curl -sfL https://get.k3s.io | sh -
```

**Change k3s service**

Change  following line in */etc/systemd/system/k3s.service*

```
ExecStart=/usr/local/bin/k3s server --disable traefik --write-kubeconfig-mode=644
```

**Change ~/.bashrc**

```
alias kc="kubectl"
source <(kubectl completion bash)
complete -F __start_kubectl kc
export KUBECONFIG=/etc/rancher/k3s/k3s.yaml
```


**Install knative**
```
scripts/knative/install-knative.sh
```

**Install local registry**

```
docker run -d -p 5000:5000 --restart=always --name registry registry:2
```

**Enable pulling from insecure registry (k3s dev):**

Add the following to the specified files.

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

Run following to enable settings:

```
sudo systemctl daemon-reload && sudo service k3s restart & sudo service docker restart
```


**Disable tag-resolving for knative**

```
kubectl apply -f scripts/config-deployment.yaml
```
