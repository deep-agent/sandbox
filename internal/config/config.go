package config

import (
	"os"
	"strconv"
)

type Config struct {
	SandboxServerPort int
	MCPHubPort        int
	Workspace         string
}

func Load() *Config {
	workspace := os.Getenv("WORKSPACE")
	if workspace == "" {
		workspace = "/home/sandbox/workspaces"
	}

	return &Config{
		SandboxServerPort: getEnvInt("SANDBOX_SRV_PORT", 8000),
		MCPHubPort:        getEnvInt("MCP_HUB_PORT", 8001),
		Workspace:         workspace,
	}
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}
