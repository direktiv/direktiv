**_Ubuntu Development Installation_**

**Install k3s**

```
curl -sfL https://get.k3s.io | sh -s - --disable traefik --write-kubeconfig-mode=644
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

You can set up a registry with TLS to use Knatives tag resolution. This creates certificates based on the hostname and starts the registry with certificates. Images can be tagged and pushed with HOSTNAME:5443. 

```
scripts/registry/setup.sh
```

**Install helm**

```
curl -fsSL -o get_helm.sh https://raw.githubusercontent.com/helm/helm/master/scripts/get-helm-3
chmod 700 get_helm.sh
./get_helm.sh
```

**Base Install**

Installs DB, Knative, Direktiv
```
make -C ../ k3s-install
```
