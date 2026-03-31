package main

import (
	"github.com/deep-agent/sandbox/internal/config"
	"github.com/deep-agent/sandbox/internal/mcp"
	"github.com/deep-agent/sandbox/pkg/logger"
)

func main() {
	cfg := config.Load()

	logger.Printf("Starting MCP Hub on port %d", cfg.MCPHubPort)
	server := mcp.NewServer("sandbox-mcp", "1.0.0", cfg.MCPHubPort)

	registry := mcp.NewRegistry()
	registry.RegisterAll(server.AddTool)

	if err := server.Start(); err != nil {
		logger.Fatalf("Failed to start MCP Hub: %v", err)
	}
}
