# Kuadrant Custom Resource

The Kuadrant CR is the main custom resource for configuring the Kuadrant operator.

## Overview

The Kuadrant custom resource configures the Kuadrant operator instance, including:
- Observability settings
- mTLS configuration between components
- Component deployment options

## Basic Example

```yaml
apiVersion: kuadrant.io/v1beta1
kind: Kuadrant
metadata:
  name: kuadrant
  namespace: kuadrant-system
spec:
  observability:
    enable: true
  mtls:
    enable: true
    limitador: true
    authorino: true
```

## Features

### Observability
Enables metrics and tracing for Kuadrant components:
- Prometheus metrics
- OpenTelemetry tracing
- Component health monitoring

### mTLS Configuration
Secures communication between:
- Gateway and Limitador (rate limiting)
- Gateway and Authorino (auth)
- Inter-component communication

For complete documentation, see: https://docs.kuadrant.io/latest/kuadrant-operator/doc/reference/kuadrant/
