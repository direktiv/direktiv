**Create Certificate for Direktiv**

```
sudo apt install golang-cfssl
```

```
cat <<EOF | cfssl genkey - | cfssljson -bare server
{
  "hosts": [
    "direktiv-flow",
    "direktiv-flow.default",
    "direktiv-flow.default.svc.cluster.local"
  ],
  "CN": "system:node:direktiv-flow.default.svc.cluster.local",
  "key": {
    "algo": "ecdsa",
    "size": 256
  },
  "names": [
    {
      "O": "system:nodes"
    }
  ]
}
EOF
```

```
cat <<EOF | kubectl apply -f -
apiVersion: certificates.k8s.io/v1
kind: CertificateSigningRequest
metadata:
  name: svc.default
spec:
  request: $(cat server.csr | base64 | tr -d '\n')
  signerName: kubernetes.io/kubelet-serving
  usages:
  - digital signature
  - key encipherment
  - server auth
EOF
```

```
kubectl certificate approve svc.default
```

```
kubectl get csr svc.default -o jsonpath='{.status.certificate}' \
    | base64 --decode > server.crt
```
