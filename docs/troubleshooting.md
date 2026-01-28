# Troubleshooting Guide

Common issues and solutions when working with Kuadrant policies.

## Policy Not Taking Effect

### Check Policy Status
```bash
# Check RateLimitPolicy status
kubectl get ratelimitpolicy -n <namespace> <policy-name> -o yaml

# Look for status conditions
status:
  conditions:
  - type: Available
    status: "True"    # Should be True
    reason: Available
```

### Common Causes:
1. **Wrong targetRef** - Ensure namespace and name are correct
2. **Policy conflicts** - Check for multiple policies targeting same resource
3. **Gateway not labeled** - Ensure `kuadrant.io/gateway: "true"` label

## Rate Limiting Issues

### Rate Limits Not Applied
```bash
# Check Limitador deployment
kubectl get deployment limitador-limitador -n kuadrant-system

# Check for rate limit configurations
kubectl get configmap limitador-config -n kuadrant-system -o yaml
```

### Debug Rate Limit Counters
```bash
# Port-forward to Limitador
kubectl port-forward -n kuadrant-system svc/limitador-limitador 8080:8080

# Check counters
curl http://localhost:8080/counters
```

## Authentication Failures

### 401 Unauthorized Errors
1. **Check Authorino logs**:
```bash
kubectl logs -n kuadrant-system deployment/authorino -f | grep <request-id>
```

2. **Verify API Key Secret**:
```bash
# Check secret has correct labels
kubectl get secret <api-key-secret> -o yaml

# Required labels:
# - authorino.kuadrant.io/managed-by: authorino
# - Your custom selector labels
```

3. **JWT Issues**:
- Verify issuer URL is accessible
- Check token expiration
- Validate audience claim matches policy

## DNS Not Resolving

### Check DNSRecord Status
```bash
kubectl get dnsrecord -A
kubectl describe dnsrecord <record-name> -n <namespace>
```

### Common Issues:
1. **Provider credentials** - Verify secret exists and has correct type
2. **DNS zone** - Ensure zone exists in provider
3. **Health checks failing** - Check endpoint accessibility

## TLS Certificate Issues

### Certificate Not Issued
```bash
# Check Certificate resource
kubectl get certificate -n <namespace>
kubectl describe certificate <cert-name> -n <namespace>

# Check cert-manager logs
kubectl logs -n cert-manager deployment/cert-manager
```

### Common Problems:
1. **ACME challenges failing** - Ensure HTTP01/DNS01 solver can complete
2. **Rate limits** - Let's Encrypt has rate limits; use staging for testing
3. **Issuer problems** - Verify ClusterIssuer/Issuer configuration

## Gateway Issues

### Gateway Not Ready
```bash
kubectl get gateway -n <namespace> <gateway-name> -o yaml

# Check status
status:
  conditions:
  - type: Programmed
    status: "True"  # Should be True
  - type: Accepted
    status: "True"  # Should be True
```

### No Address Assigned
- Check Gateway controller (Istio/Envoy Gateway) is running
- Verify LoadBalancer service is created
- Check cloud provider load balancer provisioning

## Debugging Tools

### Enable Debug Logging
```yaml
# For Authorino
kubectl set env deployment/authorino -n kuadrant-system LOG_LEVEL=debug

# For Limitador
kubectl set env deployment/limitador-limitador -n kuadrant-system RUST_LOG=debug
```

### Useful Commands
```bash
# Get all Kuadrant resources
kubectl get gateway,httproute,ratelimitpolicy,authpolicy,dnspolicy,tlspolicy -A

# Check policy conditions
kubectl get ratelimitpolicy,authpolicy -A -o json | jq '.items[] | select(.status.conditions[]?.status != "True") | {name: .metadata.name, namespace: .metadata.namespace, conditions: .status.conditions}'

# Watch events
kubectl get events -A --field-selector reason=PolicyConflict -w
```

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

1. Check policy events: `kubectl describe <policy-type> <name>`
2. Review controller logs: `kubectl logs -n kuadrant-system deployment/kuadrant-operator-controller-manager`
3. Enable debug logging for detailed traces
4. Check Kuadrant documentation: https://docs.kuadrant.io
