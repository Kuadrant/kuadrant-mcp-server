# Kuadrant MCP Server

A Model Context Protocol (MCP) server that generates Kuadrant policy manifests. Designed to work alongside [mcp-server-kubernetes](https://github.com/Flux159/mcp-server-kubernetes) for applying resources to clusters.

![Kuadrant MCP Server Demo](kuadrant-mcp-server-demo.gif)

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

## Tools

| Tool | Description |
|------|-------------|
| `create_gateway` | Gateway manifest with Kuadrant annotations |
| `create_httproute` | HTTPRoute manifest |
| `create_dnspolicy` | DNSPolicy for DNS management |
| `create_tlspolicy` | TLSPolicy for certificate management |
| `create_ratelimitpolicy` | RateLimitPolicy for rate limiting |
| `create_tokenratelimitpolicy` | TokenRateLimitPolicy for AI/LLM APIs |
| `create_authpolicy` | AuthPolicy for authentication/authorisation |

**Rate limit format**: Use `limit` and `window` fields (e.g., `"limit": 100, "window": "60s"`).

### Example Prompts

```
Create a Gateway named 'api-gateway' in namespace 'production' with HTTPS on port 443

Create a RateLimitPolicy for HTTPRoute 'api-route' that limits to 100 requests per minute

Set up an AuthPolicy requiring JWT auth from https://auth.example.com

Show me the Kuadrant rate limiting documentation

Help me configure DNS with Route53 for my gateway
```

## Resources

Documentation is fetched from upstream repos and cached for 15 minutes.

| Resource | Description |
|----------|-------------|
| `kuadrant://docs/gateway-api` | Gateway API overview |
| `kuadrant://docs/dnspolicy` | DNSPolicy reference |
| `kuadrant://docs/ratelimitpolicy` | RateLimitPolicy reference |
| `kuadrant://docs/tokenratelimitpolicy` | TokenRateLimitPolicy reference |
| `kuadrant://docs/authpolicy` | AuthPolicy reference |
| `kuadrant://docs/tlspolicy` | TLSPolicy reference |
| `kuadrant://docs/telemetrypolicy` | TelemetryPolicy reference |
| `kuadrant://docs/kuadrant` | Kuadrant CR reference |
| `kuadrant://docs/authorino-features` | Authorino features |
| `kuadrant://docs/planpolicy` | PlanPolicy extension |
| `kuadrant://docs/secure-protect-connect` | Full walkthrough |
| `kuadrant://docs/simple-ratelimiting` | Rate limiting guide |
| `kuadrant://docs/auth-for-developers` | Auth guide |

## Kubernetes Integration

Combine with mcp-server-kubernetes for a complete workflow:

```bash
# Add both servers
claude mcp add -s user kuadrant docker -- run -i --rm ghcr.io/kuadrant/kuadrant-mcp-server:latest
claude mcp add -s user kubernetes npx -- @flux159/mcp-server-kubernetes
```

Then ask Claude to generate and deploy policies in one step.

## API Versions

| Resource | API Version |
|----------|-------------|
| Gateway/HTTPRoute | `gateway.networking.k8s.io/v1` |
| DNSPolicy | `kuadrant.io/v1` |
| TLSPolicy | `kuadrant.io/v1alpha1` |
| RateLimitPolicy | `kuadrant.io/v1` |
| TokenRateLimitPolicy | `kuadrant.io/v1` |
| AuthPolicy | `kuadrant.io/v1` |

## Adding Resources

Add to `resourceMapping` in `resources.go`:

```go
"kuadrant://docs/newpolicy": {
    url:         "https://raw.githubusercontent.com/Kuadrant/kuadrant-operator/main/doc/reference/newpolicy.md",
    name:        "NewPolicy Reference",
    description: "Description",
    fallback:    "# NewPolicy\n\nSee: https://docs.kuadrant.io/...",
},
```

## Releases

See [RELEASE.md](RELEASE.md). Images are published to `ghcr.io/kuadrant/kuadrant-mcp-server`.

## Licence

Apache 2.0
