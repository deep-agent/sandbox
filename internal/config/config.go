package config

import (
	"os"
	"strconv"
)

type Config struct {
	SandboxServerPort int
	MCPHubPort        int
}

func Load() *Config {
	return &Config{
		SandboxServerPort: getEnvInt("SANDBOX_SRV_PORT", 8000),
		MCPHubPort:        getEnvInt("MCP_HUB_PORT", 8001),
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
