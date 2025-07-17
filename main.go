package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"gopkg.in/yaml.v3"
)

// Input parameter types for tools
type CreateGatewayParams struct {
	Name             string                   `json:"name" jsonschema:"required,description=Name of the Gateway resource"`
	Namespace        string                   `json:"namespace" jsonschema:"required,description=Kubernetes namespace for the Gateway"`
	GatewayClassName string                   `json:"gatewayClassName,omitempty" jsonschema:"description=Gateway implementation to use (default: istio)"`
	Listeners        []map[string]interface{} `json:"listeners,omitempty" jsonschema:"description=Gateway listeners configuration"`
	KuadrantEnabled  bool                     `json:"kuadrantEnabled,omitempty" jsonschema:"description=Enable Kuadrant policy attachment (default: true)"`
}

type CreateHTTPRouteParams struct {
	Name       string                   `json:"name" jsonschema:"required,description=Name of the HTTPRoute resource"`
	Namespace  string                   `json:"namespace" jsonschema:"required,description=Kubernetes namespace for the HTTPRoute"`
	ParentRefs []interface{}            `json:"parentRefs" jsonschema:"required,description=References to Gateway resources"`
	Hostnames  []interface{}            `json:"hostnames,omitempty" jsonschema:"description=Hostnames this route handles"`
	Rules      []interface{}            `json:"rules,omitempty" jsonschema:"description=Routing rules configuration"`
}

type CreateDNSPolicyParams struct {
	Name           string                 `json:"name" jsonschema:"required,description=Name of the DNSPolicy resource"`
	Namespace      string                 `json:"namespace" jsonschema:"required,description=Kubernetes namespace for the DNSPolicy"`
	TargetRef      map[string]interface{} `json:"targetRef" jsonschema:"required,description=Reference to the target Gateway"`
	ProviderRefs   []interface{}          `json:"providerRefs,omitempty" jsonschema:"description=DNS provider configurations"`
	ProviderRef    map[string]interface{} `json:"providerRef,omitempty" jsonschema:"description=Legacy single provider reference"`
	LoadBalancing  map[string]interface{} `json:"loadBalancing,omitempty" jsonschema:"description=Load balancing configuration"`
	HealthCheck    map[string]interface{} `json:"healthCheck,omitempty" jsonschema:"description=Health check configuration"`
}

type CreateTLSPolicyParams struct {
	Name        string                 `json:"name" jsonschema:"required,description=Name of the TLSPolicy resource"`
	Namespace   string                 `json:"namespace" jsonschema:"required,description=Kubernetes namespace for the TLSPolicy"`
	TargetRef   map[string]interface{} `json:"targetRef" jsonschema:"required,description=Reference to the target Gateway"`
	IssuerRef   map[string]interface{} `json:"issuerRef" jsonschema:"required,description=Reference to the cert-manager issuer"`
	CommonName  string                 `json:"commonName,omitempty" jsonschema:"description=Common name for the certificate"`
	Duration    string                 `json:"duration,omitempty" jsonschema:"description=Certificate duration (e.g. 90d)"`
	RenewBefore string                 `json:"renewBefore,omitempty" jsonschema:"description=When to renew before expiry (e.g. 30d)"`
}

// RateLimit represents a single rate limit configuration
type RateLimit struct {
	Limit  int    `json:"limit" jsonschema:"required,description=Number of requests allowed"`
	Window string `json:"window" jsonschema:"required,description=Time window (e.g. 10s, 5m, 1h)"`
}

// LimitDefinition represents a named limit with rates and optional conditions
type LimitDefinition struct {
	Rates []RateLimit              `json:"rates" jsonschema:"required,description=Array of rate limit rules"`
	When  []map[string]interface{} `json:"when,omitempty" jsonschema:"description=Optional conditions for applying this limit"`
}

type CreateRateLimitPolicyParams struct {
	Name      string                         `json:"name" jsonschema:"required,description=Name of the RateLimitPolicy resource"`
	Namespace string                         `json:"namespace" jsonschema:"required,description=Kubernetes namespace for the RateLimitPolicy"`
	TargetRef map[string]interface{}         `json:"targetRef" jsonschema:"required,description=Reference to the target Gateway or HTTPRoute"`
	Limits    map[string]LimitDefinition     `json:"limits,omitempty" jsonschema:"description=Named rate limit configurations"`
	Defaults  map[string]interface{}         `json:"defaults,omitempty" jsonschema:"description=Default rate limit rules"`
	Overrides map[string]interface{}         `json:"overrides,omitempty" jsonschema:"description=Override rate limit rules"`
}

type CreateAuthPolicyParams struct {
	Name      string                 `json:"name" jsonschema:"required,description=Name of the AuthPolicy resource"`
	Namespace string                 `json:"namespace" jsonschema:"required,description=Kubernetes namespace for the AuthPolicy"`
	TargetRef map[string]interface{} `json:"targetRef" jsonschema:"required,description=Reference to the target Gateway or HTTPRoute"`
	Rules     map[string]interface{} `json:"rules,omitempty" jsonschema:"description=Authentication and authorization rules"`
	Defaults  map[string]interface{} `json:"defaults,omitempty" jsonschema:"description=Default auth rules"`
	Overrides map[string]interface{} `json:"overrides,omitempty" jsonschema:"description=Override auth rules"`
}

// Tool handlers
func createGatewayHandler(ctx context.Context, params CreateGatewayParams) (string, error) {
	name := params.Name
	namespace := params.Namespace
	
	log.Printf("[KUADRANT MCP] create_gateway called with name=%s, namespace=%s", name, namespace)
	if name == "" || namespace == "" {
		return "Error: name and namespace are required", nil
	}

	gatewayClassName := params.GatewayClassName
	if gatewayClassName == "" {
		gatewayClassName = "istio"
	}

	listeners := params.Listeners
	if len(listeners) == 0 {
		listeners = []map[string]interface{}{
			{
				"name":     "http",
				"port":     80,
				"protocol": "HTTP",
			},
		}
	}

	kuadrantEnabled := params.KuadrantEnabled
	if params.KuadrantEnabled == false {
		kuadrantEnabled = false
	} else {
		kuadrantEnabled = true
	}

	gateway := map[string]interface{}{
		"apiVersion": "gateway.networking.k8s.io/v1",
		"kind":       "Gateway",
		"metadata": map[string]interface{}{
			"name":      name,
			"namespace": namespace,
		},
		"spec": map[string]interface{}{
			"gatewayClassName": gatewayClassName,
			"listeners":        listeners,
		},
	}

	if kuadrantEnabled {
		metadata := gateway["metadata"].(map[string]interface{})
		metadata["annotations"] = map[string]string{
			"kuadrant.io/policy": "enabled",
		}
	}

	content, err := yaml.Marshal(gateway)
	if err != nil {
		return fmt.Sprintf("Error: Failed to generate YAML: %v", err), nil
	}

	return string(content), nil
}

func createHTTPRouteHandler(ctx context.Context, params CreateHTTPRouteParams) (string, error) {
	name := params.Name
	namespace := params.Namespace
	if name == "" || namespace == "" {
		return "Error: name and namespace are required", nil
	}

	parentRefs := params.ParentRefs
	if len(parentRefs) == 0 {
		return "Error: parentRefs is required", nil
	}

	hostnames := params.Hostnames
	rules := params.Rules

	httproute := map[string]interface{}{
		"apiVersion": "gateway.networking.k8s.io/v1",
		"kind":       "HTTPRoute",
		"metadata": map[string]interface{}{
			"name":      name,
			"namespace": namespace,
		},
		"spec": map[string]interface{}{
			"parentRefs": parentRefs,
		},
	}

	spec := httproute["spec"].(map[string]interface{})
	if len(hostnames) > 0 {
		spec["hostnames"] = hostnames
	}
	if len(rules) > 0 {
		spec["rules"] = rules
	}

	content, err := yaml.Marshal(httproute)
	if err != nil {
		return fmt.Sprintf("Error: Failed to generate YAML: %v", err), nil
	}

	return string(content), nil
}

func createDNSPolicyHandler(ctx context.Context, params CreateDNSPolicyParams) (string, error) {
	name := params.Name
	namespace := params.Namespace
	targetRef := params.TargetRef
	
	if name == "" || namespace == "" || targetRef == nil {
		return "Error: name, namespace, and targetRef are required", nil
	}

	// Ensure targetRef has required fields
	if targetRef["kind"] == nil || targetRef["name"] == nil {
		return "Error: targetRef must have kind and name", nil
	}

	if targetRef["group"] == nil {
		targetRef["group"] = "gateway.networking.k8s.io"
	}

	dnsPolicy := map[string]interface{}{
		"apiVersion": "kuadrant.io/v1",
		"kind":       "DNSPolicy",
		"metadata": map[string]interface{}{
			"name":      name,
			"namespace": namespace,
		},
		"spec": map[string]interface{}{
			"targetRef": targetRef,
		},
	}

	// providerRefs is required
	if len(params.ProviderRefs) > 0 {
		dnsPolicy["spec"].(map[string]interface{})["providerRefs"] = params.ProviderRefs
	} else if params.ProviderRef != nil {
		// Support legacy single providerRef
		dnsPolicy["spec"].(map[string]interface{})["providerRefs"] = []interface{}{params.ProviderRef}
	} else {
		return "Error: providerRefs is required", nil
	}

	// Optional fields
	if params.LoadBalancing != nil {
		dnsPolicy["spec"].(map[string]interface{})["loadBalancing"] = params.LoadBalancing
	}

	if params.HealthCheck != nil {
		dnsPolicy["spec"].(map[string]interface{})["healthCheck"] = params.HealthCheck
	}

	content, err := yaml.Marshal(dnsPolicy)
	if err != nil {
		return fmt.Sprintf("Error: Failed to generate YAML: %v", err), nil
	}

	return string(content), nil
}

func createTLSPolicyHandler(ctx context.Context, params CreateTLSPolicyParams) (string, error) {
	name := params.Name
	namespace := params.Namespace
	targetRef := params.TargetRef
	issuerRef := params.IssuerRef
	
	if name == "" || namespace == "" || targetRef == nil || issuerRef == nil {
		return "Error: name, namespace, targetRef, and issuerRef are required", nil
	}

	// Ensure refs have required fields
	if targetRef["kind"] == nil || targetRef["name"] == nil {
		return "Error: targetRef must have kind and name", nil
	}
	if issuerRef["kind"] == nil || issuerRef["name"] == nil {
		return "Error: issuerRef must have kind and name", nil
	}

	if targetRef["group"] == nil {
		targetRef["group"] = "gateway.networking.k8s.io"
	}
	if issuerRef["group"] == nil {
		issuerRef["group"] = "cert-manager.io"
	}

	tlsPolicy := map[string]interface{}{
		"apiVersion": "kuadrant.io/v1alpha1", // Note: v1alpha1
		"kind":       "TLSPolicy",
		"metadata": map[string]interface{}{
			"name":      name,
			"namespace": namespace,
		},
		"spec": map[string]interface{}{
			"targetRef": targetRef,
			"issuerRef": issuerRef,
		},
	}

	spec := tlsPolicy["spec"].(map[string]interface{})
	if params.CommonName != "" {
		spec["commonName"] = params.CommonName
	}
	if params.Duration != "" {
		spec["duration"] = params.Duration
	}
	if params.RenewBefore != "" {
		spec["renewBefore"] = params.RenewBefore
	}

	content, err := yaml.Marshal(tlsPolicy)
	if err != nil {
		return fmt.Sprintf("Error: Failed to generate YAML: %v", err), nil
	}

	return string(content), nil
}

// Helper function to validate window format
func validateWindow(window string) error {
	// Simple validation - must end with s, m, or h
	if len(window) < 2 {
		return fmt.Errorf("window must be at least 2 characters (e.g., '1s')")
	}
	lastChar := window[len(window)-1]
	if lastChar != 's' && lastChar != 'm' && lastChar != 'h' {
		return fmt.Errorf("window must end with 's' (seconds), 'm' (minutes), or 'h' (hours)")
	}
	// Check if the rest is a number
	numPart := window[:len(window)-1]
	if _, err := strconv.Atoi(numPart); err != nil {
		return fmt.Errorf("window must start with a number (e.g., '60s', '5m', '1h')")
	}
	return nil
}

func createRateLimitPolicyHandler(ctx context.Context, params CreateRateLimitPolicyParams) (string, error) {
	name := params.Name
	namespace := params.Namespace
	targetRef := params.TargetRef
	
	if name == "" || namespace == "" || targetRef == nil {
		return "Error: name, namespace, and targetRef are required", nil
	}

	// Ensure targetRef has required fields
	if targetRef["kind"] == nil || targetRef["name"] == nil {
		return "Error: targetRef must have kind and name", nil
	}

	if targetRef["group"] == nil {
		targetRef["group"] = "gateway.networking.k8s.io"
	}
	
	// Validate rate limit windows
	for limitName, limitDef := range params.Limits {
		for i, rate := range limitDef.Rates {
			if err := validateWindow(rate.Window); err != nil {
				return fmt.Sprintf("Error: Invalid window format in limit '%s' rate[%d]: %v", limitName, i, err), nil
			}
		}
	}

	rateLimitPolicy := map[string]interface{}{
		"apiVersion": "kuadrant.io/v1",
		"kind":       "RateLimitPolicy",
		"metadata": map[string]interface{}{
			"name":      name,
			"namespace": namespace,
		},
		"spec": map[string]interface{}{
			"targetRef": targetRef,
		},
	}

	spec := rateLimitPolicy["spec"].(map[string]interface{})
	
	// Handle limits - convert typed structure to map for YAML marshaling
	if params.Limits != nil && len(params.Limits) > 0 {
		limitsMap := make(map[string]interface{})
		for name, limitDef := range params.Limits {
			limitMap := map[string]interface{}{
				"rates": limitDef.Rates,
			}
			if len(limitDef.When) > 0 {
				limitMap["when"] = limitDef.When
			}
			limitsMap[name] = limitMap
		}
		spec["limits"] = limitsMap
	} else {
		// If no limits provided, add a default global limit
		spec["limits"] = map[string]interface{}{
			"global": map[string]interface{}{
				"rates": []RateLimit{
					{
						Limit:  10,
						Window: "60s",
					},
				},
			},
		}
	}
	
	if params.Defaults != nil && len(params.Defaults) > 0 {
		spec["defaults"] = params.Defaults
	}
	if params.Overrides != nil && len(params.Overrides) > 0 {
		spec["overrides"] = params.Overrides
	}

	content, err := yaml.Marshal(rateLimitPolicy)
	if err != nil {
		return fmt.Sprintf("Error: Failed to generate YAML: %v", err), nil
	}

	return string(content), nil
}

func createAuthPolicyHandler(ctx context.Context, params CreateAuthPolicyParams) (string, error) {
	name := params.Name
	namespace := params.Namespace
	targetRef := params.TargetRef
	
	if name == "" || namespace == "" || targetRef == nil {
		return "Error: name, namespace, and targetRef are required", nil
	}

	// Ensure targetRef has required fields
	if targetRef["kind"] == nil || targetRef["name"] == nil {
		return "Error: targetRef must have kind and name", nil
	}

	if targetRef["group"] == nil {
		targetRef["group"] = "gateway.networking.k8s.io"
	}

	authPolicy := map[string]interface{}{
		"apiVersion": "kuadrant.io/v1",
		"kind":       "AuthPolicy",
		"metadata": map[string]interface{}{
			"name":      name,
			"namespace": namespace,
		},
		"spec": map[string]interface{}{
			"targetRef": targetRef,
		},
	}

	spec := authPolicy["spec"].(map[string]interface{})
	if params.Rules != nil && len(params.Rules) > 0 {
		spec["rules"] = params.Rules
	}
	if params.Defaults != nil && len(params.Defaults) > 0 {
		spec["defaults"] = params.Defaults
	}
	if params.Overrides != nil && len(params.Overrides) > 0 {
		spec["overrides"] = params.Overrides
	}

	content, err := yaml.Marshal(authPolicy)
	if err != nil {
		return fmt.Sprintf("Error: Failed to generate YAML: %v", err), nil
	}

	return string(content), nil
}

func main() {
	// Parse command line flags
	var (
		transport = flag.String("transport", "stdio", "Transport type: stdio, sse, http")
		addr      = flag.String("addr", ":8080", "Address to listen on (for sse/http transports)")
	)
	flag.Parse()

	log.Printf("[KUADRANT MCP] Starting server with transport=%s", *transport)

	// Create server
	server := mcp.NewServer("kuadrant-mcp", "1.0.0", nil)

	// Register tools using NewServerTool with type inference
	server.AddTools(
		mcp.NewServerTool(
			"create_gateway",
			"Generate a Gateway manifest with Kuadrant annotations",
			func(ctx context.Context, _ *mcp.ServerSession, params *mcp.CallToolParamsFor[CreateGatewayParams]) (*mcp.CallToolResultFor[string], error) {
				result, err := createGatewayHandler(ctx, params.Arguments)
				if err != nil {
					return nil, err
				}
				return &mcp.CallToolResultFor[string]{
					Content: []mcp.Content{&mcp.TextContent{Text: result}},
				}, nil
			},
		),
		mcp.NewServerTool(
			"create_httproute",
			"Generate an HTTPRoute manifest",
			func(ctx context.Context, _ *mcp.ServerSession, params *mcp.CallToolParamsFor[CreateHTTPRouteParams]) (*mcp.CallToolResultFor[string], error) {
				result, err := createHTTPRouteHandler(ctx, params.Arguments)
				if err != nil {
					return nil, err
				}
				return &mcp.CallToolResultFor[string]{
					Content: []mcp.Content{&mcp.TextContent{Text: result}},
				}, nil
			},
		),
		mcp.NewServerTool(
			"create_dnspolicy",
			"Generate a Kuadrant DNSPolicy manifest",
			func(ctx context.Context, _ *mcp.ServerSession, params *mcp.CallToolParamsFor[CreateDNSPolicyParams]) (*mcp.CallToolResultFor[string], error) {
				result, err := createDNSPolicyHandler(ctx, params.Arguments)
				if err != nil {
					return nil, err
				}
				return &mcp.CallToolResultFor[string]{
					Content: []mcp.Content{&mcp.TextContent{Text: result}},
				}, nil
			},
		),
		mcp.NewServerTool(
			"create_tlspolicy",
			"Generate a Kuadrant TLSPolicy manifest",
			func(ctx context.Context, _ *mcp.ServerSession, params *mcp.CallToolParamsFor[CreateTLSPolicyParams]) (*mcp.CallToolResultFor[string], error) {
				result, err := createTLSPolicyHandler(ctx, params.Arguments)
				if err != nil {
					return nil, err
				}
				return &mcp.CallToolResultFor[string]{
					Content: []mcp.Content{&mcp.TextContent{Text: result}},
				}, nil
			},
		),
		mcp.NewServerTool(
			"create_ratelimitpolicy",
			"Generate a Kuadrant RateLimitPolicy manifest",
			func(ctx context.Context, _ *mcp.ServerSession, params *mcp.CallToolParamsFor[CreateRateLimitPolicyParams]) (*mcp.CallToolResultFor[string], error) {
				result, err := createRateLimitPolicyHandler(ctx, params.Arguments)
				if err != nil {
					return nil, err
				}
				return &mcp.CallToolResultFor[string]{
					Content: []mcp.Content{&mcp.TextContent{Text: result}},
				}, nil
			},
		),
		mcp.NewServerTool(
			"create_authpolicy",
			"Generate a Kuadrant AuthPolicy manifest",
			func(ctx context.Context, _ *mcp.ServerSession, params *mcp.CallToolParamsFor[CreateAuthPolicyParams]) (*mcp.CallToolResultFor[string], error) {
				result, err := createAuthPolicyHandler(ctx, params.Arguments)
				if err != nil {
					return nil, err
				}
				return &mcp.CallToolResultFor[string]{
					Content: []mcp.Content{&mcp.TextContent{Text: result}},
				}, nil
			},
		),
	)

	// Add resources for Kuadrant documentation (from resources.go)
	addKuadrantResources(server)

	ctx := context.Background()

	switch *transport {
	case "stdio":
		// Run with stdio transport (default)
		if err := server.Run(ctx, mcp.NewStdioTransport()); err != nil {
			log.Fatal(err)
		}

	case "sse":
		// Run with SSE transport
		log.Printf("Starting SSE server on %s", *addr)
		handler := mcp.NewSSEHandler(func(r *http.Request) *mcp.Server {
			return server
		})
		if err := http.ListenAndServe(*addr, handler); err != nil {
			log.Fatal(err)
		}

	case "http":
		// Run with StreamableHTTP transport
		log.Printf("Starting StreamableHTTP server on %s", *addr)
		handler := mcp.NewStreamableHTTPHandler(func(r *http.Request) *mcp.Server {
			return server
		}, nil)
		if err := http.ListenAndServe(*addr, handler); err != nil {
			log.Fatal(err)
		}

	default:
		log.Fatalf("Unknown transport: %s", *transport)
	}
}

