package main

import (
	"fmt"
	"log"

	"github.com/deep-agent/sandbox/internal/config"
	"github.com/deep-agent/sandbox/internal/mcp"
)

func main() {
	cfg := config.Load()

	log.Printf("Starting MCP Hub on port %d", cfg.MCPHubPort)
	server := mcp.NewServer("sandbox-mcp", "1.0.0", cfg.MCPHubPort)

	registry := mcp.NewRegistry(mcp.ToolConfig{
		Workspace: cfg.Workspace,
		CDPURL:    fmt.Sprintf("ws://localhost:%d", cfg.BrowserCDPPort),
	})
	registry.RegisterAll(server.AddTool)

	if err := server.Start(); err != nil {
		log.Fatalf("Failed to start MCP Hub: %v", err)
	}
}
