#!/usr/bin/env sh
set -eu

# ---- Config (overridable) ----
: "${CERT_MANAGER_VERSION:=v1.18.2}"
: "${AUTO_MTLS_MANIFEST_URL:=https://raw.githubusercontent.com/kupher-tools/auto-mtls/refs/heads/main/deploy/auto-mtls-deploy.yaml}"

need() { command -v "$1" >/dev/null 2>&1 || { echo "ERROR: missing '$1' on PATH"; exit 1; }; }

echo "[*] Preflight checks"
need kubectl
need helm

echo "[*] Install/upgrade cert-manager ($CERT_MANAGER_VERSION)"
helm upgrade --install cert-manager oci://quay.io/jetstack/charts/cert-manager \
  --version "$CERT_MANAGER_VERSION" \
  --namespace cert-manager \
  --create-namespace \
  --set crds.enabled=true

echo "[*] Waiting for cert-manager components..."
kubectl -n cert-manager rollout status deploy/cert-manager --timeout=180s
kubectl -n cert-manager rollout status deploy/cert-manager-webhook --timeout=180s
kubectl -n cert-manager rollout status deploy/cert-manager-cainjector --timeout=180s
echo "[+] cert-manager ready"

echo "[*] Deploy Auto-mTLS Operator"
kubectl apply -f "$AUTO_MTLS_MANIFEST_URL"

echo "[*] Waiting for auto-mtls operator..."
kubectl -n auto-mtls rollout status deploy/auto-mtls-operator --timeout=180s
echo "[✓] Auto-mTLS Operator ready"

# Execution

# chmod +x deploy/setup-auto-mtls.sh
# deploy/setup-auto-mtls.sh
# or with overrides:
# CERT_MANAGER_VERSION=v1.18.2 \
# AUTO_MTLS_MANIFEST_URL="https://raw.githubusercontent.com/kupher-tools/auto-mtls/refs/heads/main/deploy/auto-mtls-deploy.yaml" \
# deploy/setup-auto-mtls.sh
