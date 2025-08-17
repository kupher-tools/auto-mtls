#!/bin/bash
set -e

echo "[+] Installing cert-manager..."

kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.15.1/cert-manager.yaml

echo "[+] Waiting for cert-manager pods to be ready..."
kubectl wait --namespace cert-manager \
  --for=condition=Ready pods --all \
  --timeout=180s

echo "[+] cert-manager installed successfully!"
