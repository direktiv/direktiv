**_Ubuntu Development Installation_**

# sudo sysctl fs.inotify.max_user_watches=524288
# sudo sysctl fs.inotify.max_user_instances=512



**Install kind and kubectl**

https://kind.sigs.k8s.io/docs/user/quick-start/

https://kubernetes.io/docs/tasks/tools/

***Linux***

echo fs.inotify.max_user_watches=524288 | sudo tee -a /etc/sysctl.conf
echo fs.inotify.max_user_instances=512 | sudo tee -a /etc/sysctl.conf

**Change ~/.bashrc for code completion**

```
alias kc="kubectl"
source <(kubectl completion bash)
complete -F __start_kubectl kc
```

**Install helm**

```
curl -fsSL -o get_helm.sh https://raw.githubusercontent.com/helm/helm/master/scripts/get-helm-3
chmod 700 get_helm.sh
./get_helm.sh
```

**Base Install**

The common make targets are:

*cluster-setup*: Sets up the full cluster including Knative, Direktiv and Database

*cluster-build*: Creates an empty cluster

*cluster-build*: Builds Direktiv and pushes it to the registry at localhost:5001/direktiv:dev

*cluster-direktiv*: Installs only Direktiv (for chart development)

*cluster-direktiv-delete*: Removes Direktiv from cluster

*cluster-direktiv-run*: Builds Direktiv and replaces the pod with a new version and tails the log

For cluster tests use 'KIND_CONFIG=kind-config-cluster.yaml make cluster-setup'