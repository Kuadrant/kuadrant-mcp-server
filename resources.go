package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// docSource defines where to fetch a document from
type docSource struct {
	url         string // raw GitHub URL
	name        string
	description string
	fallback    string // minimal fallback content if fetch fails
}

// resourceMapping maps URI paths to document sources
var resourceMapping = map[string]docSource{
	"kuadrant://docs/gateway-api": {
		url:         "https://raw.githubusercontent.com/Kuadrant/kuadrant-operator/main/doc/overviews/gateway-api.md",
		name:        "Gateway API Overview",
		description: "Overview of Gateway API and Kuadrant integration",
		fallback:    "# Gateway API\n\nSee: https://docs.kuadrant.io/latest/kuadrant-operator/doc/overviews/gateway-api/",
	},
	"kuadrant://docs/dnspolicy": {
		url:         "https://raw.githubusercontent.com/Kuadrant/kuadrant-operator/main/doc/reference/dnspolicy.md",
		name:        "DNSPolicy Reference",
		description: "Complete DNSPolicy specification and examples",
		fallback:    "# DNSPolicy\n\nSee: https://docs.kuadrant.io/latest/kuadrant-operator/doc/reference/dnspolicy/",
	},
	"kuadrant://docs/ratelimitpolicy": {
		url:         "https://raw.githubusercontent.com/Kuadrant/kuadrant-operator/main/doc/reference/ratelimitpolicy.md",
		name:        "RateLimitPolicy Reference",
		description: "Complete RateLimitPolicy specification and examples",
		fallback:    "# RateLimitPolicy\n\nSee: https://docs.kuadrant.io/latest/kuadrant-operator/doc/reference/ratelimitpolicy/",
	},
	"kuadrant://docs/authpolicy": {
		url:         "https://raw.githubusercontent.com/Kuadrant/kuadrant-operator/main/doc/reference/authpolicy.md",
		name:        "AuthPolicy Reference",
		description: "Complete AuthPolicy specification and examples",
		fallback:    "# AuthPolicy\n\nSee: https://docs.kuadrant.io/latest/kuadrant-operator/doc/reference/authpolicy/",
	},
	"kuadrant://docs/tlspolicy": {
		url:         "https://raw.githubusercontent.com/Kuadrant/kuadrant-operator/main/doc/reference/tlspolicy.md",
		name:        "TLSPolicy Reference",
		description: "Complete TLSPolicy specification and examples",
		fallback:    "# TLSPolicy\n\nSee: https://docs.kuadrant.io/latest/kuadrant-operator/doc/reference/tlspolicy/",
	},
	"kuadrant://docs/tokenratelimitpolicy": {
		url:         "https://raw.githubusercontent.com/Kuadrant/kuadrant-operator/main/doc/reference/tokenratelimitpolicy.md",
		name:        "TokenRateLimitPolicy Reference",
		description: "Token-based rate limiting for AI/LLM services",
		fallback:    "# TokenRateLimitPolicy\n\nSee: https://docs.kuadrant.io/latest/kuadrant-operator/doc/reference/tokenratelimitpolicy/",
	},
	"kuadrant://docs/kuadrant": {
		url:         "https://raw.githubusercontent.com/Kuadrant/kuadrant-operator/main/doc/reference/kuadrant.md",
		name:        "Kuadrant CR Reference",
		description: "Main Kuadrant custom resource configuration",
		fallback:    "# Kuadrant CR\n\nSee: https://docs.kuadrant.io/latest/kuadrant-operator/doc/reference/kuadrant/",
	},
	"kuadrant://docs/authorino-features": {
		url:         "https://raw.githubusercontent.com/Kuadrant/authorino/main/docs/features.md",
		name:        "Authorino Features",
		description: "Complete guide to Authorino authentication and authorization features",
		fallback:    "# Authorino Features\n\nSee: https://docs.kuadrant.io/latest/authorino/docs/features/",
	},
	"kuadrant://docs/telemetrypolicy": {
		url:         "https://raw.githubusercontent.com/Kuadrant/kuadrant-operator/main/doc/reference/telemetrypolicy.md",
		name:        "TelemetryPolicy Reference",
		description: "Custom metrics labels for Gateway API resources",
		fallback:    "# TelemetryPolicy\n\nSee: https://docs.kuadrant.io/latest/kuadrant-operator/doc/reference/telemetrypolicy/",
	},
	"kuadrant://docs/planpolicy": {
		url:         "https://raw.githubusercontent.com/Kuadrant/kuadrant-operator/main/doc/extensions/planpolicy.md",
		name:        "PlanPolicy Extension",
		description: "Plan-based rate limiting for tiered service offerings",
		fallback:    "# PlanPolicy\n\nSee: https://docs.kuadrant.io/latest/kuadrant-operator/doc/extensions/planpolicy/",
	},
	"kuadrant://docs/secure-protect-connect": {
		url:         "https://raw.githubusercontent.com/Kuadrant/kuadrant-operator/main/doc/user-guides/full-walkthrough/secure-protect-connect.md",
		name:        "Secure, Protect and Connect",
		description: "Full walkthrough: securing, protecting and connecting services with Kuadrant",
		fallback:    "# Secure, Protect and Connect\n\nSee: https://docs.kuadrant.io/latest/kuadrant-operator/doc/user-guides/full-walkthrough/secure-protect-connect/",
	},
	"kuadrant://docs/simple-ratelimiting": {
		url:         "https://raw.githubusercontent.com/Kuadrant/kuadrant-operator/main/doc/user-guides/ratelimiting/simple-rl-for-app-developers.md",
		name:        "Simple Rate Limiting Guide",
		description: "Getting started with rate limiting for application developers",
		fallback:    "# Simple Rate Limiting\n\nSee: https://docs.kuadrant.io/latest/kuadrant-operator/doc/user-guides/ratelimiting/simple-rl-for-app-developers/",
	},
	"kuadrant://docs/auth-for-developers": {
		url:         "https://raw.githubusercontent.com/Kuadrant/kuadrant-operator/main/doc/user-guides/auth/auth-for-app-devs-and-platform-engineers.md",
		name:        "Auth for Developers",
		description: "Authentication and authorization guide for app developers and platform engineers",
		fallback:    "# Auth for Developers\n\nSee: https://docs.kuadrant.io/latest/kuadrant-operator/doc/user-guides/auth/auth-for-app-devs-and-platform-engineers/",
	},
}

// docCache stores fetched documents with TTL
type docCache struct {
	mu     sync.RWMutex
	docs   map[string]cachedDoc
	ttl    time.Duration
	client *http.Client
}

type cachedDoc struct {
	content   string
	fetchedAt time.Time
}

var cache = &docCache{
	docs: make(map[string]cachedDoc),
	ttl:  15 * time.Minute,
	client: &http.Client{
		Timeout: 30 * time.Second,
	},
}

// fetch retrieves a document, using cache if available and fresh
func (c *docCache) fetch(ctx context.Context, url, fallback string) (string, error) {
	c.mu.RLock()
	if doc, ok := c.docs[url]; ok && time.Since(doc.fetchedAt) < c.ttl {
		c.mu.RUnlock()
		return doc.content, nil
	}
	c.mu.RUnlock()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return c.fallbackOrError(url, fallback, err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return c.fallbackOrError(url, fallback, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return c.fallbackOrError(url, fallback, fmt.Errorf("HTTP %d", resp.StatusCode))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return c.fallbackOrError(url, fallback, err)
	}

	content := string(body)

	c.mu.Lock()
	c.docs[url] = cachedDoc{
		content:   content,
		fetchedAt: time.Now(),
	}
	c.mu.Unlock()

	return content, nil
}

func (c *docCache) fallbackOrError(url, fallback string, err error) (string, error) {
	if fallback != "" {
		log.Printf("[KUADRANT MCP] Failed to fetch %s: %v, using fallback", url, err)
		return fallback, nil
	}
	return "", fmt.Errorf("failed to fetch %s: %w", url, err)
}

// createResourceHandler creates a handler that fetches from upstream
func createResourceHandler(source docSource) func(context.Context, *mcp.ServerSession, *mcp.ReadResourceParams) (*mcp.ReadResourceResult, error) {
	return func(ctx context.Context, ss *mcp.ServerSession, params *mcp.ReadResourceParams) (*mcp.ReadResourceResult, error) {
		log.Printf("[KUADRANT MCP] Resource requested: %s", params.URI)

		content, err := cache.fetch(ctx, source.url, source.fallback)
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

// addKuadrantResources registers all MCP resources
func addKuadrantResources(server *mcp.Server) {
	var resources []*mcp.ServerResource

	for uri, source := range resourceMapping {
		src := source // capture for closure

		resources = append(resources, &mcp.ServerResource{
			Resource: &mcp.Resource{
				URI:         uri,
				Name:        src.name,
				Description: src.description,
				MIMEType:    "text/markdown",
			},
			Handler: createResourceHandler(src),
		})
	}

	server.AddResources(resources...)
}
