# Auto-mTLS Operator
 #### Zero-Touch Mutual TLS for Kubernetes Workloads
Auto-mTLS is a lightweight Kubernetes operator that enables mutual TLS (mTLS) between services automatically, without requiring a full service mesh.

## Description
`Auto-mTLS Operator` automatically manages mutual TLS (mTLS) for Kubernetes Services annotated with `auto-mtls.kupher.io/enabled=true`. It creates certificates, secrets, and cleans up when services are deleted.

This operator does not use any CRDs — it works entirely with built-in Kubernetes resources (Services, Secrets, Certificates).

## 🔑 Key Features:
- **Zero-Touch Setup** – No manual cert management; certificates are issued, rotated, and revoked automatically.

- **Works on top of cert-manager** – Leverages cert-manager to handle PKI operations securely.

- **mTLS Only, No Overhead** – Focused purely on mutual TLS; **no heavy service mesh components**.

- **Lightweight & Cloud-Native** – Minimal resource footprint, works with any Kubernetes cluster.

## 📌 When to Use auto-mTLS

- When security (mTLS) is needed but a full service mesh is overkill.

- For DevSecOps & Platform Engineering teams who want secure service-to-service communication without complexity.

- In production workloads where simplicity, performance, and security matter.

## Install Auto-mTLS Operator

1. Deploy Cert-Manager v1.18.2 with below command:
 ```sh
   
  helm install \
  cert-manager oci://quay.io/jetstack/charts/cert-manager \
  --version v1.18.2 \
  --namespace cert-manager \
  --create-namespace \
  --set crds.enabled=true
```

2.  Deploy auto-mtls operator using below command:
```sh

kubectl apply -f https://raw.githubusercontent.com/kupher-tools/auto-mtls/refs/heads/main/deploy/auto-mtls-deploy.yaml
```
## 🔐 Setup Zero-Touch mTLS with Auto-MTLS Operator

The **Auto-MTLS Operator** automatically provisions TLS certificates, CA bundles, and mounts them into your workloads — no manual secret management required.
It leverages cert-manager under the hood, but keeps things lightweight compared to a full service mesh.

### 1. Deploy the Server
Deploy a server Service + Deployment.

Notice that **no TLS certificates are mounted manually** — the operator detects the annotation `auto-mtls.kupher.io/enabled=true` and handles certificate + CA injection automatically.

```sh
apiVersion: v1
kind: Service
metadata:
  name: mtls-server
  annotations:
    auto-mtls.kupher.io/enabled: "true"
spec:
  selector:
    app: mtls-server
  ports:
    - protocol: TCP
      port: 8443
      targetPort: 8443
  type: ClusterIP
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: mtls-server
  labels:
    app: mtls-server
spec:
  replicas: 1
  selector:
    matchLabels:
      app: mtls-server
  template:
    metadata:
      labels:
        app: mtls-server
    spec:
      containers:
        - name: mtls-server
          image: kupher/mtls-server-example:v0.0.3
          ports:
            - containerPort: 8443
```
Or apply directly:

```sh
kubectl apply -f https://raw.githubusercontent.com/kupher-tools/auto-mtls/refs/heads/main/examples/mtls-server/deploy/mtls-server.yaml

```
➡️ Once created, you will see below resources created for Server Application:

- A Certificate resource (`kubectl get certificate mtls-server-cert`)

- A TLS Secret with tls.crt + tls.key (`kubectl get secret mtls-server-cert-tls`)

- The CA cert Secret (`kubectl get secret auto-mtls-ca-cert`)

- Secrets automatically mounted into Server Pod.(`kubectl describe pod <pod-name>`)

2. Deploy the Client

Similarly, deploy a client workload. Again, **no manual TLS secrets — the operator injects them.**

```sh
apiVersion: v1
kind: Service
metadata:
  name: mtls-client
  namespace: default
  annotations:
    auto-mtls.kupher.io/enabled: "true"   # Example annotation to trigger operator
spec:
  selector:
    app: mtls-client
  ports:
    - protocol: TCP
      port: 8443
      targetPort: 8443
  type: ClusterIP
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: mtls-client
  labels:
    app: mtls-client
spec:
  replicas: 1
  selector:
    matchLabels:
      app: mtls-client
  template:
    metadata:
      labels:
        app: mtls-client
    spec:
      containers:
        - name: mtls-client
          image: kupher/mtls-client-example:v0.0.1
          env:
            - name: MTLS_SERVER_HOST
              value: "mtls-server"  # Service name of the mTLS server
          ports:
            - containerPort: 8443
```
Or apply directly:

```sh
kubectl apply -f https://raw.githubusercontent.com/kupher-tools/auto-mtls/refs/heads/main/examples/mtls-client/deploy/mtls-client.yaml

```

➡️ Once created, you will see below resources created for Client Application:

- A Certificate resource (`kubectl get certificate mtls-client-cert`)

- A TLS Secret with tls.crt + tls.key (`kubectl get secret mtls-client-cert-tls`)

- The CA cert Secret (`kubectl get secret auto-mtls-ca-cert`)

- Secrets automatically mounted into the Client Pod.(`kubectl describe pod <pod-name>`)

3. Verify mTLS

When both Pods are running:

- The Server only accepts connections authenticated with client certificates

- The Client uses the mounted TLS/CA bundle to authenticate itself

- All traffic between them is mutually authenticated (mTLS)

⚡ That’s it! You now have **Zero-Touch mTLS** — no need to manually create, distribute, or rotate TLS certs.


### Un-Install Auto-mTLS Operator
**Delete the Auto-mTLS Operator from the cluster:**

```sh
kubectl delete -f https://raw.githubusercontent.com/kupher-tools/auto-mtls/refs/heads/main/deploy/auto-mtls-deploy.yaml
```

**Delete the Cert-Manager from the cluster:**

```sh
helm uninstall cert-manager -n cert-manager
```



## Contribution

Contributions are welcome! 🎉  

If you'd like to help improve this project, please check out our [Contributing Guide](CONTRIBUTING.md) for details on how to get started. 

