package main

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func addDebugPromptsSimple(server *mcp.Server) {
	server.AddPrompts(
		// NOTE: We include simplified debugging guidance inline in the prompt response
		// rather than only referencing the full resource guide. This approach works better
		// with smaller LLMs (e.g., llama3.2:1b) which struggle to:
		// - Follow complex multi-step instructions in long documents
		// - Avoid hallucinating when presented with multiple examples
		// - Respect "DO NOT" directives buried in detailed guides
		// The inline prompt provides immediate, focused guidance for the most common issues,
		// with a reference to the full guide for additional details.
		&mcp.ServerPrompt{
			Prompt: &mcp.Prompt{
				Name:        "debug-authpolicy",
				Description: "Debug AuthPolicy TargetNotFound or enforcement issues",
				Arguments: []*mcp.PromptArgument{
					{Name: "policy-name", Required: true, Description: "AuthPolicy name"},
					{Name: "namespace", Required: false, Description: "Namespace"},
				},
			},
			Handler: func(ctx context.Context, _ *mcp.ServerSession, params *mcp.GetPromptParams) (*mcp.GetPromptResult, error) {
				policyName := params.Arguments["policy-name"]
				namespace := params.Arguments["namespace"]

				nsInfo := ""
				if namespace != "" {
					nsInfo = fmt.Sprintf(" in namespace %s", namespace)
				}

				result := fmt.Sprintf(`# Debugging AuthPolicy "%s"%s

First, use the kubernetes MCP server to get the AuthPolicy and read status.conditions[].message.

## If message says "was not found" or "target ... was not found"

The Gateway/HTTPRoute does NOT exist in the cluster. The targetRef is correct - don't change it.

**Fix:**
1. Check if Gateway exists: kubectl get gateway <target-name> -n <namespace>
2. If NotFound, create the Gateway (or fix the name if it's a typo)
3. The targetRef group/kind fields are already correct - DO NOT modify them

## If message says "spec.targetRef: Required value"

The AuthPolicy is missing targetRef entirely. Add:
` + "```yaml" + `
spec:
  targetRef:
    group: gateway.networking.k8s.io
    kind: Gateway
    name: <gateway-name>
` + "```" + `

## If status shows Accepted=True but Enforced=False

Authorino may not be running. Check: kubectl get pods -n kuadrant-system -l app=authorino

---

**Full debugging guide:** kuadrant://debug/authpolicy`, policyName, nsInfo)

				return &mcp.GetPromptResult{
					Messages: []*mcp.PromptMessage{
						{
							Role:    "user",
							Content: &mcp.TextContent{Text: result},
						},
					},
				}, nil
			},
		},
	)
}
