package main

import (
	"context"
	"log"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Resource handlers with enhanced documentation from official Kuadrant docs

func gatewayAPIResourceHandler(ctx context.Context, ss *mcp.ServerSession, params *mcp.ReadResourceParams) (*mcp.ReadResourceResult, error) {
	log.Printf("[KUADRANT MCP] Resource requested: %s", params.URI)
	content := `# Gateway API and Kuadrant

The Gateway API is a Kubernetes API for managing ingress traffic. Kuadrant extends the Gateway API with additional policies for:

- **DNS management** (DNSPolicy) - Automatic DNS record management
- **TLS certificate management** (TLSPolicy) - Automated cert provisioning
- **Rate limiting** (RateLimitPolicy) - Request rate control
- **Authentication and authorization** (AuthPolicy) - Identity and access management

## Enabling Kuadrant on a Gateway

Add the label to your Gateway to enable Kuadrant features:

` + "```yaml" + `
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: external
  namespace: api-gateway
  labels:
    kuadrant.io/gateway: "true"  # Enable Kuadrant
spec:
  gatewayClassName: istio
  listeners:
    - name: http
      protocol: HTTP
      port: 80
      allowedRoutes:
        namespaces:
          from: All
` + "```" + `

## Policy Attachment

Policies attach to Gateway API resources:
- **Gateway**: Affects all routes through the gateway
- **HTTPRoute**: Affects specific routes or rules (use sectionName)

## Common Gateway Patterns

### Gateway with Hostname
` + "```yaml" + `
listeners:
  - name: api
    hostname: "*.example.com"
    port: 443
    protocol: HTTPS
    tls:
      mode: Terminate
      certificateRefs:
        - name: example-com-tls
          kind: Secret
` + "```" + `

### Multi-Protocol Gateway
` + "```yaml" + `
listeners:
  - name: http
    protocol: HTTP
    port: 80
  - name: https
    protocol: HTTPS
    port: 443
    tls:
      mode: Terminate
` + "```" + `
`
	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{
			{
				URI:      params.URI,
				MIMEType: "text/plain",
				Text:     content,
			},
		},
	}, nil
}

func dnsPolicyResourceHandler(ctx context.Context, ss *mcp.ServerSession, params *mcp.ReadResourceParams) (*mcp.ReadResourceResult, error) {
	log.Printf("[KUADRANT MCP] Resource requested: %s", params.URI)
	content := `# DNSPolicy

DNSPolicy enables automatic DNS management for Gateway listeners. It creates and manages DNS records in supported providers (AWS Route53, Google Cloud DNS, Azure DNS).

## API Version
` + "```yaml" + `
apiVersion: kuadrant.io/v1
kind: DNSPolicy
` + "```" + `

## Key Features

- **Automatic DNS record creation** based on Gateway listeners
- **Multi-cluster load balancing** with geo and weighted routing
- **Health checking** for endpoint availability
- **Provider integration** via credential secrets

## Basic Example
` + "```yaml" + `
apiVersion: kuadrant.io/v1
kind: DNSPolicy
metadata:
  name: prod-dns
  namespace: gateway-system
spec:
  targetRef:
    name: prod-gateway
    group: gateway.networking.k8s.io
    kind: Gateway
  providerRefs:  
    - name: aws-credentials
` + "```" + `

## Load Balancing Configuration

### Geographic Load Balancing
` + "```yaml" + `
spec:
  loadBalancing:
    geo:
      zones:
        - id: us-east-1
          weight: 100
        - id: eu-west-1
          weight: 100
      defaultGeo: true  # Catch-all for unmatched regions
` + "```" + `

### Weighted Load Balancing
` + "```yaml" + `
spec:
  loadBalancing:
    weighted:
      weight: 120  # Default weight for this cluster
` + "```" + `

## Health Checks
` + "```yaml" + `
spec:
  healthCheck:
    path: /health
    port: 80
    protocol: HTTP
    failureThreshold: 3
    interval: 5min
    additionalHeaders:
      X-Health-Check: "dns-operator"
` + "```" + `

## Provider Credentials

### AWS Route53
` + "```bash" + `
kubectl create secret generic aws-credentials \
  --type=kuadrant.io/aws \
  --from-literal=AWS_ACCESS_KEY_ID=$AWS_ACCESS_KEY_ID \
  --from-literal=AWS_SECRET_ACCESS_KEY=$AWS_SECRET_ACCESS_KEY
` + "```" + `

### Google Cloud DNS
` + "```bash" + `
kubectl create secret generic gcp-credentials \
  --type=kuadrant.io/gcp \
  --from-file=GOOGLE=$HOME/.config/gcloud/application_default_credentials.json
` + "```" + `

## Important Notes

- Only one provider reference is currently supported
- Requires appropriate DNS zones to be pre-configured in the provider
- Changing from non-loadbalanced to loadbalanced requires policy recreation
- Health checks run from the DNS operator pod
`
	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{
			{
				URI:      params.URI,
				MIMEType: "text/plain",
				Text:     content,
			},
		},
	}, nil
}

func rateLimitPolicyResourceHandler(ctx context.Context, ss *mcp.ServerSession, params *mcp.ReadResourceParams) (*mcp.ReadResourceResult, error) {
	log.Printf("[KUADRANT MCP] Resource requested: %s", params.URI)
	content := `# RateLimitPolicy

RateLimitPolicy provides fine-grained rate limiting for APIs based on various attributes like user identity, request path, method, headers, and more.

## API Version
` + "```yaml" + `
apiVersion: kuadrant.io/v1
kind: RateLimitPolicy
` + "```" + `

## Core Concepts

- **Limits**: Named sets of rate limit rules
- **Rates**: Time-based request allowances (e.g., 100 requests per minute)
- **Counters**: Attributes that create separate limit buckets
- **When conditions**: CEL predicates for conditional activation

## Simple Rate Limiting

### Basic API Protection
` + "```yaml" + `
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
` + "```" + `

### Per-Endpoint Limits
` + "```yaml" + `
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
` + "```" + `

## Advanced Patterns

### Per-User Rate Limiting
` + "```yaml" + `
limits:
  "per-user":
    rates:
      - limit: 100
        window: 1m
    counters:
      - auth.identity.userid  # Each user gets their own counter
` + "```" + `

### Multi-Tier Rate Limits
` + "```yaml" + `
limits:
  "burst-protection":
    rates:
      - limit: 50
        window: 10s  # Burst protection
      - limit: 200
        window: 1m   # Sustained rate
` + "```" + `

### Conditional Rate Limiting
` + "```yaml" + `
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
` + "```" + `

## Gateway-Level Policies

### Default Limits for All Routes
` + "```yaml" + `
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
` + "```" + `

### Override Limits
` + "```yaml" + `
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
` + "```" + `

## Well-Known Attributes

Common attributes for counters and conditions:
- ` + "`request.path`" + ` - Request path
- ` + "`request.method`" + ` - HTTP method
- ` + "`request.headers.x-foo`" + ` - Request headers
- ` + "`auth.identity.userid`" + ` - User identifier
- ` + "`auth.identity.groups`" + ` - User groups
- ` + "`source.address`" + ` - Client IP address
`
	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{
			{
				URI:      params.URI,
				MIMEType: "text/plain",
				Text:     content,
			},
		},
	}, nil
}

func authPolicyResourceHandler(ctx context.Context, ss *mcp.ServerSession, params *mcp.ReadResourceParams) (*mcp.ReadResourceResult, error) {
	log.Printf("[KUADRANT MCP] Resource requested: %s", params.URI)
	content := `# AuthPolicy

AuthPolicy provides authentication and authorization for APIs using various identity sources and access control methods.

## API Version
` + "```yaml" + `
apiVersion: kuadrant.io/v1
kind: AuthPolicy
` + "```" + `

## Authentication Methods

### API Key Authentication
` + "```yaml" + `
apiVersion: kuadrant.io/v1
kind: AuthPolicy
metadata:
  name: api-key-auth
spec:
  targetRef:
    group: gateway.networking.k8s.io
    kind: HTTPRoute
    name: api-route
  rules:
    authentication:
      "api-key":
        apiKey:
          selector:
            matchLabels:
              app: myapp
          credentials:
            authorizationHeader:
              prefix: "Bearer"
` + "```" + `

API Key Secret:
` + "```yaml" + `
apiVersion: v1
kind: Secret
metadata:
  name: user-key-1
  labels:
    authorino.kuadrant.io/managed-by: authorino
    app: myapp
  annotations:
    kuadrant.io/userid: "user1"
    kuadrant.io/groups: "users,beta"
stringData:
  api_key: secret-key-value
type: Opaque
` + "```" + `

### JWT Authentication
` + "```yaml" + `
rules:
  authentication:
    "jwt":
      jwt:
        issuerUrl: https://auth.example.com/realms/api
        audiences:
          - api.example.com
        credentials:
          authorizationHeader:
            prefix: "Bearer"
` + "```" + `

### Kubernetes Token Review
` + "```yaml" + `
rules:
  authentication:
    "k8s-tokens":
      kubernetes:
        audiences:
          - https://kubernetes.default.svc.cluster.local
` + "```" + `

## Authorization Methods

### OPA/Rego Authorization
` + "```yaml" + `
rules:
  authorization:
    "admin-only":
      opa:
        rego: |
          groups := split(object.get(input.auth.identity.metadata.annotations, "kuadrant.io/groups", ""), ",")
          allow { groups[_] == "admin" }
` + "```" + `

### Kubernetes SubjectAccessReview
` + "```yaml" + `
rules:
  authorization:
    "k8s-rbac":
      kubernetes:
        user:
          valueFrom:
            authJSON: auth.identity.username
        resourceAttributes:
          namespace: { value: default }
          resource: { value: services }
          verb: { value: get }
` + "```" + `

### Pattern Matching
` + "```yaml" + `
patterns:
  "admin-path":
    - selector: request.path
      operator: startsWith
      value: "/admin"
      
rules:
  authorization:
    "admin-routes":
      patternMatching:
        patterns:
          - patternRef: admin-path
` + "```" + `

## Advanced Features

### Gateway-Level Deny All
` + "```yaml" + `
apiVersion: kuadrant.io/v1
kind: AuthPolicy
metadata:
  name: gateway-auth
  namespace: gateway-system
spec:
  targetRef:
    group: gateway.networking.k8s.io
    kind: Gateway
    name: prod-gateway
  defaults:
    strategy: atomic
    rules:
      authorization:
        "deny-all":
          opa:
            rego: "allow = false"
      response:
        unauthorized:
          headers:
            "content-type":
              value: application/json
          body:
            value: |
              {
                "error": "Forbidden",
                "message": "Access denied. Please contact your administrator."
              }
` + "```" + `

### Custom Response Headers
` + "```yaml" + `
rules:
  response:
    success:
      headers:
        "x-user-id":
          valueFrom:
            authJSON: auth.identity.userid
        "x-user-groups":
          valueFrom:
            authJSON: auth.identity.groups
` + "```" + `

### Metadata Enrichment
` + "```yaml" + `
rules:
  metadata:
    "user-info":
      userInfo:
        identitySource: jwt
    "external-data":
      http:
        url: https://api.example.com/user/{auth.identity.userid}
        method: GET
        headers:
          "Authorization":
            value: "Bearer {auth.identity.access_token}"
` + "```" + `

## Important Patterns

1. **Anonymous Access**: At least one auth rule must pass (unless using anonymous)
2. **All Authorization Must Pass**: All authorization rules must allow the request
3. **Defaults vs Overrides**: Use defaults for mergeable policies, overrides for strict enforcement
4. **Priority Groups**: Control authentication method precedence with priority values
`
	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{
			{
				URI:      params.URI,
				MIMEType: "text/plain",
				Text:     content,
			},
		},
	}, nil
}

func tlsPolicyResourceHandler(ctx context.Context, ss *mcp.ServerSession, params *mcp.ReadResourceParams) (*mcp.ReadResourceResult, error) {
	log.Printf("[KUADRANT MCP] Resource requested: %s", params.URI)
	content := `# TLSPolicy

TLSPolicy automates TLS certificate management for Gateways using cert-manager.

## API Version
` + "```yaml" + `
apiVersion: kuadrant.io/v1
kind: TLSPolicy
` + "```" + `

## Prerequisites

- cert-manager must be installed in the cluster
- An Issuer or ClusterIssuer must be configured
- Gateway must have HTTPS listeners

## Basic TLS Policy

### Self-Signed Certificate
` + "```yaml" + `
# First, create an issuer
apiVersion: cert-manager.io/v1
kind: Issuer
metadata:
  name: selfsigned-issuer
  namespace: gateway-system
spec:
  selfSigned: {}
---
# Then create the TLS policy
apiVersion: kuadrant.io/v1
kind: TLSPolicy
metadata:
  name: gateway-tls
  namespace: gateway-system
spec:
  targetRef:
    name: prod-gateway
    group: gateway.networking.k8s.io
    kind: Gateway
  issuerRef:
    group: cert-manager.io
    kind: Issuer
    name: selfsigned-issuer
` + "```" + `

### Let's Encrypt Production
` + "```yaml" + `
apiVersion: kuadrant.io/v1
kind: TLSPolicy
metadata:
  name: prod-tls
  namespace: gateway-system
spec:
  targetRef:
    name: prod-gateway
    group: gateway.networking.k8s.io
    kind: Gateway
  issuerRef:
    group: cert-manager.io
    kind: ClusterIssuer
    name: letsencrypt-prod
  commonName: "*.example.com"
  duration: 90d
  renewBefore: 30d
` + "```" + `

## Advanced Configuration

### Specific Listener Targeting
` + "```yaml" + `
spec:
  targetRef:
    name: multi-listener-gateway
    group: gateway.networking.k8s.io
    kind: Gateway
    sectionName: https-api  # Target specific listener
` + "```" + `

### Custom Certificate Parameters
` + "```yaml" + `
spec:
  duration: 2160h        # 90 days
  renewBefore: 720h      # 30 days before expiry
  revisionHistoryLimit: 3
  privateKey:
    algorithm: RSA
    size: 4096
    rotationPolicy: Always
  usages:
    - digital signature
    - key encipherment
    - server auth
    - client auth
` + "```" + `

## Common Issuer Configurations

### Let's Encrypt Staging (for testing)
` + "```yaml" + `
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt-staging
spec:
  acme:
    server: https://acme-staging-v02.api.letsencrypt.org/directory
    email: admin@example.com
    privateKeySecretRef:
      name: letsencrypt-staging
    solvers:
    - http01:
        ingress:
          ingressClassName: istio
` + "```" + `

### Let's Encrypt Production
` + "```yaml" + `
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt-prod
spec:
  acme:
    server: https://acme-v02.api.letsencrypt.org/directory
    email: admin@example.com
    privateKeySecretRef:
      name: letsencrypt-prod
    solvers:
    - dns01:
        route53:
          region: us-east-1
          accessKeyIDSecretRef:
            name: route53-credentials
            key: access-key-id
          secretAccessKeySecretRef:
            name: route53-credentials
            key: secret-access-key
` + "```" + `

## Certificate Usage

The generated certificate will be stored in a Secret with the name format:
- ` + "`{gateway-name}-{listener-name}`" + `

The Gateway's listener will automatically reference this certificate:
` + "```yaml" + `
listeners:
  - name: https
    protocol: HTTPS
    port: 443
    tls:
      mode: Terminate
      certificateRefs:
        - name: prod-gateway-https  # Auto-generated by TLSPolicy
          kind: Secret
` + "```" + `

## Important Notes

- TLSPolicy only targets Gateways (not HTTPRoutes)
- One TLSPolicy per Gateway listener
- Certificate renewal is automatic based on renewBefore setting
- DNS01 challenges are recommended for wildcard certificates
- HTTP01 challenges require the Gateway to be accessible on port 80
`
	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{
			{
				URI:      params.URI,
				MIMEType: "text/plain",
				Text:     content,
			},
		},
	}, nil
}

// New example resources
func exampleBasicSetupResourceHandler(ctx context.Context, ss *mcp.ServerSession, params *mcp.ReadResourceParams) (*mcp.ReadResourceResult, error) {
	log.Printf("[KUADRANT MCP] Resource requested: %s", params.URI)
	content := `# Example: Basic API Setup

This example shows a complete setup for a simple API with rate limiting and API key authentication.

## 1. Create the Gateway
` + "```yaml" + `
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
` + "```" + `

## 2. Create the HTTPRoute
` + "```yaml" + `
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
` + "```" + `

## 3. Add Rate Limiting
` + "```yaml" + `
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
` + "```" + `

## 4. Add Authentication
` + "```yaml" + `
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
` + "```" + `

## 5. Create API Keys
` + "```yaml" + `
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
` + "```" + `

## Testing

### Public endpoint (no auth required):
` + "```bash" + `
curl -H "Host: api.example.com" http://gateway.example.com/v1/products
` + "```" + `

### Admin endpoint (auth required):
` + "```bash" + `
curl -H "Host: api.example.com" \
     -H "Authorization: Bearer my-secret-admin-key" \
     http://gateway.example.com/v1/admin/users
` + "```" + `
`
	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{
			{
				URI:      params.URI,
				MIMEType: "text/plain",
				Text:     content,
			},
		},
	}, nil
}

func exampleProductionSetupResourceHandler(ctx context.Context, ss *mcp.ServerSession, params *mcp.ReadResourceParams) (*mcp.ReadResourceResult, error) {
	log.Printf("[KUADRANT MCP] Resource requested: %s", params.URI)
	content := `# Example: Production API Setup

Complete production setup with TLS, DNS, advanced rate limiting, and JWT authentication.

## 1. Production Gateway with TLS
` + "```yaml" + `
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
` + "```" + `

## 2. TLS Certificate Automation
` + "```yaml" + `
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
` + "```" + `

## 3. DNS Management
` + "```yaml" + `
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
` + "```" + `

## 4. Multi-Service HTTPRoute
` + "```yaml" + `
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
` + "```" + `

## 5. Advanced Rate Limiting
` + "```yaml" + `
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
        - auth.identity.sub
      when:
        - predicate: 'auth.identity.tier == "standard"'
    
    "per-user-premium":
      rates:
        - limit: 1000
          window: 1m
        - limit: 100000
          window: 1h
      counters:
        - auth.identity.sub
      when:
        - predicate: 'auth.identity.tier == "premium"'
    
    "expensive-endpoints":
      rates:
        - limit: 10
          window: 1m
      counters:
        - auth.identity.sub
      when:
        - predicate: 'request.path in ["/v1/reports/generate", "/v1/export/large"]'
    
    "global-ddos-protection":
      rates:
        - limit: 10000
          window: 10s
` + "```" + `

## 6. JWT Authentication with OIDC
` + "```yaml" + `
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
` + "```" + `

## 7. Gateway-Level Defaults
` + "```yaml" + `
apiVersion: kuadrant.io/v1
kind: AuthPolicy
metadata:
  name: gateway-defaults
  namespace: gateway-system
spec:
  targetRef:
    group: gateway.networking.k8s.io
    kind: Gateway
    name: production-api
  defaults:
    rules:
      authentication:
        "anonymous":
          anonymous: {}
          priority: 100
      response:
        unauthorized:
          code: 403
---
apiVersion: kuadrant.io/v1
kind: RateLimitPolicy
metadata:
  name: gateway-defaults
  namespace: gateway-system
spec:
  targetRef:
    group: gateway.networking.k8s.io
    kind: Gateway
    name: production-api
  defaults:
    limits:
      "basic-protection":
        rates:
          - limit: 1000
            window: 10s
` + "```" + `

## Monitoring & Testing

### Health Check Endpoint
` + "```bash" + `
curl https://api.example.com/health
` + "```" + `

### Test Rate Limiting
` + "```bash" + `
# Get JWT token
TOKEN=$(curl -s -X POST https://auth.example.com/realms/production/protocol/openid-connect/token \
  -d "client_id=api-client" \
  -d "client_secret=$CLIENT_SECRET" \
  -d "grant_type=client_credentials" \
  | jq -r .access_token)

# Test API call
curl -H "Authorization: Bearer $TOKEN" \
     https://api.example.com/v1/data
` + "```" + `
`
	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{
			{
				URI:      params.URI,
				MIMEType: "text/plain",
				Text:     content,
			},
		},
	}, nil
}

func troubleshootingResourceHandler(ctx context.Context, ss *mcp.ServerSession, params *mcp.ReadResourceParams) (*mcp.ReadResourceResult, error) {
	log.Printf("[KUADRANT MCP] Resource requested: %s", params.URI)
	content := `# Troubleshooting Guide

Common issues and solutions when working with Kuadrant policies.

## Policy Not Taking Effect

### Check Policy Status
` + "```bash" + `
# Check RateLimitPolicy status
kubectl get ratelimitpolicy -n <namespace> <policy-name> -o yaml

# Look for status conditions
status:
  conditions:
  - type: Available
    status: "True"    # Should be True
    reason: Available
` + "```" + `

### Common Causes:
1. **Wrong targetRef** - Ensure namespace and name are correct
2. **Policy conflicts** - Check for multiple policies targeting same resource
3. **Gateway not labeled** - Ensure ` + "`kuadrant.io/gateway: \"true\"`" + ` label

## Rate Limiting Issues

### Rate Limits Not Applied
` + "```bash" + `
# Check Limitador deployment
kubectl get deployment limitador-limitador -n kuadrant-system

# Check for rate limit configurations
kubectl get configmap limitador-config -n kuadrant-system -o yaml
` + "```" + `

### Debug Rate Limit Counters
` + "```bash" + `
# Port-forward to Limitador
kubectl port-forward -n kuadrant-system svc/limitador-limitador 8080:8080

# Check counters
curl http://localhost:8080/counters
` + "```" + `

## Authentication Failures

### 401 Unauthorized Errors
1. **Check Authorino logs**:
` + "```bash" + `
kubectl logs -n kuadrant-system deployment/authorino -f | grep <request-id>
` + "```" + `

2. **Verify API Key Secret**:
` + "```bash" + `
# Check secret has correct labels
kubectl get secret <api-key-secret> -o yaml

# Required labels:
# - authorino.kuadrant.io/managed-by: authorino
# - Your custom selector labels
` + "```" + `

3. **JWT Issues**:
- Verify issuer URL is accessible
- Check token expiration
- Validate audience claim matches policy

## DNS Not Resolving

### Check DNSRecord Status
` + "```bash" + `
kubectl get dnsrecord -A
kubectl describe dnsrecord <record-name> -n <namespace>
` + "```" + `

### Common Issues:
1. **Provider credentials** - Verify secret exists and has correct type
2. **DNS zone** - Ensure zone exists in provider
3. **Health checks failing** - Check endpoint accessibility

## TLS Certificate Issues

### Certificate Not Issued
` + "```bash" + `
# Check Certificate resource
kubectl get certificate -n <namespace>
kubectl describe certificate <cert-name> -n <namespace>

# Check cert-manager logs
kubectl logs -n cert-manager deployment/cert-manager
` + "```" + `

### Common Problems:
1. **ACME challenges failing** - Ensure HTTP01/DNS01 solver can complete
2. **Rate limits** - Let's Encrypt has rate limits; use staging for testing
3. **Issuer problems** - Verify ClusterIssuer/Issuer configuration

## Gateway Issues

### Gateway Not Ready
` + "```bash" + `
kubectl get gateway -n <namespace> <gateway-name> -o yaml

# Check status
status:
  conditions:
  - type: Programmed
    status: "True"  # Should be True
  - type: Accepted
    status: "True"  # Should be True
` + "```" + `

### No Address Assigned
- Check Gateway controller (Istio/Envoy Gateway) is running
- Verify LoadBalancer service is created
- Check cloud provider load balancer provisioning

## Debugging Tools

### Enable Debug Logging
` + "```yaml" + `
# For Authorino
kubectl set env deployment/authorino -n kuadrant-system LOG_LEVEL=debug

# For Limitador  
kubectl set env deployment/limitador-limitador -n kuadrant-system RUST_LOG=debug
` + "```" + `

### Useful Commands
` + "```bash" + `
# Get all Kuadrant resources
kubectl get gateway,httproute,ratelimitpolicy,authpolicy,dnspolicy,tlspolicy -A

# Check policy conditions
kubectl get ratelimitpolicy,authpolicy -A -o json | jq '.items[] | select(.status.conditions[]?.status != "True") | {name: .metadata.name, namespace: .metadata.namespace, conditions: .status.conditions}'

# Watch events
kubectl get events -A --field-selector reason=PolicyConflict -w
` + "```" + `

## Performance Issues

### High Latency
1. Check Authorino cache settings
2. Reduce external metadata calls
3. Enable response caching in policies

### Memory Usage
1. Adjust Limitador storage settings
2. Review counter cardinality
3. Set appropriate history limits

## Getting Help

1. Check policy events: ` + "`kubectl describe <policy-type> <name>`" + `
2. Review controller logs: ` + "`kubectl logs -n kuadrant-system deployment/kuadrant-operator-controller-manager`" + `
3. Enable debug logging for detailed traces
4. Check Kuadrant documentation: https://docs.kuadrant.io
`
	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{
			{
				URI:      params.URI,
				MIMEType: "text/plain",
				Text:     content,
			},
		},
	}, nil
}

// addKuadrantResources adds all MCP resources
func addKuadrantResources(server *mcp.Server) {
	// Existing documentation resources
	server.AddResources(
		&mcp.ServerResource{
			Resource: &mcp.Resource{
				URI:         "kuadrant://docs/gateway-api",
				Name:        "Gateway API Overview",
				Description: "Overview of Gateway API and Kuadrant integration",
				MIMEType:    "text/plain",
			},
			Handler: gatewayAPIResourceHandler,
		},
		&mcp.ServerResource{
			Resource: &mcp.Resource{
				URI:         "kuadrant://docs/dnspolicy",
				Name:        "DNSPolicy Reference",
				Description: "Complete DNSPolicy specification and examples",
				MIMEType:    "text/plain",
			},
			Handler: dnsPolicyResourceHandler,
		},
		&mcp.ServerResource{
			Resource: &mcp.Resource{
				URI:         "kuadrant://docs/ratelimitpolicy",
				Name:        "RateLimitPolicy Reference",
				Description: "Complete RateLimitPolicy specification and examples",
				MIMEType:    "text/plain",
			},
			Handler: rateLimitPolicyResourceHandler,
		},
		&mcp.ServerResource{
			Resource: &mcp.Resource{
				URI:         "kuadrant://docs/authpolicy",
				Name:        "AuthPolicy Reference",
				Description: "Complete AuthPolicy specification and examples",
				MIMEType:    "text/plain",
			},
			Handler: authPolicyResourceHandler,
		},
		&mcp.ServerResource{
			Resource: &mcp.Resource{
				URI:         "kuadrant://docs/tlspolicy",
				Name:        "TLSPolicy Reference",
				Description: "Complete TLSPolicy specification and examples",
				MIMEType:    "text/plain",
			},
			Handler: tlsPolicyResourceHandler,
		},
	)

	// New example resources
	server.AddResources(
		&mcp.ServerResource{
			Resource: &mcp.Resource{
				URI:         "kuadrant://examples/basic-setup",
				Name:        "Basic API Setup Example",
				Description: "Complete example of basic API with rate limiting and auth",
				MIMEType:    "text/plain",
			},
			Handler: exampleBasicSetupResourceHandler,
		},
		&mcp.ServerResource{
			Resource: &mcp.Resource{
				URI:         "kuadrant://examples/production-setup",
				Name:        "Production API Setup Example",
				Description: "Full production setup with TLS, DNS, and advanced policies",
				MIMEType:    "text/plain",
			},
			Handler: exampleProductionSetupResourceHandler,
		},
	)

	// Troubleshooting resource
	server.AddResources(
		&mcp.ServerResource{
			Resource: &mcp.Resource{
				URI:         "kuadrant://troubleshooting",
				Name:        "Troubleshooting Guide",
				Description: "Common issues and debugging techniques for Kuadrant",
				MIMEType:    "text/plain",
			},
			Handler: troubleshootingResourceHandler,
		},
	)
}