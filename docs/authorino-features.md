# Authorino Features

Authorino is Kuadrant's authorization service that implements AuthPolicy.

## Core Features

### Authentication Methods
- **API Keys**: Simple key-based authentication
- **JWT/OIDC**: OpenID Connect token validation
- **mTLS**: Mutual TLS certificate authentication
- **Kubernetes tokens**: TokenReview integration
- **OAuth 2.0**: Token introspection
- **Plain HTTP**: Basic/custom authentication

### Authorization Methods
- **OPA/Rego**: Open Policy Agent policies
- **Pattern matching**: JSON pattern authorization
- **Kubernetes RBAC**: SubjectAccessReview
- **External HTTP**: Webhook authorization
- **CEL expressions**: Common Expression Language rules

### Additional Capabilities
- **Metadata injection**: Enrich requests with external data
- **Response manipulation**: Modify responses dynamically
- **Rate limit integration**: Pass identity to rate limiter
- **Caching**: Cache auth decisions for performance
- **Metrics & tracing**: Observability integration

## Example: Multi-method Authentication

```yaml
apiVersion: kuadrant.io/v1
kind: AuthPolicy
metadata:
  name: multi-auth
spec:
  targetRef:
    group: gateway.networking.k8s.io
    kind: HTTPRoute
    name: api
  rules:
    authentication:
      "api-key":
        apiKey:
          selector:
            matchLabels:
              app: myapp
      "jwt":
        jwt:
          issuerUrl: https://auth.example.com
    authorization:
      "rbac":
        patternMatching:
          patterns:
          - selector: auth.identity.role
            operator: eq
            value: admin
    response:
      success:
        headers:
          "x-user-id":
            value: auth.identity.sub
```

For complete documentation, see: https://docs.kuadrant.io/latest/authorino/docs/features/
