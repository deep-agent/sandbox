package main

import (
	"log"

	"github.com/deep-agent/sandbox/internal/api"
	"github.com/deep-agent/sandbox/internal/config"
)

func main() {
	cfg := config.Load()

	log.Printf("Starting Sandbox Server on port %d", cfg.SandboxServerPort)
	log.Printf("Workspace: %s", cfg.Workspace)
	log.Printf("Browser CDP Port: %d", cfg.BrowserCDPPort)

	router := api.NewRouter(cfg)
	router.Setup()

	if err := router.Run(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
