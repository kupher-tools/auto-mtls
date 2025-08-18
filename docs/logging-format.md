# Auto-mTLS Operator Logging Format

## Overview
The auto-mTLS operator uses structured logging with consistent formatting across all components. All logs include contextual information such as timestamps, log levels, component names, and relevant Kubernetes resource details.

## Log Format
```
2025-01-27T10:30:45Z    INFO    auto-mtls       Starting auto-mTLS reconciliation      {"service": "mtls-server", "namespace": "default"}
2025-01-27T10:30:45Z    INFO    auto-mtls       Certificate created successfully        {"certificate": "mtls-server-cert", "namespace": "default", "service": "mtls-server"}
2025-01-27T10:30:45Z    INFO    auto-mtls       mTLS certificates mounted successfully {"service": "mtls-server", "namespace": "default"}
```

## Log Levels
- **INFO**: Normal operational messages
- **ERROR**: Error conditions that need attention
- **V(1)**: Verbose/debug information (use `-v=1` flag)

## Component Names
- `setup`: Main application setup and initialization
- `cert-mgr`: Certificate manager infrastructure
- `ca-cert-mount`: CA certificate mounting operations
- `auto-mtls`: Main auto-mTLS service reconciliation

## Key Fields
All log entries include relevant contextual fields:
- `service`: Kubernetes service name
- `namespace`: Kubernetes namespace
- `certificate`: Certificate resource name
- `secret`: Secret resource name
- `deployment`: Deployment resource name
- `issuer`: Certificate issuer name
- `volume`: Volume name

## Usage Examples

### Enable verbose logging
```bash
kubectl logs -f deployment/auto-mtls-controller-manager -n auto-mtls-system -- --v=1
```

### Filter logs by component
```bash
kubectl logs deployment/auto-mtls-controller-manager -n auto-mtls-system | grep "cert-mgr"
```