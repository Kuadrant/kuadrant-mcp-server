# PlanPolicy Extension

PlanPolicy provides plan-based rate limiting for tiered service offerings.

## API Version
```yaml
apiVersion: extensions.kuadrant.io/v1alpha1
kind: PlanPolicy
```

## Overview

PlanPolicy enables different rate limits based on user subscription plans (e.g., gold, silver, bronze). It evaluates CEL predicates against authenticated identity to determine the user's plan tier.

## How It Works

1. AuthPolicy handles authentication and stores identity in secrets
2. PlanPolicy evaluates predicates to determine user's plan
3. Rate limits are automatically applied based on the plan tier

## Basic Example

```yaml
apiVersion: extensions.kuadrant.io/v1alpha1
kind: PlanPolicy
metadata:
  name: api-plans
spec:
  targetRef:
    group: gateway.networking.k8s.io
    kind: HTTPRoute
    name: api-route
  plans:
    - tier: gold
      predicate: |
        has(auth.identity) &&
        auth.identity.metadata.annotations["secret.kuadrant.io/plan-id"] == "gold"
      limits:
        daily: 10000
        weekly: 50000
    - tier: silver
      predicate: |
        has(auth.identity) &&
        auth.identity.metadata.annotations["secret.kuadrant.io/plan-id"] == "silver"
      limits:
        daily: 1000
        weekly: 5000
    - tier: bronze
      predicate: |
        has(auth.identity) &&
        auth.identity.metadata.annotations["secret.kuadrant.io/plan-id"] == "bronze"
      limits:
        daily: 100
        weekly: 500
```

## Plan Configuration

### Predicates

CEL expressions that match users to plans:

```yaml
# Based on secret annotation
predicate: |
  has(auth.identity) &&
  auth.identity.metadata.annotations["secret.kuadrant.io/plan-id"] == "gold"

# Based on JWT claim
predicate: |
  has(auth.identity) && auth.identity.tier == "premium"

# Multiple conditions
predicate: |
  has(auth.identity) &&
  auth.identity.metadata.labels["org"] == "enterprise" &&
  auth.identity.metadata.annotations["verified"] == "true"
```

### Limits

Available limit periods:

```yaml
limits:
  daily: 1000      # per day
  weekly: 5000     # per week
  monthly: 20000   # per month
  yearly: 200000   # per year
  custom:          # custom windows
    - limit: 100
      window: "1h"
    - limit: 10
      window: "1m"
```

## Prerequisites

- Kuadrant operator installed
- Gateway API resources configured
- AuthPolicy configured for authentication

## Important Notes

- Plans are evaluated in order (first match wins)
- More specific plans should come before general ones
- Requires authentication via AuthPolicy
- PlanPolicy must be in the same namespace as target

For complete documentation, see: https://docs.kuadrant.io/latest/kuadrant-operator/doc/extensions/planpolicy/
