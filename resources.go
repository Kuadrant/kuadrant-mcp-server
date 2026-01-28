package main

import (
	"context"
	"embed"
	"fmt"
	"log"
	"path"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

//go:embed docs/*.md docs/examples/*.md
var docsFS embed.FS

// resourceMapping maps URI paths to file paths in the embedded filesystem
var resourceMapping = map[string]struct {
	file        string
	name        string
	description string
}{
	"kuadrant://docs/gateway-api":         {"docs/gateway-api.md", "Gateway API Overview", "Overview of Gateway API and Kuadrant integration"},
	"kuadrant://docs/dnspolicy":           {"docs/dnspolicy.md", "DNSPolicy Reference", "Complete DNSPolicy specification and examples"},
	"kuadrant://docs/ratelimitpolicy":     {"docs/ratelimitpolicy.md", "RateLimitPolicy Reference", "Complete RateLimitPolicy specification and examples"},
	"kuadrant://docs/authpolicy":          {"docs/authpolicy.md", "AuthPolicy Reference", "Complete AuthPolicy specification and examples"},
	"kuadrant://docs/tlspolicy":           {"docs/tlspolicy.md", "TLSPolicy Reference", "Complete TLSPolicy specification and examples"},
	"kuadrant://docs/tokenratelimitpolicy": {"docs/tokenratelimitpolicy.md", "TokenRateLimitPolicy Reference", "Token-based rate limiting for AI/LLM services"},
	"kuadrant://docs/kuadrant":            {"docs/kuadrant.md", "Kuadrant CR Reference", "Main Kuadrant custom resource configuration"},
	"kuadrant://docs/authorino-features":  {"docs/authorino-features.md", "Authorino Features", "Complete guide to Authorino authentication and authorization features"},
	"kuadrant://docs/telemetrypolicy":     {"docs/telemetrypolicy.md", "TelemetryPolicy Reference", "Custom metrics labels for Gateway API resources"},
	"kuadrant://docs/planpolicy":          {"docs/planpolicy.md", "PlanPolicy Extension", "Plan-based rate limiting for tiered service offerings"},
	"kuadrant://examples/basic-setup":     {"docs/examples/basic-setup.md", "Basic API Setup Example", "Complete example of basic API with rate limiting and auth"},
	"kuadrant://examples/production-setup": {"docs/examples/production-setup.md", "Production API Setup Example", "Full production setup with TLS, DNS, and advanced policies"},
	"kuadrant://troubleshooting":          {"docs/troubleshooting.md", "Troubleshooting Guide", "Common issues and debugging techniques for Kuadrant"},
}

// readEmbeddedDoc reads a document from the embedded filesystem
func readEmbeddedDoc(filePath string) (string, error) {
	content, err := docsFS.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read embedded doc %s: %w", filePath, err)
	}
	return string(content), nil
}

// createResourceHandler creates a handler for an embedded document
func createResourceHandler(filePath string) func(context.Context, *mcp.ServerSession, *mcp.ReadResourceParams) (*mcp.ReadResourceResult, error) {
	return func(ctx context.Context, ss *mcp.ServerSession, params *mcp.ReadResourceParams) (*mcp.ReadResourceResult, error) {
		log.Printf("[KUADRANT MCP] Resource requested: %s", params.URI)

		content, err := readEmbeddedDoc(filePath)
		if err != nil {
			return nil, err
		}

		return &mcp.ReadResourceResult{
			Contents: []*mcp.ResourceContents{
				{
					URI:      params.URI,
					MIMEType: "text/markdown",
					Text:     content,
				},
			},
		}, nil
	}
}

// addKuadrantResources adds all MCP resources from embedded docs
func addKuadrantResources(server *mcp.Server) {
	var resources []*mcp.ServerResource

	for uri, info := range resourceMapping {
		// capture for closure
		filePath := info.file

		resources = append(resources, &mcp.ServerResource{
			Resource: &mcp.Resource{
				URI:         uri,
				Name:        info.name,
				Description: info.description,
				MIMEType:    "text/markdown",
			},
			Handler: createResourceHandler(filePath),
		})
	}

	server.AddResources(resources...)
}

// listEmbeddedDocs returns a list of all embedded documentation files (for debugging)
func listEmbeddedDocs() ([]string, error) {
	var files []string

	entries, err := docsFS.ReadDir("docs")
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			// read subdirectory
			subEntries, err := docsFS.ReadDir(path.Join("docs", entry.Name()))
			if err != nil {
				continue
			}
			for _, subEntry := range subEntries {
				if !subEntry.IsDir() && strings.HasSuffix(subEntry.Name(), ".md") {
					files = append(files, path.Join("docs", entry.Name(), subEntry.Name()))
				}
			}
		} else if strings.HasSuffix(entry.Name(), ".md") {
			files = append(files, path.Join("docs", entry.Name()))
		}
	}

	return files, nil
}
