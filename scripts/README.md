**_Ubuntu Development Installation_**

**Install k3s**

```
curl -sfL https://get.k3s.io | sh -s - --disable traefik --write-kubeconfig-mode=644 --kube-apiserver-arg feature-gates=TTLAfterFinished=true
```

**Change ~/.bashrc for code completion**

```
alias kc="kubectl"
source <(kubectl completion bash)
complete -F __start_kubectl kc
export KUBECONFIG=/etc/rancher/k3s/k3s.yaml
```

**Install local registry**

```
docker run -d -p 5000:5000 --restart=always --name registry registry:2
```

**Install helm**

```
curl -fsSL -o get_helm.sh https://raw.githubusercontent.com/helm/helm/master/scripts/get-helm-3
chmod 700 get_helm.sh
./get_helm.sh
```

**Base Install**

Installs DB, Knative
```
scripts/resetAll.sh
```

Installs Direktiv
```
make cluster
```
