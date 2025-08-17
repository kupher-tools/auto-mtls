#!/bin/bash
set -e

echo "[+] Installing Istio (demo profile)..."

istioctl install --set profile=demo -y

echo "[+] Labeling default namespace for Istio injection..."
kubectl label namespace default istio-injection=enabled --overwrite

echo "[+] Istio installed successfully!"
