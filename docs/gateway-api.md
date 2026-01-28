# Gateway API and Kuadrant

The Gateway API is a Kubernetes API for managing ingress traffic. Kuadrant extends the Gateway API with additional policies for:

- **DNS management** (DNSPolicy) - Automatic DNS record management
- **TLS certificate management** (TLSPolicy) - Automated cert provisioning
- **Rate limiting** (RateLimitPolicy) - Request rate control
- **Authentication and authorization** (AuthPolicy) - Identity and access management

## Enabling Kuadrant on a Gateway

Add the label to your Gateway to enable Kuadrant features:

```yaml
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
```

## Policy Attachment

Policies attach to Gateway API resources:
- **Gateway**: Affects all routes through the gateway
- **HTTPRoute**: Affects specific routes or rules (use sectionName)

## Common Gateway Patterns

### Gateway with Hostname
```yaml
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
```

### Multi-Protocol Gateway
```yaml
listeners:
  - name: http
    protocol: HTTP
    port: 80
  - name: https
    protocol: HTTPS
    port: 443
    tls:
      mode: Terminate
```
