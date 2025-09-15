// +build ignore

package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
)

// Template for generating resource handler functions
const resourceHandlerTemplate = `
func {{.FuncName}}ResourceHandler(ctx context.Context, ss *mcp.ServerSession, params *mcp.ReadResourceParams) (*mcp.ReadResourceResult, error) {
	log.Printf("[KUADRANT MCP] Resource requested: %s", params.URI)
	content := ` + "`" + `{{.Content}}` + "`" + `

	return &mcp.ReadResourceResult{
		Contents: []mcp.ResourceContent{
			{
				Text: content,
				URI:  params.URI,
			},
		},
	}, nil
}
`

type ResourceDoc struct {
	Name     string
	FuncName string
	Content  string
}

func main() {
	docsDir := "./extracted-docs"

	if _, err := os.Stat(docsDir); os.IsNotExist(err) {
		fmt.Println("❌ No extracted docs found. Run ./update-docs.sh first")
		os.Exit(1)
	}

	resources := []ResourceDoc{}

	// Process reference documentation with new directory structure
	refFiles := map[string]string{
		"authpolicy":           "authPolicy",
		"dnspolicy":            "dnsPolicy",
		"ratelimitpolicy":      "rateLimitPolicy",
		"tlspolicy":            "tlsPolicy",
		"tokenratelimitpolicy": "tokenRateLimitPolicy",
		"kuadrant":             "kuadrant",
	}

	for filename, funcName := range refFiles {
		refPath := filepath.Join(docsDir, "kuadrant-operator/doc/reference", filename+".md")
		if content, err := processMarkdownFile(refPath); err == nil {
			resources = append(resources, ResourceDoc{
				Name:     filename,
				FuncName: funcName,
				Content:  content,
			})
			fmt.Printf("✓ Processed %s reference\n", filename)
		} else {
			fmt.Printf("⚠ Could not process %s: %v\n", filename, err)
		}
	}

	// Also process Authorino features
	authorinoPath := filepath.Join(docsDir, "authorino/docs/features.md")
	if content, err := processMarkdownFile(authorinoPath); err == nil {
		resources = append(resources, ResourceDoc{
			Name:     "authorino-features",
			FuncName: "authorinoFeatures",
			Content:  content,
		})
		fmt.Printf("✓ Processed Authorino features\n")
	}

	// Generate Go code
	generateResourcesFile(resources)

	fmt.Println("\n✅ Generated resources-generated.go")
	fmt.Println("Review the file and integrate it into resources.go")
}

func processMarkdownFile(path string) (string, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}

	// Convert markdown to a format suitable for Go string literals
	processed := string(content)

	// Extract key sections (API spec, examples, important notes)
	processed = extractKeyContent(processed)

	// Clean up for Go string literal
	processed = cleanForGoString(processed)

	return processed, nil
}

func extractKeyContent(content string) string {
	// This is a simplified extraction - you can enhance it based on needs

	// Remove excessive whitespace
	content = regexp.MustCompile(`\n{3,}`).ReplaceAllString(content, "\n\n")

	// Remove HTML comments
	content = regexp.MustCompile(`<!--.*?-->`).ReplaceAllString(content, "")

	// Remove navigation/metadata sections if present
	content = regexp.MustCompile(`(?i)^#+\s*(navigation|metadata|table of contents).*?\n`).ReplaceAllString(content, "")

	// IMPORTANT: Translate AuthConfig to AuthPolicy for Kuadrant context
	// AuthConfig is used in standalone Authorino, AuthPolicy is the Kuadrant equivalent
	content = strings.ReplaceAll(content, "AuthConfig", "AuthPolicy")
	content = strings.ReplaceAll(content, "authconfig", "authpolicy")
	content = strings.ReplaceAll(content, "authconfigs", "authpolicies")

	// Fix any doubled replacements
	content = strings.ReplaceAll(content, "AuthPolicyConfig", "AuthPolicy")

	// Limit to reasonable size (first 10KB of content)
	if len(content) > 10000 {
		content = content[:10000] + "\n\n[Content truncated for space - see full docs at docs.kuadrant.io]"
	}

	return content
}

func cleanForGoString(content string) string {
	// Escape backticks for Go string literals
	// We'll use a placeholder for code blocks

	// First, protect code blocks by replacing them with placeholders
	codeBlocks := []string{}
	re := regexp.MustCompile("(?s)```.*?```")
	content = re.ReplaceAllStringFunc(content, func(match string) string {
		idx := len(codeBlocks)
		codeBlocks = append(codeBlocks, match)
		return fmt.Sprintf("__CODE_BLOCK_%d__", idx)
	})

	// Now we can safely work with the content
	// ... additional processing if needed ...

	// Restore code blocks with proper escaping
	for i, block := range codeBlocks {
		placeholder := fmt.Sprintf("__CODE_BLOCK_%d__", i)
		// For Go string literals, we need to handle backticks specially
		escapedBlock := strings.ReplaceAll(block, "```", "` + \"```\" + `")
		content = strings.ReplaceAll(content, placeholder, escapedBlock)
	}

	return content
}

func generateResourcesFile(resources []ResourceDoc) {
	tmpl, err := template.New("resources").Parse(resourceHandlerTemplate)
	if err != nil {
		panic(err)
	}

	file, err := os.Create("resources-generated.go")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	// Write file header
	file.WriteString(`// Code generated by process-docs.go. DO NOT EDIT.
// This file contains resource handlers generated from Kuadrant documentation.
// Review and integrate into resources.go as needed.

package main

import (
	"context"
	"log"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)
`)

	// Generate handler functions
	for _, resource := range resources {
		err := tmpl.Execute(file, resource)
		if err != nil {
			fmt.Printf("Error generating %s: %v\n", resource.Name, err)
		}
	}

	// Generate registration function
	file.WriteString(`
// Add this to your addKuadrantResources function:
func registerGeneratedResources(server *mcp.Server) {`)

	for _, resource := range resources {
		file.WriteString(fmt.Sprintf(`
	server.AddResource(mcp.NewResource(
		"kuadrant://docs/%s",
		"%s Reference Documentation",
		"text/plain",
		%sResourceHandler,
	))`, resource.Name, strings.Title(resource.Name), resource.FuncName))
	}

	file.WriteString(`
}
`)
}