#!/bin/bash
set -e

echo "[+] Deploying sample app (httpbin + sleep)..."

kubectl apply -f https://raw.githubusercontent.com/istio/istio/release-1.22/samples/httpbin/httpbin.yaml
kubectl apply -f https://raw.githubusercontent.com/istio/istio/release-1.22/samples/sleep/sleep.yaml

echo "[+] Waiting for sample pods to be ready..."
kubectl wait --for=condition=Ready pods --all -n default --timeout=180s

echo "[+] Sample app deployed successfully!"
