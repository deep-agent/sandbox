package main

import (
	"github.com/deep-agent/sandbox/internal/api"
	"github.com/deep-agent/sandbox/internal/config"
	"github.com/deep-agent/sandbox/pkg/logger"
)

func main() {
	cfg := config.Load()

	logger.Printf("Starting Sandbox Server on port %d", cfg.SandboxServerPort)
	logger.Printf("Workspace: %s", cfg.Workspace)
	logger.Printf("Browser CDP Port: %d", cfg.BrowserCDPPort)

	router := api.NewRouter(cfg)
	router.Setup()

	if err := router.Run(); err != nil {
		logger.Fatalf("Failed to start server: %v", err)
	}
}
