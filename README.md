# Kuadrant MCP Server

A Model Context Protocol (MCP) server for debugging Kuadrant installations. Provides structured debugging prompts and embedded troubleshooting guides. Designed to work alongside a Kubernetes MCP server (e.g. [mcp-server-kubernetes](https://github.com/Flux159/mcp-server-kubernetes)) for cluster interaction.

## Quick Start

```bash
# Add the MCP server (available in all projects)
claude mcp add -s user kuadrant docker -- run -i --rm ghcr.io/kuadrant/kuadrant-mcp-server:latest

# Verify
claude mcp list

# Start using
claude
```

## Installation

```bash
# Docker (recommended)
docker pull ghcr.io/kuadrant/kuadrant-mcp-server:latest

# Go
go install github.com/kuadrant/kuadrant-mcp-server@latest

# From source
git clone https://github.com/kuadrant/kuadrant-mcp-server && cd kuadrant-mcp-server
go build -o kuadrant-mcp-server
```

## Usage

```bash
# stdio (default)
./kuadrant-mcp-server

# SSE transport
./kuadrant-mcp-server -transport sse -addr :8080

# HTTP transport
./kuadrant-mcp-server -transport http -addr :8080

# Docker
docker run -i --rm ghcr.io/kuadrant/kuadrant-mcp-server:latest
```

### MCP Client Configuration

```json
{
  "mcpServers": {
    "kuadrant": {
      "command": "docker",
      "args": ["run", "-i", "--rm", "ghcr.io/kuadrant/kuadrant-mcp-server:latest"]
    }
  }
}
```

## Prompts

Structured debugging workflows that guide the LLM through diagnostic steps using a companion Kubernetes MCP server.

| Prompt | Description |
|--------|-------------|
| `debug-installation` | Verify operator, CRDs, Kuadrant CR, Istio, Limitador, Authorino |
| `debug-gateway` | Gateway not accepting traffic, listeners, Istio proxy |
| `debug-dnspolicy` | DNS records not created, provider config, zone issues |
| `debug-tlspolicy` | Certificates not issuing, issuer problems, cert-manager |
| `debug-ratelimitpolicy` | Rate limits not enforced, Limitador health, targeting |
| `debug-authpolicy` | Auth not enforced, Authorino health, rule matching |
| `debug-telemetrypolicy` | Custom metrics not appearing, CEL expression issues |
| `debug-tokenratelimitpolicy` | Token-based rate limiting not working |
| `debug-policy-status` | Interpret status conditions on any policy |
| `debug-policy-conflicts` | Override/default conflicts, policy hierarchy |

### Example Usage

```
Debug my Kuadrant installation in the kuadrant-system namespace

Why isn't my RateLimitPolicy 'api-limits' being enforced?

Help me understand the status conditions on my AuthPolicy

My DNSPolicy isn't creating DNS records - what's wrong?

Check if there are policy conflicts in the production namespace
```

## Resources

Embedded debugging guides bundled into the binary. No network access required.

| Resource | Description |
|----------|-------------|
| `kuadrant://debug/installation` | Operator, CRDs, Kuadrant CR, Istio health |
| `kuadrant://debug/gateway-istio` | Istio gateway proxy, listeners, envoy config |
| `kuadrant://debug/dnspolicy` | DNS provider, zone config, record creation |
| `kuadrant://debug/tlspolicy` | cert-manager, issuer, certificate lifecycle |
| `kuadrant://debug/ratelimitpolicy` | Limitador health, rate limit enforcement |
| `kuadrant://debug/authpolicy` | Authorino health, auth rule matching |
| `kuadrant://debug/telemetrypolicy` | Custom metrics, CEL expressions |
| `kuadrant://debug/tokenratelimitpolicy` | Token-based rate limiting |
| `kuadrant://debug/status-conditions` | All status conditions across all policy types |
| `kuadrant://debug/policy-conflicts` | Override/default hierarchy, multi-policy resolution |

## Kubernetes Integration

Combine with a Kubernetes MCP server for a complete debugging workflow:

```bash
# Add both servers
claude mcp add -s user kuadrant docker -- run -i --rm ghcr.io/kuadrant/kuadrant-mcp-server:latest
claude mcp add -s user kubernetes npx -- @flux159/mcp-server-kubernetes
```

The debugging prompts direct the LLM to use the Kubernetes MCP server for cluster queries — checking pod status, reading resource conditions, fetching events, and reading logs.

## Releases

See [RELEASE.md](RELEASE.md). Images are published to `ghcr.io/kuadrant/kuadrant-mcp-server`.

## Licence

Apache 2.0
