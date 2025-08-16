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

## Installation

1. Install Cert-Manager v1.18.2 with below command:
 ```sh
   
  helm install \
  cert-manager oci://quay.io/jetstack/charts/cert-manager \
  --version v1.18.2 \
  --namespace cert-manager \
  --create-namespace \
  --set crds.enabled=true
```
2.  

### Prerequisites
- Cert-Manager v1.18.2


### To Deploy on the cluster
**Build and push your image to the location specified by `IMG`:**

```sh
make docker-build docker-push IMG=<some-registry>/auto-mtls:tag
```

**NOTE:** This image ought to be published in the personal registry you specified.
And it is required to have access to pull the image from the working environment.
Make sure you have the proper permission to the registry if the above commands don’t work.

**Install the CRDs into the cluster:**

```sh
make install
```

**Deploy the Manager to the cluster with the image specified by `IMG`:**

```sh
make deploy IMG=<some-registry>/auto-mtls:tag
```

> **NOTE**: If you encounter RBAC errors, you may need to grant yourself cluster-admin
privileges or be logged in as admin.

**Create instances of your solution**
You can apply the samples (examples) from the config/sample:

```sh
kubectl apply -k config/samples/
```

>**NOTE**: Ensure that the samples has default values to test it out.

### To Uninstall
**Delete the instances (CRs) from the cluster:**

```sh
kubectl delete -k config/samples/
```

**Delete the APIs(CRDs) from the cluster:**

```sh
make uninstall
```

**UnDeploy the controller from the cluster:**

```sh
make undeploy
```

## Project Distribution

Following the options to release and provide this solution to the users.

### By providing a bundle with all YAML files

1. Build the installer for the image built and published in the registry:

```sh
make build-installer IMG=<some-registry>/auto-mtls:tag
```

**NOTE:** The makefile target mentioned above generates an 'install.yaml'
file in the dist directory. This file contains all the resources built
with Kustomize, which are necessary to install this project without its
dependencies.

2. Using the installer

Users can just run 'kubectl apply -f <URL for YAML BUNDLE>' to install
the project, i.e.:

```sh
kubectl apply -f https://raw.githubusercontent.com/<org>/auto-mtls/<tag or branch>/dist/install.yaml
```

### By providing a Helm Chart

1. Build the chart using the optional helm plugin

```sh
operator-sdk edit --plugins=helm/v1-alpha
```

2. See that a chart was generated under 'dist/chart', and users
can obtain this solution from there.

**NOTE:** If you change the project, you need to update the Helm Chart
using the same command above to sync the latest changes. Furthermore,
if you create webhooks, you need to use the above command with
the '--force' flag and manually ensure that any custom configuration
previously added to 'dist/chart/values.yaml' or 'dist/chart/manager/manager.yaml'
is manually re-applied afterwards.

## Contributing
// TODO(user): Add detailed information on how you would like others to contribute to this project

**NOTE:** Run `make help` for more information on all potential `make` targets

More information can be found via the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html)

## License

Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

