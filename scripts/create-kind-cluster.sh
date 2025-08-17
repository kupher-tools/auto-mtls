#!/bin/bash
set -e

CLUSTER_NAME="dx-cluster"

echo "[+] Creating kind cluster: $CLUSTER_NAME"
kind create cluster --name $CLUSTER_NAME --config - <<EOF
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
- role: worker
- role: worker
EOF

echo "[+] Kind cluster '$CLUSTER_NAME' created successfully!"
