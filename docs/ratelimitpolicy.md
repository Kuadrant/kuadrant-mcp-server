# RateLimitPolicy

RateLimitPolicy provides fine-grained rate limiting for APIs based on various attributes like user identity, request path, method, headers, and more.

## API Version
```yaml
apiVersion: kuadrant.io/v1
kind: RateLimitPolicy
```

## Core Concepts

- **Limits**: Named sets of rate limit rules
- **Rates**: Time-based request allowances (e.g., 100 requests per minute)
- **Counters**: Attributes that create separate limit buckets
- **When conditions**: CEL predicates for conditional activation

## Simple Rate Limiting

### Basic API Protection
```yaml
apiVersion: kuadrant.io/v1
kind: RateLimitPolicy
metadata:
  name: api-limit
  namespace: api
spec:
  targetRef:
    group: gateway.networking.k8s.io
    kind: HTTPRoute
    name: api-route
  limits:
    "global":
      rates:
        - limit: 1000
          window: 1m
```

### Per-Endpoint Limits
```yaml
spec:
  targetRef:
    group: gateway.networking.k8s.io
    kind: HTTPRoute
    name: api-route
    sectionName: expensive-endpoints  # Target specific rule
  limits:
    "expensive-ops":
      rates:
        - limit: 10
          window: 1m
      when:
        - predicate: 'request.path.startsWith("/api/v1/expensive")'
```

## Advanced Patterns

### Per-User Rate Limiting
```yaml
limits:
  "per-user":
    rates:
      - limit: 100
        window: 1m
    counters:
      - expression: auth.identity.userid  # Each user gets their own counter
```

### Multi-Tier Rate Limits
```yaml
limits:
  "burst-protection":
    rates:
      - limit: 50
        window: 10s  # Burst protection
      - limit: 200
        window: 1m   # Sustained rate
```

### Conditional Rate Limiting
```yaml
limits:
  "public-api":
    rates:
      - limit: 10
        window: 1m
    when:
      - predicate: '!auth.identity.anonymous'  # Only for authenticated users
  "private-api":
    rates:
      - limit: 1000
        window: 1m
    when:
      - predicate: 'auth.identity.groups.contains("premium")'
```

## Gateway-Level Policies

### Default Limits for All Routes
```yaml
spec:
  targetRef:
    group: gateway.networking.k8s.io
    kind: Gateway
    name: prod-gateway
  defaults:
    limits:
      "gateway-default":
        rates:
          - limit: 10000
            window: 1m
```

### Override Limits
```yaml
spec:
  targetRef:
    group: gateway.networking.k8s.io
    kind: Gateway
    name: prod-gateway
  overrides:  # Forces these limits, cannot be overridden by route policies
    limits:
      "ddos-protection":
        rates:
          - limit: 100000
            window: 1m
```

## Well-Known Attributes

Common attributes for counters and conditions:
- `request.path` - Request path
- `request.method` - HTTP method
- `request.headers.x-foo` - Request headers
- `auth.identity.userid` - User identifier
- `auth.identity.groups` - User groups
- `source.address` - Client IP address
