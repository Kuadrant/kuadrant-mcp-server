# Example: Basic API Setup

This example shows a complete setup for a simple API with rate limiting and API key authentication.

## 1. Create the Gateway
```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: api-gateway
  namespace: gateway-system
  labels:
    kuadrant.io/gateway: "true"
spec:
  gatewayClassName: istio
  listeners:
    - name: api
      protocol: HTTP
      port: 80
      hostname: api.example.com
      allowedRoutes:
        namespaces:
          from: All
```

## 2. Create the HTTPRoute
```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: toystore-api
  namespace: toystore
spec:
  parentRefs:
  - name: api-gateway
    namespace: gateway-system
  hostnames:
  - api.example.com
  rules:
  - name: public-endpoints
    matches:
    - path:
        type: PathPrefix
        value: "/v1/products"
      method: GET
    backendRefs:
    - name: toystore-service
      port: 8080
  - name: admin-endpoints
    matches:
    - path:
        type: PathPrefix
        value: "/v1/admin"
    backendRefs:
    - name: toystore-service
      port: 8080
```

## 3. Add Rate Limiting
```yaml
apiVersion: kuadrant.io/v1
kind: RateLimitPolicy
metadata:
  name: toystore-rate-limits
  namespace: toystore
spec:
  targetRef:
    group: gateway.networking.k8s.io
    kind: HTTPRoute
    name: toystore-api
  limits:
    "public-api":
      rates:
        - limit: 100
          window: 1m
      when:
        - predicate: 'request.path.startsWith("/v1/products")'
    "admin-api":
      rates:
        - limit: 1000
          window: 1m
      when:
        - predicate: 'request.path.startsWith("/v1/admin")'
```

## 4. Add Authentication
```yaml
apiVersion: kuadrant.io/v1
kind: AuthPolicy
metadata:
  name: toystore-auth
  namespace: toystore
spec:
  targetRef:
    group: gateway.networking.k8s.io
    kind: HTTPRoute
    name: toystore-api
    sectionName: admin-endpoints
  rules:
    authentication:
      "api-key":
        apiKey:
          selector:
            matchLabels:
              app: toystore
          credentials:
            authorizationHeader:
              prefix: "Bearer"
```

## 5. Create API Keys
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: admin-user-key
  namespace: toystore
  labels:
    authorino.kuadrant.io/managed-by: authorino
    app: toystore
  annotations:
    kuadrant.io/userid: "admin"
    kuadrant.io/groups: "admin,users"
stringData:
  api_key: my-secret-admin-key
type: Opaque
---
apiVersion: v1
kind: Secret
metadata:
  name: regular-user-key
  namespace: toystore
  labels:
    authorino.kuadrant.io/managed-by: authorino
    app: toystore
  annotations:
    kuadrant.io/userid: "user1"
    kuadrant.io/groups: "users"
stringData:
  api_key: my-secret-user-key
type: Opaque
```

## Testing

### Public endpoint (no auth required):
```bash
curl -H "Host: api.example.com" http://gateway.example.com/v1/products
```

### Admin endpoint (auth required):
```bash
curl -H "Host: api.example.com" \
     -H "Authorization: Bearer my-secret-admin-key" \
     http://gateway.example.com/v1/admin/users
```
