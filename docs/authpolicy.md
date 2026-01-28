# AuthPolicy

AuthPolicy provides authentication and authorization for APIs using various identity sources and access control methods.

## API Version
```yaml
apiVersion: kuadrant.io/v1
kind: AuthPolicy
```

## Authentication Methods

### API Key Authentication
```yaml
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
```

API Key Secret:
```yaml
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
```

### JWT Authentication
```yaml
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
```

### Kubernetes Token Review
```yaml
rules:
  authentication:
    "k8s-tokens":
      kubernetes:
        audiences:
          - https://kubernetes.default.svc.cluster.local
```

## Authorization Methods

### OPA/Rego Authorization
```yaml
rules:
  authorization:
    "admin-only":
      opa:
        rego: |
          groups := split(object.get(input.auth.identity.metadata.annotations, "kuadrant.io/groups", ""), ",")
          allow { groups[_] == "admin" }
```

### Kubernetes SubjectAccessReview
```yaml
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
```

### Pattern Matching
```yaml
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
```

## Advanced Features

### Gateway-Level Deny All
```yaml
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
```

### Custom Response Headers
```yaml
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
```

### Metadata Enrichment
```yaml
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
```

## Important Patterns

1. **Anonymous Access**: At least one auth rule must pass (unless using anonymous)
2. **All Authorization Must Pass**: All authorization rules must allow the request
3. **Defaults vs Overrides**: Use defaults for mergeable policies, overrides for strict enforcement
4. **Priority Groups**: Control authentication method precedence with priority values
