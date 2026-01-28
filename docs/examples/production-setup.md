# Example: Production API Setup

Complete production setup with TLS, DNS, advanced rate limiting, and JWT authentication.

## 1. Production Gateway with TLS
```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: production-api
  namespace: gateway-system
  labels:
    kuadrant.io/gateway: "true"
spec:
  gatewayClassName: istio
  listeners:
    - name: https
      protocol: HTTPS
      port: 443
      hostname: "*.api.example.com"
      allowedRoutes:
        namespaces:
          from: All
      tls:
        mode: Terminate
        certificateRefs:
          - name: production-api-https
            kind: Secret
    - name: http
      protocol: HTTP
      port: 80
      hostname: "*.api.example.com"
```

## 2. TLS Certificate Automation
```yaml
apiVersion: kuadrant.io/v1
kind: TLSPolicy
metadata:
  name: production-tls
  namespace: gateway-system
spec:
  targetRef:
    name: production-api
    group: gateway.networking.k8s.io
    kind: Gateway
  issuerRef:
    group: cert-manager.io
    kind: ClusterIssuer
    name: letsencrypt-prod
  commonName: "*.api.example.com"
  duration: 90d
  renewBefore: 30d
```

## 3. DNS Management
```yaml
apiVersion: kuadrant.io/v1
kind: DNSPolicy
metadata:
  name: production-dns
  namespace: gateway-system
spec:
  targetRef:
    name: production-api
    group: gateway.networking.k8s.io
    kind: Gateway
  providerRefs:
    - name: aws-route53-credentials
  loadBalancing:
    geo:
      zones:
        - id: us-east-1
          weight: 100
        - id: eu-west-1
          weight: 100
      defaultGeo: true
  healthCheck:
    path: /health
    port: 443
    protocol: HTTPS
    failureThreshold: 3
    interval: 30s
```

## 4. Multi-Service HTTPRoute
```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: HTTPRoute
metadata:
  name: api-routes
  namespace: production
spec:
  parentRefs:
  - name: production-api
    namespace: gateway-system
  hostnames:
  - api.example.com
  - v2.api.example.com
  rules:
  - name: v2-api
    matches:
    - headers:
      - name: X-API-Version
        value: v2
    - hostname: v2.api.example.com
    backendRefs:
    - name: api-v2
      port: 8080
      weight: 100
  - name: v1-api
    matches:
    - path:
        type: PathPrefix
        value: /v1
    backendRefs:
    - name: api-v1
      port: 8080
      weight: 90
    - name: api-v1-canary
      port: 8080
      weight: 10
```

## 5. Advanced Rate Limiting
```yaml
apiVersion: kuadrant.io/v1
kind: RateLimitPolicy
metadata:
  name: production-rate-limits
  namespace: production
spec:
  targetRef:
    group: gateway.networking.k8s.io
    kind: HTTPRoute
    name: api-routes
  limits:
    "per-user-standard":
      rates:
        - limit: 100
          window: 1m
        - limit: 1000
          window: 1h
      counters:
        - expression: auth.identity.sub
      when:
        - predicate: 'auth.identity.tier == "standard"'

    "per-user-premium":
      rates:
        - limit: 1000
          window: 1m
        - limit: 100000
          window: 1h
      counters:
        - expression: auth.identity.sub
      when:
        - predicate: 'auth.identity.tier == "premium"'

    "expensive-endpoints":
      rates:
        - limit: 10
          window: 1m
      counters:
        - expression: auth.identity.sub
      when:
        - predicate: 'request.path in ["/v1/reports/generate", "/v1/export/large"]'

    "global-ddos-protection":
      rates:
        - limit: 10000
          window: 10s
```

## 6. JWT Authentication with OIDC
```yaml
apiVersion: kuadrant.io/v1
kind: AuthPolicy
metadata:
  name: production-auth
  namespace: production
spec:
  targetRef:
    group: gateway.networking.k8s.io
    kind: HTTPRoute
    name: api-routes
  rules:
    authentication:
      "jwt-auth":
        jwt:
          issuerUrl: https://auth.example.com/realms/production
          audiences:
            - api.example.com
        defaults:
          tier: standard
        overrides:
          tier:
            valueFrom:
              authJSON: auth.identity.realm_access.roles

    metadata:
      "user-profile":
        userInfo:
          identitySource: jwt

    authorization:
      "tier-check":
        opa:
          rego: |
            default allow = false

            # Public endpoints
            allow {
              input.request.path in ["/v1/public", "/health", "/metrics"]
            }

            # Authenticated users
            allow {
              input.auth.identity.sub != null
            }

            # Premium features
            allow {
              input.request.path.startsWith("/v1/premium")
              input.auth.identity.tier == "premium"
            }

    response:
      success:
        headers:
          "X-User-ID":
            valueFrom:
              authJSON: auth.identity.sub
          "X-Rate-Limit-Tier":
            valueFrom:
              authJSON: auth.identity.tier

      unauthorized:
        headers:
          "WWW-Authenticate":
            value: 'Bearer realm="production"'
          "Content-Type":
            value: "application/json"
        body:
          value: |
            {
              "error": "unauthorized",
              "message": "Invalid or missing authentication token"
            }
```

## Monitoring & Testing

### Health Check Endpoint
```bash
curl https://api.example.com/health
```

### Test Rate Limiting
```bash
# Get JWT token
TOKEN=$(curl -s -X POST https://auth.example.com/realms/production/protocol/openid-connect/token \
  -d "client_id=api-client" \
  -d "client_secret=$CLIENT_SECRET" \
  -d "grant_type=client_credentials" \
  | jq -r .access_token)

# Test API call
curl -H "Authorization: Bearer $TOKEN" \
     https://api.example.com/v1/data
```
