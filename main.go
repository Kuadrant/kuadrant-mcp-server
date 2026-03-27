package main

import (
	"context"
	"flag"
	"log"
	"net/http"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
	// Parse command line flags
	var (
		transport = flag.String("transport", "stdio", "Transport type: stdio, sse, http")
		addr      = flag.String("addr", ":8080", "Address to listen on (for sse/http transports)")
	)
	flag.Parse()

	log.Printf("[KUADRANT MCP] Starting debugging-focused server with transport=%s", *transport)

	// Create server
	server := mcp.NewServer("kuadrant-mcp", "2.0.0", nil)

	// Add debugging prompts (replaces tools)
	addDebugPromptsSimple(server)

	// Add debugging resources (embedded docs)
	addDebugResources(server)

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
