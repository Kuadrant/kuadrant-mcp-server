# Kuadrant Policy Status Conditions Reference

## Overview

All Kuadrant policies (AuthPolicy, RateLimitPolicy, DNSPolicy, TLSPolicy, etc.) use Kubernetes-style status conditions to report their state. Understanding these conditions is critical for debugging.

## Standard Conditions

### Accepted

**Meaning:** The policy spec is valid and the target resource exists.

**Possible States:**

| Status | Reason | Meaning |
|--------|--------|---------|
| True | Accepted | Policy is syntactically valid and target is found |
| False | TargetNotFound | The targetRef points to a non-existent resource |
| False | Conflicted | Policy conflicts with another policy (hierarchy/override issue) |
| False | Invalid | Policy spec has validation errors |
| False | InvalidTarget | Target resource exists but is invalid for this policy type |

**Actions when Accepted=False:**

1. **TargetNotFound:**
   - Check targetRef.group is set (commonly missed!)
   - Verify targetRef.name matches exactly
   - Confirm policy and target are in same namespace
   - Verify target resource actually exists

2. **Conflicted:**
   - List all policies in the namespace targeting same resource
   - Check policy hierarchy (Gateway overrides > defaults > HTTPRoute)
   - Review defaults vs overrides sections

3. **Invalid:**
   - Check the message field for specific validation error
   - Review policy spec against CRD schema
   - Check required fields are present

4. **InvalidTarget:**
   - Verify target is the correct kind (e.g., Gateway not HTTPRoute)
   - Check target has required fields for this policy type

### Enforced

**Meaning:** The policy has been applied and is actively affecting traffic.

**Possible States:**

| Status | Reason | Meaning |
|--------|--------|---------|
| True | Enforced | Policy is active and affecting traffic |
| False | Unknown | Backend service (Limitador/Authorino) is unavailable |
| False | Overridden | Another policy with higher precedence overrides this one |
| False | TargetNotReady | Target resource exists but is not ready |

**Actions when Enforced=False:**

1. **Unknown (AuthPolicy):**
   - Check Authorino pods are running in kuadrant-system
   - Verify Authorino operator is healthy
   - Check AuthConfig resources were created
   - Review Authorino logs for errors

2. **Unknown (RateLimitPolicy):**
   - Check Limitador pods are running in kuadrant-system
   - Verify Limitador has valid configuration
   - Check for Redis connectivity issues (if using Redis backend)
   - Review Limitador logs

3. **Unknown (DNSPolicy):**
   - Check DNS provider credentials are valid
   - Verify external-dns or similar DNS controller is running
   - Check DNSRecord resources for errors

4. **Unknown (TLSPolicy):**
   - Check cert-manager is installed and running
   - Verify issuer referenced in issuerRef exists
   - Check Certificate resources for issues
   - Review cert-manager logs

5. **Overridden:**
   - Review policy hierarchy
   - Check if a Gateway-level policy with overrides section exists
   - Consider whether this is intentional

6. **TargetNotReady:**
   - Check target Gateway/HTTPRoute status conditions
   - Verify target has listeners configured (for Gateway)
   - Verify target has parentRefs (for HTTPRoute)

## Policy-Specific Conditions

### DNSPolicy

Additional conditions:

- **RecordReady:** DNS records have been created successfully
  - True = DNS records exist in provider
  - False = DNS record creation failed (check DNSRecord resources)

### TLSPolicy

Additional conditions:

- **CertificateReady:** Certificate has been issued
  - True = Certificate is valid and not expiring soon
  - False = Certificate issue failed or is expiring (check Certificate resource)

### RateLimitPolicy

Look for:
- **LimitadorAvailable:** Limitador service is reachable
  - False = Check Limitador pod health and service endpoints

### AuthPolicy

Look for:
- **AuthorinoAvailable:** Authorino service is reachable
  - False = Check Authorino pod health and service endpoints

## Condition Structure

Conditions follow Kubernetes conventions:

```yaml
status:
  conditions:
  - type: Accepted
    status: "True"  # or "False" or "Unknown"
    reason: Accepted  # Machine-readable reason code
    message: "Policy has been accepted"  # Human-readable details
    lastTransitionTime: "2024-01-15T10:30:00Z"
    observedGeneration: 2
```

**Fields:**
- `type`: Condition name (Accepted, Enforced, etc.)
- `status`: "True", "False", or "Unknown"
- `reason`: Short machine-readable code
- `message`: Detailed human-readable explanation
- `lastTransitionTime`: When this condition last changed
- `observedGeneration`: Which generation of the resource this status reflects

## How to Read Status

### Using Kubernetes MCP

```
Request: Get the policy resource (AuthPolicy, RateLimitPolicy, etc.)
Look for: status.conditions[] array
```

### Check All Conditions

A healthy policy should have:
```yaml
status:
  conditions:
  - type: Accepted
    status: "True"
  - type: Enforced
    status: "True"
```

### Prioritize Conditions

Check in this order:
1. **Accepted** - If False, fix targetRef or spec issues first
2. **Enforced** - If False, investigate backend services (Authorino/Limitador/etc.)
3. **Policy-specific** - Check additional conditions for that policy type

## Common Status Patterns

### Pattern 1: TargetNotFound

```yaml
status:
  conditions:
  - type: Accepted
    status: "False"
    reason: TargetNotFound
    message: "targetRef not found: Gateway.gateway.networking.k8s.io my-gateway not found"
```

**Diagnosis:** targetRef.group likely missing or target doesn't exist

**Fix:**
```yaml
spec:
  targetRef:
    group: gateway.networking.k8s.io  # Add this
    kind: Gateway
    name: my-gateway
```

### Pattern 2: Policy Accepted but Not Enforced

```yaml
status:
  conditions:
  - type: Accepted
    status: "True"
  - type: Enforced
    status: "False"
    reason: Unknown
    message: "Authorino service is not available"
```

**Diagnosis:** Backend service (Authorino/Limitador) is down

**Fix:** Check backend pods in kuadrant-system namespace

### Pattern 3: Policy Overridden

```yaml
status:
  conditions:
  - type: Accepted
    status: "True"
  - type: Enforced
    status: "False"
    reason: Overridden
    message: "Policy overridden by gateway-level policy 'gateway-defaults'"
```

**Diagnosis:** Another policy with higher precedence exists

**Fix:** Check if this is intentional, review policy hierarchy

### Pattern 4: Invalid Spec

```yaml
status:
  conditions:
  - type: Accepted
    status: "False"
    reason: Invalid
    message: "spec.limits.global.rates[0].window: Invalid value: \"60\": must be a duration string like '60s'"
```

**Diagnosis:** Validation error in policy spec

**Fix:** Correct the specific field mentioned in message

## Debugging Workflow

1. **Get the policy**
   ```
   Use kubernetes MCP to retrieve the full policy resource
   ```

2. **Check Accepted condition**
   - If False, focus on targetRef and spec validation
   - If True, move to next step

3. **Check Enforced condition**
   - If False, focus on backend services and target readiness
   - If True, policy is working (investigate traffic/test issues)

4. **Check policy-specific conditions**
   - Look for additional conditions specific to policy type
   - These provide more detailed diagnostics

5. **Check observedGeneration**
   ```yaml
   status:
     observedGeneration: 2
   metadata:
     generation: 3
   ```
   If observedGeneration < generation, status is stale (controller hasn't processed latest spec yet)

## Related Resources

- AuthPolicy Debugging: `kuadrant://debug/authpolicy`
- RateLimitPolicy Debugging: `kuadrant://debug/ratelimitpolicy`
- DNSPolicy Debugging: `kuadrant://debug/dnspolicy`
- TLSPolicy Debugging: `kuadrant://debug/tlspolicy`
- Policy Conflicts: `kuadrant://debug/policy-conflicts`
