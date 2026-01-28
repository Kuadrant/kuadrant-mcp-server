# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Quick Reference

**Main Files:**
- `main.go` - Server setup, tool registration, and tool handlers
- `resources.go` - Documentation resource handlers
- `update-docs.sh` - Extract documentation from official repos
- `process-docs.go` - Convert markdown docs to Go handlers

**Key Commands:**
```bash
go build -o kuadrant-mcp-server  # Build server
./update-docs.sh                 # Update documentation
go run process-docs.go           # Generate resource handlers
```

## Project Overview

This is a Model Context Protocol (MCP) server that generates Kubernetes manifests for Kuadrant policies. It provides tools for creating Gateway API resources with Kuadrant-specific annotations and policies.

### Current Status
- **Tools**: All 6 manifest generation tools with full JSON schema support
- **Resources**: 11 documentation resources (policies, examples, troubleshooting, Authorino features)
- **Compatibility**: Works with Claude Code CLI and Claude Desktop
- **Integration**: Designed to work with Kubernetes MCP servers
- **Docker**: Images published to ghcr.io/jasonmadigan/kuadrant-mcp-server
- **Documentation**: Auto-extraction from official Kuadrant repos via mkdocs.yml

## Common Development Commands

### Building and Running

```bash
# Build the binary locally
go build -o kuadrant-mcp-server

# Run the server with different transports
./kuadrant-mcp-server                          # stdio (default)
./kuadrant-mcp-server -transport sse -addr :8080    # SSE transport
./kuadrant-mcp-server -transport http -addr :8080   # StreamableHTTP transport

# Build Docker image
docker build -t kuadrant-mcp-server:latest .

# Run with Docker
docker run -i --rm kuadrant-mcp-server:latest  # stdio
docker run -i --rm -p 8080:8080 kuadrant-mcp-server:latest -transport sse -addr :8080  # SSE
docker run -i --rm -p 8080:8080 kuadrant-mcp-server:latest -transport http -addr :8080 # HTTP

# Use docker-compose
docker-compose build
docker-compose run --rm kuadrant-mcp
```

### Development

```bash
# Download dependencies
go mod download

# Update dependencies
go mod tidy

# Format code (Go standard formatting)
go fmt ./...

# Vet code for common issues
go vet ./...
```

### Testing

Currently, there are no tests in the project. When adding tests:

```bash
# Run tests
go test ./...

# Run tests with coverage
go test -cover ./...
```

## Architecture Overview

This is a Model Context Protocol (MCP) server that generates Kubernetes manifests for Kuadrant policies. The architecture is straightforward:

### Transport Support

The server supports three transport protocols:
1. **stdio** (default) - For CLI tools and desktop applications
2. **sse** - Server-Sent Events for web-based clients
3. **http** - StreamableHTTP for modern web applications with session management

### Core Components

1. **MCP Server Framework** (`github.com/modelcontextprotocol/go-sdk/mcp`)
   - Handles stdio-based communication protocol
   - Manages tool registration and execution
   - Provides resource serving capabilities

2. **Tool Handlers** (main.go)
   - Each handler generates a specific Kuadrant CRD manifest
   - All handlers follow the same pattern:
     - Extract and validate arguments
     - Build resource map structure
     - Marshal to YAML
     - Return formatted YAML string

3. **Resource Handlers** (resources.go)
   - Serve documentation as MCP resources
   - Provide examples and specifications for each policy type
   - Include TokenRateLimitPolicy, Kuadrant CR, and Authorino features

### Manifest Generation Pattern

Each tool handler follows this structure:
1. Extract required fields (name, namespace, targetRef)
2. Validate required parameters
3. Set default values for groups if not specified
4. Build nested map structure matching Kubernetes API
5. Marshal to YAML using `gopkg.in/yaml.v3`
6. Return YAML string or error message

### Kuadrant Policy Types

- **Gateway**: Entry point with Kuadrant annotations (gateway.networking.k8s.io/v1)
- **HTTPRoute**: Route configuration (gateway.networking.k8s.io/v1)
- **DNSPolicy**: DNS management (kuadrant.io/v1)
- **TLSPolicy**: Certificate management (kuadrant.io/v1alpha1)
- **RateLimitPolicy**: Rate limiting (kuadrant.io/v1)
- **TokenRateLimitPolicy**: Token-based rate limiting for AI/LLM services (kuadrant.io/v1)
- **AuthPolicy**: Authentication/authorization (kuadrant.io/v1)
- **TelemetryPolicy**: Custom metrics labels via CEL (kuadrant.io/v1alpha1)
- **Kuadrant CR**: Main operator configuration (kuadrant.io/v1beta1)
- **PlanPolicy** (extension): Plan-based rate limiting for tiered services (extensions.kuadrant.io/v1alpha1)

### Error Handling

- Input validation returns user-friendly error messages
- YAML marshaling errors are caught and reported
- All errors returned as text content, not exceptions

### Integration Points

Designed to work with Kubernetes MCP servers:
- Generated YAML can be applied via k8s-mcp-server
- Compatible with mcp-server-kubernetes (npm install -g @flux159/mcp-server-kubernetes)
- Outputs match actual Kuadrant CRD specifications

### Kubernetes MCP Server Integration

The recommended companion is `mcp-server-kubernetes` which:
- Uses your `~/.kube/config` automatically
- Connects to current kubectl context
- Provides full Kubernetes resource management
- Supports safe mode to prevent destructive operations

Configuration example:
```json
{
  "mcpServers": {
    "kuadrant": {
      "command": "docker",
      "args": ["run", "-i", "--rm", "ghcr.io/jasonmadigan/kuadrant-mcp-server:latest"]
    },
    "kubernetes": {
      "command": "npx",
      "args": ["@flux159/mcp-server-kubernetes"]
    }
  }
}
```

## Key Implementation Details

- **No external dependencies** beyond MCP SDK and YAML library
- **Stateless operation** - each request is independent
- **Type safety** through map[string]interface{} with validation
- **Flexible input** - supports both new and legacy field names
- **Default values** align with Kuadrant best practices

## MCP Go SDK Usage Notes

### Tool InputSchema Requirements

When creating tools with the MCP Go SDK, the `InputSchema` field is crucial for tool discovery in Claude Desktop. Without proper schemas, tools show as "Disabled".

**Current Implementation** (Working):
The server uses `NewServerTool` with type parameters, which automatically generates schemas from Go structs:

```go
// Define parameter struct with jsonschema tags
type CreateGatewayParams struct {
    Name      string `json:"name" jsonschema:"required,description=Name of the Gateway resource"`
    Namespace string `json:"namespace" jsonschema:"required,description=Kubernetes namespace"`
    // ... other fields
}

// Register tool with automatic schema generation
server.AddTools(
    mcp.NewServerTool(
        "create_gateway",
        "Generate a Gateway manifest with Kuadrant annotations",
        func(ctx context.Context, _ *mcp.ServerSession, params *mcp.CallToolParamsFor[CreateGatewayParams]) (*mcp.CallToolResultFor[string], error) {
            result, err := createGatewayHandler(ctx, params.Arguments)
            // ... handle result
        },
    ),
)
```

This approach:
- Automatically generates JSON schemas from struct tags
- Provides parameter validation
- Enables tool discovery in Claude Desktop
- Shows proper parameter hints in Claude

### Key Points:
- The `jsonschema.Schema` struct from `github.com/modelcontextprotocol/go-sdk/jsonschema` provides full JSON Schema support
- `InputSchema` cannot be nil - it must be a valid schema object or tools will be "Disabled" in Claude Desktop
- The SDK provides `jsonschema.For[T]()` to generate schemas from Go types
- Schema validation happens automatically when the tool is called
- Use jsonschema struct tags for better documentation: `jsonschema:"required,description=..."` 

## Recent Updates

### Schema Implementation (Fixed)
- Refactored all tools to use typed parameter structs
- Added proper jsonschema tags for documentation
- Tools now show as enabled in Claude Desktop with full parameter hints
- Verified schema generation with test script

### RateLimitPolicy Format (Important)
- Rate limits MUST use `limit` and `window` fields
- Window format: time duration strings like "60s", "5m", "1h"
- Do NOT use `duration` and `unit` fields (these will cause errors)
- Window validation ensures format is correct (number + s/m/h)
- Typed structure enforces correct format:
  ```go
  type RateLimit struct {
      Limit  int    `json:"limit"`   // Number of requests
      Window string `json:"window"`  // Time window (e.g., "60s")
  }
  ```
- Example correct format:
  ```yaml
  rates:
  - limit: 100
    window: 60s
  ```

### Testing and Validation

```bash
# Test tool listing and schemas
echo -e '{"jsonrpc":"2.0","method":"initialize","params":{"protocolVersion":"2025-06-18","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}},"id":1}\n{"jsonrpc":"2.0","method":"tools/list","params":{},"id":2}' | ./kuadrant-mcp-server 2>/dev/null | jq '.result.tools[].name'

# Test resource listing
echo -e '{"jsonrpc":"2.0","method":"initialize","params":{"protocolVersion":"2025-06-18","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}},"id":1}\n{"jsonrpc":"2.0","method":"resources/list","params":{},"id":2}' | ./kuadrant-mcp-server 2>/dev/null | jq '.result.resources[].uri' | sort

# Test reading a specific resource
echo -e '{"jsonrpc":"2.0","method":"initialize","params":{"protocolVersion":"2025-06-18","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}},"id":1}\n{"jsonrpc":"2.0","method":"resources/read","params":{"uri":"kuadrant://docs/authpolicy"},"id":2}' | ./kuadrant-mcp-server 2>/dev/null | jq -r '.result.contents[0].text' | head -20

# Test tool invocation (create a gateway)
echo -e '{"jsonrpc":"2.0","method":"initialize","params":{"protocolVersion":"2025-06-18","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}},"id":1}\n{"jsonrpc":"2.0","method":"tools/call","params":{"name":"create_gateway","arguments":{"name":"test-gw","namespace":"default","gatewayClassName":"istio","listeners":[{"name":"http","port":80,"protocol":"HTTP"}]}},"id":2}' | ./kuadrant-mcp-server 2>/dev/null | jq -r '.result.content[0].text'
```

## Current MCP Resources

The server provides 13 documentation resources, fetched from upstream GitHub repos:

**Policy References:**
- `kuadrant://docs/gateway-api` - Gateway API overview
- `kuadrant://docs/dnspolicy` - DNSPolicy reference
- `kuadrant://docs/ratelimitpolicy` - RateLimitPolicy reference
- `kuadrant://docs/tokenratelimitpolicy` - TokenRateLimitPolicy reference
- `kuadrant://docs/authpolicy` - AuthPolicy reference
- `kuadrant://docs/tlspolicy` - TLSPolicy reference
- `kuadrant://docs/telemetrypolicy` - TelemetryPolicy reference
- `kuadrant://docs/kuadrant` - Kuadrant CR reference
- `kuadrant://docs/authorino-features` - Authorino authentication/authorization features

**Extensions:**
- `kuadrant://docs/planpolicy` - PlanPolicy extension (plan-based rate limiting)

**User Guides:**
- `kuadrant://docs/secure-protect-connect` - Full walkthrough
- `kuadrant://docs/simple-ratelimiting` - Simple rate limiting guide
- `kuadrant://docs/auth-for-developers` - Auth for app devs and platform engineers

Resources are fetched from raw.githubusercontent.com and cached for 15 minutes.

## Potential Enhancements

### Additional Resources
- User guides from extracted docs (rate limiting, auth, DNS, TLS walkthroughs)
- Integration examples with specific frameworks (Istio, Envoy Gateway)
- Performance tuning guides
- Migration guides from other API management solutions

### Tool Improvements
- Support for more complex policy configurations
- Policy validation before generation
- Policy composition helpers (combining multiple policies)
- Import existing resources and convert to Kuadrant policies

## Docker Image

Published automatically via GitHub Actions to:
- `ghcr.io/jasonmadigan/kuadrant-mcp-server:latest`

Users can run directly without building:
```bash
docker run -i --rm ghcr.io/jasonmadigan/kuadrant-mcp-server:latest
```

## Resource Documentation Architecture

Documentation is fetched at runtime from upstream GitHub repos (raw.githubusercontent.com) and cached in memory for 15 minutes.

### How It Works

1. When a resource is requested, the server fetches from the raw GitHub URL
2. Content is cached in memory with a 15-minute TTL
3. If fetch fails (network issues, etc.), a fallback with a link to docs.kuadrant.io is returned
4. No local files to maintain - docs.kuadrant.io is the single source of truth

### Adding New Resources

Add an entry to `resourceMapping` in `resources.go`:

```go
"kuadrant://docs/newpolicy": {
    url:         "https://raw.githubusercontent.com/Kuadrant/kuadrant-operator/main/doc/reference/newpolicy.md",
    name:        "NewPolicy Reference",
    description: "Description of the new policy",
    fallback:    "# NewPolicy\n\nSee: https://docs.kuadrant.io/latest/kuadrant-operator/doc/reference/newpolicy/",
},
```

### Documentation Sources

Resources are fetched from:
- **kuadrant-operator** - Policy references and user guides
- **authorino** - Authentication/authorization features