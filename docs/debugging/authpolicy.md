# Debugging AuthPolicy

## STOP - Read This First

**If the status message contains "was not found" or "target ... was not found":**

The Gateway or HTTPRoute **does not exist in the cluster**. The targetRef YAML is correct - don't modify it.

**What to do:**
1. Run: `kubectl get gateway <target-name> -n <namespace>`
2. If you get "NotFound", the Gateway doesn't exist - you need to create it
3. If it exists, check the namespace matches

**DO NOT waste time adding or changing the `group` field in targetRef - it's already correct.**

---

## TargetNotFound Error

**Message:** "target <name> was not found" or "AuthPolicy target ... was not found"

**Root cause:** The Gateway or HTTPRoute referenced in targetRef does not exist in Kubernetes.

**This is NOT a YAML syntax error.** The targetRef structure is correct. The resource is missing.

**Solution:**
1. Verify the target exists: `kubectl get gateway <name> -n <namespace>`
2. If NotFound, create the Gateway/HTTPRoute resource
3. If it exists in a different namespace, move the AuthPolicy to match (they must be in the same namespace)
4. If the name is wrong, fix the `name` field in targetRef (NOT the group or kind)

---

### Missing targetRef Error

**Message says:** "spec.targetRef: Required value"

**What it means:** Your AuthPolicy has no targetRef at all.

**Fix:** Add targetRef with all required fields:
```yaml
spec:
  targetRef:
    group: gateway.networking.k8s.io  # Required
    kind: Gateway                      # or HTTPRoute
    name: my-gateway                   # Must exist
```

---

### Policy Not Enforced

**Status shows:** `Accepted: True` but `Enforced: False`

**Common causes:**
1. **Authorino not running** - Check: `kubectl get pods -n kuadrant-system -l app=authorino`
2. **AuthConfig not created** - Check: `kubectl get authconfig -n <namespace>`
3. **EnvoyFilter missing** - Check: `kubectl get envoyfilter -n kuadrant-system`

**Fix:** Ensure Authorino pods are Running and check operator logs for errors.

---

## Common targetRef Issues

### Missing group field (most common)
```yaml
# WRONG
targetRef:
  kind: Gateway
  name: my-gateway

# CORRECT
targetRef:
  group: gateway.networking.k8s.io
  kind: Gateway
  name: my-gateway
```

### Wrong namespace
AuthPolicy and its target **must be in the same namespace**. If your Gateway is in `istio-system` but your AuthPolicy is in `production`, it won't work.

**Fix:** Move the AuthPolicy to the same namespace as the Gateway, OR target an HTTPRoute in the same namespace instead.

---

## Diagnostic Workflow

1. **Get the AuthPolicy** and read `status.conditions[].message`
2. **If TargetNotFound:** Check if the Gateway/HTTPRoute exists (see above)
3. **If targetRef missing:** Add the targetRef field with group/kind/name
4. **If Accepted but not Enforced:** Check Authorino is running
5. **Check events:** `kubectl describe authpolicy <name> -n <namespace>`

---

## Policy Hierarchy

Multiple AuthPolicies can target the same resource:
- **Gateway-level with `overrides`** - Takes precedence over everything
- **Gateway-level with `defaults`** - Applies only if no route-level policy exists
- **HTTPRoute-level** - Most specific, applies to that route only

If your policy isn't working, check for conflicts with other policies targeting the same Gateway/HTTPRoute.

---

## Authentication Not Working

**All requests succeed without authentication:**
1. Check policy is `Enforced: True`
2. Verify HTTPRoute routes through the protected Gateway
3. Check Istio gateway pods have the Authorino filter configured
4. Look for Gateway-level policy with `overrides` that might be disabling auth

**All requests fail with 401/403:**
1. Check auth rule selectors aren't too restrictive
2. Verify OIDC discovery URLs are accessible from the cluster
3. Check API key secrets exist and are in correct format
4. Review Authorino pod logs while sending requests

---

## Related Resources

- Status Conditions: `kuadrant://debug/status-conditions`
- Policy Conflicts: `kuadrant://debug/policy-conflicts`
- Authorino Features: `kuadrant://docs/authorino-features`
