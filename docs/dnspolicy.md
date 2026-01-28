# DNSPolicy

DNSPolicy enables automatic DNS management for Gateway listeners. It creates and manages DNS records in supported providers (AWS Route53, Google Cloud DNS, Azure DNS).

## API Version
```yaml
apiVersion: kuadrant.io/v1
kind: DNSPolicy
```

## Key Features

- **Automatic DNS record creation** based on Gateway listeners
- **Multi-cluster load balancing** with geo and weighted routing
- **Health checking** for endpoint availability
- **Provider integration** via credential secrets

## Basic Example
```yaml
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
```

## Load Balancing Configuration

### Geographic Load Balancing
```yaml
spec:
  loadBalancing:
    geo:
      zones:
        - id: us-east-1
          weight: 100
        - id: eu-west-1
          weight: 100
      defaultGeo: true  # Catch-all for unmatched regions
```

### Weighted Load Balancing
```yaml
spec:
  loadBalancing:
    weighted:
      weight: 120  # Default weight for this cluster
```

## Health Checks
```yaml
spec:
  healthCheck:
    path: /health
    port: 80
    protocol: HTTP
    failureThreshold: 3
    interval: 5min
    additionalHeaders:
      X-Health-Check: "dns-operator"
```

## Provider Credentials

### AWS Route53
```bash
kubectl create secret generic aws-credentials \
  --type=kuadrant.io/aws \
  --from-literal=AWS_ACCESS_KEY_ID=$AWS_ACCESS_KEY_ID \
  --from-literal=AWS_SECRET_ACCESS_KEY=$AWS_SECRET_ACCESS_KEY
```

### Google Cloud DNS
```bash
kubectl create secret generic gcp-credentials \
  --type=kuadrant.io/gcp \
  --from-file=GOOGLE=$HOME/.config/gcloud/application_default_credentials.json
```

## Important Notes

- Only one provider reference is currently supported
- Requires appropriate DNS zones to be pre-configured in the provider
- Changing from non-loadbalanced to loadbalanced requires policy recreation
- Health checks run from the DNS operator pod
