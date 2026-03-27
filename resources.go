package main

import (
	"context"
	"embed"
	"fmt"
	"log"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

//go:embed docs/debugging/*.md
var embeddedDocs embed.FS

// resourceDef defines an embedded resource
type resourceDef struct {
	path        string // path within embed.FS
	name        string
	description string
}

// debugResourceMapping maps URIs to embedded debugging docs
var debugResourceMapping = map[string]resourceDef{
	"kuadrant://debug/authpolicy": {
		path:        "docs/debugging/authpolicy.md",
		name:        "AuthPolicy Debugging Guide",
		description: "Diagnose and fix AuthPolicy issues including target not found, policy not enforced, and Authorino problems",
	},
	"kuadrant://debug/status-conditions": {
		path:        "docs/debugging/status-conditions.md",
		name:        "Status Conditions Reference",
		description: "Comprehensive guide to interpreting Kuadrant policy status conditions (Accepted, Enforced, etc.)",
	},
	// TODO: Uncomment when debugging docs are created
	// "kuadrant://debug/installation": {
	// 	path:        "docs/debugging/installation.md",
	// 	name:        "Installation Debugging Guide",
	// 	description: "Diagnose Kuadrant operator, CRDs, Istio, Limitador, and Authorino installation issues",
	// },
	// "kuadrant://debug/gateway-istio": {
	// 	path:        "docs/debugging/gateway-istio.md",
	// 	name:        "Gateway/Istio Debugging Guide",
	// 	description: "Diagnose Gateway listener issues and Istio proxy configuration",
	// },
	// "kuadrant://debug/dnspolicy": {
	// 	path:        "docs/debugging/dnspolicy.md",
	// 	name:        "DNSPolicy Debugging Guide",
	// 	description: "Diagnose DNS record creation, provider configuration, and zone issues",
	// },
	// "kuadrant://debug/tlspolicy": {
	// 	path:        "docs/debugging/tlspolicy.md",
	// 	name:        "TLSPolicy Debugging Guide",
	// 	description: "Diagnose certificate issuance, cert-manager, and issuer problems",
	// },
	// "kuadrant://debug/ratelimitpolicy": {
	// 	path:        "docs/debugging/ratelimitpolicy.md",
	// 	name:        "RateLimitPolicy Debugging Guide",
	// 	description: "Diagnose rate limit enforcement and Limitador health issues",
	// },
	// "kuadrant://debug/telemetrypolicy": {
	// 	path:        "docs/debugging/telemetrypolicy.md",
	// 	name:        "TelemetryPolicy Debugging Guide",
	// 	description: "Diagnose custom metrics and CEL expression issues",
	// },
	// "kuadrant://debug/tokenratelimitpolicy": {
	// 	path:        "docs/debugging/tokenratelimitpolicy.md",
	// 	name:        "TokenRateLimitPolicy Debugging Guide",
	// 	description: "Diagnose token-based rate limiting for AI/LLM services",
	// },
	// "kuadrant://debug/policy-conflicts": {
	// 	path:        "docs/debugging/policy-conflicts.md",
	// 	name:        "Policy Conflicts Debugging Guide",
	// 	description: "Diagnose override/default conflicts and policy hierarchy issues",
	// },
}

// createEmbeddedResourceHandler creates a handler that serves from embedded FS
func createEmbeddedResourceHandler(def resourceDef) func(context.Context, *mcp.ServerSession, *mcp.ReadResourceParams) (*mcp.ReadResourceResult, error) {
	return func(ctx context.Context, ss *mcp.ServerSession, params *mcp.ReadResourceParams) (*mcp.ReadResourceResult, error) {
		log.Printf("[KUADRANT MCP] Resource requested: %s", params.URI)

		// Read from embedded filesystem
		content, err := embeddedDocs.ReadFile(def.path)
		if err != nil {
			return nil, fmt.Errorf("failed to read embedded resource %s: %w", def.path, err)
		}

		return &mcp.ReadResourceResult{
			Contents: []*mcp.ResourceContents{
				{
					URI:      params.URI,
					MIMEType: "text/markdown",
					Text:     string(content),
				},
			},
		}, nil
	}
}

// addDebugResources registers all debugging resources with the server
func addDebugResources(server *mcp.Server) {
	var resources []*mcp.ServerResource

	// Add embedded debugging resources (replacing old HTTP-fetched docs)
	for uri, def := range debugResourceMapping {
		resDef := def // capture for closure

		resources = append(resources, &mcp.ServerResource{
			Resource: &mcp.Resource{
				URI:         uri,
				Name:        resDef.name,
				Description: resDef.description,
				MIMEType:    "text/markdown",
			},
			Handler: createEmbeddedResourceHandler(resDef),
		})
	}

	server.AddResources(resources...)
	log.Printf("[KUADRANT MCP] Registered %d debugging resources", len(resources))
}
