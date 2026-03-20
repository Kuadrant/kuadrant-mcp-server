# Feature: Debugging-Focused MCP Server

## Summary

Pivot the kuadrant-mcp-server from a manifest-generation MCP to a debugging-focused MCP that serves structured prompts and embedded resources for troubleshooting Kuadrant installations. The server provides guided debugging workflows via MCP prompts, backed by embedded markdown documentation, and delegates all cluster interaction to a companion Kubernetes MCP server.

## Goals

- Provide structured debugging prompts for all Kuadrant policy types and installation issues
- Embed debugging guides into the binary for offline, zero-network operation
- Cover both platform engineers (production issues) and developers (local/staging issues)
- Direct the LLM to use a companion Kubernetes MCP server for cluster queries — no kubectl instructions
- Support Istio as the primary gateway provider with Istio-specific debugging guidance

## Non-Goals

- Active cluster interaction (no tools that query Kubernetes directly)
- Manifest generation (removed entirely)
- Runtime doc fetching from upstream GitHub repos
- Supporting gateway providers other than Istio (may be added later)

## Design

### Backwards Compatibility

This is a complete pivot. All existing tools and resources are removed. Users relying on the manifest-generation tools will need to stop using this server for that purpose. The server name and transport options remain the same.

### Architecture Changes

```
kuadrant-mcp-server (before)
├── main.go        — 6 manifest-generation tools + transport
├── resources.go   — 13 resources fetched via HTTP from GitHub
└── process-docs.go / update-docs.sh — doc extraction tooling

kuadrant-mcp-server (after)
├── main.go        — Server setup, prompt + resource registration, transport
├── prompts.go     — 10 prompt definitions and handlers
├── resources.go   — 10 resources served from embedded filesystem
└── docs/
    └── debugging/ — Embedded markdown debugging guides (10 files)
```

Key changes:
- `go:embed` bundles `docs/debugging/*.md` into the binary
- No HTTP client, no caching, no network dependency for serving content
- `gopkg.in/yaml.v3` dependency removed (no YAML generation)
- `process-docs.go` and `update-docs.sh` removed (no upstream doc extraction)

### API Changes

No CRD or Kubernetes API changes. The MCP API surface changes as follows:

**Tools: all removed**

The 6 manifest-generation tools (`create_gateway`, `create_httproute`, `create_dnspolicy`, `create_tlspolicy`, `create_ratelimitpolicy`, `create_authpolicy`) are removed.

**Prompts: 10 new**

| Prompt Name | Description | Arguments |
|---|---|---|
| `debug-installation` | Verify operator, CRDs, Kuadrant CR, Istio, Limitador, Authorino | `namespace` (default: `kuadrant-system`) |
| `debug-gateway` | Gateway not accepting traffic, listeners, Istio proxy | `gateway-name` (required), `namespace` (default: `kuadrant-system`) |
| `debug-dnspolicy` | DNS records not created, provider config, zone issues | `policy-name` (required), `namespace` (optional) |
| `debug-tlspolicy` | Certificates not issuing, issuer problems, cert-manager | `policy-name` (required), `namespace` (optional) |
| `debug-ratelimitpolicy` | Rate limits not enforced, Limitador health, targeting | `policy-name` (required), `namespace` (optional) |
| `debug-authpolicy` | Auth not enforced, Authorino health, rule matching | `policy-name` (required), `namespace` (optional) |
| `debug-telemetrypolicy` | Custom metrics not appearing, CEL expression issues | `policy-name` (required), `namespace` (optional) |
| `debug-tokenratelimitpolicy` | Token-based rate limiting not working | `policy-name` (required), `namespace` (optional) |
| `debug-policy-status` | Interpret status conditions on any policy | `policy-name` (required), `namespace` (optional), `policy-kind` (required) |
| `debug-policy-conflicts` | Override/default conflicts, policy hierarchy | `namespace` (optional) |

Each prompt follows a consistent output structure:
1. **Context** — what the component/policy does and common failure modes
2. **Prerequisites** — what should be checked first
3. **Diagnostic steps** — ordered steps directing the LLM to use the Kubernetes MCP server (get resource, list pods, get events, read logs)
4. **Resource references** — URIs to embedded debugging resources
5. **Common fixes** — most frequent resolutions

Prompts use generic Kubernetes MCP tool patterns (e.g. "use the kubernetes MCP server to get the resource") rather than tool names from a specific implementation.

Arguments with no default instruct the LLM to ask the user if not provided.

**Resources: 10 new (replacing 13 old)**

| URI | Embedded File | Description |
|---|---|---|
| `kuadrant://debug/installation` | `docs/debugging/installation.md` | Operator, CRDs, Kuadrant CR, Istio health |
| `kuadrant://debug/gateway-istio` | `docs/debugging/gateway-istio.md` | Istio gateway proxy, listeners, envoy config |
| `kuadrant://debug/dnspolicy` | `docs/debugging/dnspolicy.md` | DNS provider, zone config, record creation |
| `kuadrant://debug/tlspolicy` | `docs/debugging/tlspolicy.md` | cert-manager, issuer, certificate lifecycle |
| `kuadrant://debug/ratelimitpolicy` | `docs/debugging/ratelimitpolicy.md` | Limitador health, rate limit enforcement |
| `kuadrant://debug/authpolicy` | `docs/debugging/authpolicy.md` | Authorino health, auth rule matching |
| `kuadrant://debug/telemetrypolicy` | `docs/debugging/telemetrypolicy.md` | Custom metrics, CEL expressions |
| `kuadrant://debug/tokenratelimitpolicy` | `docs/debugging/tokenratelimitpolicy.md` | Token-based rate limiting |
| `kuadrant://debug/status-conditions` | `docs/debugging/status-conditions.md` | All status conditions across all policy types |
| `kuadrant://debug/policy-conflicts` | `docs/debugging/policy-conflicts.md` | Override/default hierarchy, multi-policy resolution |

### Component Changes

**main.go:**
- Remove all tool parameter structs and handler functions
- Remove `server.AddTools()` calls
- Remove `validateWindow` helper
- Remove `gopkg.in/yaml.v3` import
- Add calls to `addDebugPrompts(server)` and `addDebugResources(server)`
- Transport handling (stdio/sse/http) remains unchanged

**resources.go (rewrite):**
- Remove `docCache`, `cachedDoc`, HTTP client, `fetch()`, `fallbackOrError()`
- Remove `docSource` struct and HTTP-based `resourceMapping`
- Add `//go:embed docs/debugging/*.md` directive
- New `resourceDef` struct mapping URI to embedded path, name, description
- `addDebugResources()` reads from `embed.FS` and registers as MCP resources

**prompts.go (new):**
- Prompt definitions with argument schemas
- Handler functions that build the structured debugging output
- Template substitution for arguments (policy-name, namespace)
- `addDebugPrompts()` registers all prompts with the server

**Removed files:**
- `process-docs.go`
- `update-docs.sh`

### Security Considerations

- No cluster access — the server never touches Kubernetes directly
- Embedded content is static and compiled in — no risk of fetching malicious content at runtime
- No credentials or secrets involved
- The companion Kubernetes MCP server handles its own RBAC/auth

## Testing Strategy

- **Unit tests**: Prompt handlers return expected output structure for given arguments. Resource handlers serve correct embedded content for each URI.
- **Integration tests**: MCP protocol-level tests — initialize server, call `prompts/list`, invoke a prompt, call `resources/list`, read a resource. Verify JSON-RPC responses.
- **Manual testing**: End-to-end with Claude Code + Kubernetes MCP server against a live cluster.

## Use Cases

### Platform Admin: DNS records not being created for gateways

**Persona:** Platform engineer responsible for Kuadrant installation and infrastructure.

**Scenario:** A DNSPolicy targeting a Gateway is created but no DNS records appear in Route53. The policy may show Accepted but records are missing from the hosted zone.

**Prompt:** `debug-dnspolicy` with `policy-name` argument.

**Diagnostic flow:**
1. Check DNSPolicy .status.conditions (Accepted/Enforced)
2. Verify target Gateway exists and listeners have explicit hostnames
3. Check DNS provider credentials Secret exists and has correct Route53 permissions (route53:ChangeResourceRecordSets, route53:ListHostedZones)
4. List DNSRecord CRs — these are the intermediate resources DNSPolicy creates. Check their .status for provider-specific errors
5. Check kuadrant-operator logs for DNS reconciliation errors
6. Verify Route53 hosted zone covers the listener hostnames

**Common fixes:**
- Credentials Secret missing or in wrong namespace
- IAM permissions insufficient for Route53 operations
- Gateway listeners missing explicit hostnames
- Hosted zone doesn't cover the listener hostnames
- targetRef.group missing from the policy

**Resources used:** `kuadrant://debug/dnspolicy`, `kuadrant://debug/status-conditions`

### Developer: RateLimitPolicy not enforcing rate limits

**Persona:** Application developer creating policies for their services.

**Scenario:** A RateLimitPolicy is created targeting a Gateway or HTTPRoute, but traffic is not being rate limited. Requests that should return 429 are passing through.

**Prompt:** `debug-ratelimitpolicy` with `policy-name` argument.

**Diagnostic flow:**
1. Check RateLimitPolicy .status.conditions (Accepted/Enforced)
2. Verify targetRef points to an existing Gateway or HTTPRoute with correct kind, name, and group
3. Inspect rate limit configuration — rates must have limit (int) and window (duration string)
4. Check for policy conflicts — multiple policies targeting the same resource, defaults vs overrides hierarchy
5. Verify Limitador pods are running and processing the policy (check logs)
6. Test by sending requests exceeding the limit within the window
7. Check Istio/Envoy — verify EnvoyFilter resources exist for the policy

**Common fixes:**
- targetRef.group missing (should be "gateway.networking.k8s.io")
- targetRef points to wrong or non-existent resource
- Policy Accepted but not Enforced — Limitador not running
- Targeting Gateway when per-route limiting is needed (should target HTTPRoute)
- "when" conditions too restrictive — limits never triggered
- Gateway-level overrides taking precedence over route-level policy
- Limitador crashlooping due to config errors or Redis connectivity

**Resources used:** `kuadrant://debug/ratelimitpolicy`, `kuadrant://debug/status-conditions`, `kuadrant://debug/policy-conflicts`

## Open Questions

- None currently

## Execution

### Todo

- [ ] Scaffold embedded docs structure and write debugging guides
  - [ ] `docs/debugging/installation.md`
  - [ ] `docs/debugging/gateway-istio.md`
  - [ ] `docs/debugging/dnspolicy.md`
  - [ ] `docs/debugging/tlspolicy.md`
  - [ ] `docs/debugging/ratelimitpolicy.md`
  - [ ] `docs/debugging/authpolicy.md`
  - [ ] `docs/debugging/telemetrypolicy.md`
  - [ ] `docs/debugging/tokenratelimitpolicy.md`
  - [ ] `docs/debugging/status-conditions.md`
  - [ ] `docs/debugging/policy-conflicts.md`
- [ ] Rewrite `resources.go` to serve from embedded FS
  - [ ] Unit tests for resource serving
- [ ] Create `prompts.go` with all prompt definitions and handlers
  - [ ] Unit tests for prompt handlers
- [ ] Simplify `main.go` — remove tools, wire up prompts + resources
- [ ] Remove `process-docs.go`, `update-docs.sh`, `gopkg.in/yaml.v3` dependency
- [ ] Integration tests (MCP protocol-level)
- [ ] Update `CLAUDE.md` and `README.md`
- [ ] Update Dockerfile if needed

### Completed

## Change Log

### 2026-03-20 — Initial design

- Decided to pivot from manifest-generation to debugging-focused MCP
- Chose Approach A: prompts as structured workflows + resources as reference material
- Embedded docs (go:embed) over runtime HTTP fetching for offline reliability
- No tools — all cluster interaction via companion Kubernetes MCP server
- Istio as primary gateway provider
- Generic Kubernetes MCP tool references in prompts (not vendor-specific)
- Namespace arguments default sensibly but are always overridable
