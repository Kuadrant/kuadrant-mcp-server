# TokenRateLimitPolicy Reference

TokenRateLimitPolicy is a Kuadrant policy for rate limiting based on token consumption (e.g., AI/LLM tokens).

## Overview

TokenRateLimitPolicy enables rate limiting based on token usage from AI/LLM responses. It automatically tracks token consumption by monitoring the usage.total_tokens field in response bodies.

## Basic Example

```yaml
apiVersion: kuadrant.io/v1
kind: TokenRateLimitPolicy
metadata:
  name: ai-rate-limit
  namespace: default
spec:
  targetRef:
    group: gateway.networking.k8s.io
    kind: HTTPRoute
    name: ai-service
  limits:
    "users":
      rates:
      - limit: 10000
        window: 1h
      counters:
      - expression: "auth.identity.sub"
```

## Key Features

- **Token-based limiting**: Tracks actual token consumption from AI/LLM responses
- **User-specific limits**: Use counters to apply limits per user or API key
- **Flexible windows**: Support for various time windows (1m, 1h, 24h, etc.)
- **CEL expressions**: Use Common Expression Language for dynamic conditions

For complete documentation, see: https://docs.kuadrant.io/latest/kuadrant-operator/doc/reference/tokenratelimitpolicy/
