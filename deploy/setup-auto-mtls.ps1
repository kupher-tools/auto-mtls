#!/usr/bin/env pwsh
#Requires -Version 5.1
$ErrorActionPreference = "Stop"

# ---- Config (overridable via env) ----
$CERT_MANAGER_VERSION   = $env:CERT_MANAGER_VERSION
if ([string]::IsNullOrWhiteSpace($CERT_MANAGER_VERSION)) { $CERT_MANAGER_VERSION = "v1.18.2" }

$AUTO_MTLS_MANIFEST_URL = $env:AUTO_MTLS_MANIFEST_URL
if ([string]::IsNullOrWhiteSpace($AUTO_MTLS_MANIFEST_URL)) { $AUTO_MTLS_MANIFEST_URL = "https://raw.githubusercontent.com/kupher-tools/auto-mtls/refs/heads/main/deploy/auto-mtls-deploy.yaml" }

function Need($cmd) {
  if (-not (Get-Command $cmd -ErrorAction SilentlyContinue)) {
    Write-Error "Missing dependency: '$cmd' not found on PATH"
  }
}

Write-Host "[*] Preflight checks"
Need "kubectl"
Need "helm"

Write-Host "[*] Install/upgrade cert-manager ($CERT_MANAGER_VERSION)"
helm upgrade --install cert-manager oci://quay.io/jetstack/charts/cert-manager `
  --version $CERT_MANAGER_VERSION `
  --namespace cert-manager `
  --create-namespace `
  --set crds.enabled=true | Out-Host

Write-Host "[*] Waiting for cert-manager components..."
kubectl -n cert-manager rollout status deploy/cert-manager --timeout=180s | Out-Host
kubectl -n cert-manager rollout status deploy/cert-manager-webhook --timeout=180s | Out-Host
kubectl -n cert-manager rollout status deploy/cert-manager-cainjector --timeout=180s | Out-Host
Write-Host "[+] cert-manager ready"

Write-Host "[*] Deploy Auto-mTLS Operator"
kubectl apply -f $AUTO_MTLS_MANIFEST_URL | Out-Host

Write-Host "[*] Waiting for auto-mtls operator..."
kubectl -n auto-mtls rollout status deploy/auto-mtls-operator --timeout=180s | Out-Host
Write-Host "[✓] Auto-mTLS Operator ready"



# .\deploy\setup-auto-mtls.ps1
# or with overrides:
# $env:CERT_MANAGER_VERSION = "v1.18.2"
# $env:AUTO_MTLS_MANIFEST_URL = "https://raw.githubusercontent.com/kupher-tools/auto-mtls/refs/heads/main/deploy/auto-mtls-deploy.yaml"
# .\deploy\setup-auto-mtls.ps1
