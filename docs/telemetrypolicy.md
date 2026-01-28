# TelemetryPolicy Reference

TelemetryPolicy enables custom metrics labels for Gateway API resources using CEL expressions.

## API Version
```yaml
apiVersion: kuadrant.io/v1alpha1
kind: TelemetryPolicy
```

## Overview

TelemetryPolicy attaches to Gateway or HTTPRoute resources and adds custom labels to metrics. This is useful for tracking token consumption, user tiers, or other business metrics.

## Basic Example

```yaml
apiVersion: kuadrant.io/v1alpha1
kind: TelemetryPolicy
metadata:
  name: ai-metrics
  namespace: default
spec:
  targetRef:
    group: gateway.networking.k8s.io
    kind: HTTPRoute
    name: ai-service
  metrics:
    default:
      labels:
        user_id: auth.identity.sub
        plan_tier: auth.identity.metadata.annotations["plan-id"]
        model: request.headers["x-model-name"]
```

## Key Concepts

- **targetRef**: Reference to Gateway or HTTPRoute (same namespace required)
- **metrics.default.labels**: Map of label names to CEL expressions
- Labels are evaluated per-request and added to metrics

## Use Cases

### Token Metrics by User
```yaml
metrics:
  default:
    labels:
      user_id: auth.identity.sub
      org_id: auth.identity.org
```

### Model Usage Tracking
```yaml
metrics:
  default:
    labels:
      model_name: request.headers["x-model"]
      api_version: request.path.split("/")[2]
```

## Integration with TokenRateLimitPolicy

TelemetryPolicy works alongside TokenRateLimitPolicy to provide:
- Token consumption tracking via labels
- Per-user/per-org metrics aggregation
- Custom dashboards based on business attributes

For complete documentation, see: https://docs.kuadrant.io/latest/kuadrant-operator/doc/reference/telemetrypolicy/
