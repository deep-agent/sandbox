package config

import (
	"os"
	"strconv"
)

type Config struct {
	SandboxServerPort int
	MCPHubPort        int
	VNCServerPort     int
	WebSocketPort     int
	BrowserCDPPort    int
	HomeDir           string
}

func Load() *Config {
	homeDir := os.Getenv("HOME")
	if homeDir == "" {
		homeDir = "/home/sandbox/workspace"
	}

	return &Config{
		SandboxServerPort: getEnvInt("SANDBOX_SRV_PORT", 8000),
		MCPHubPort:        getEnvInt("MCP_HUB_PORT", 8001),
		VNCServerPort:     getEnvInt("VNC_SERVER_PORT", 5900),
		WebSocketPort:     getEnvInt("WEBSOCKET_PROXY_PORT", 6080),
		BrowserCDPPort:    getEnvInt("BROWSER_REMOTE_DEBUGGING_PORT", 9222),
		HomeDir:           homeDir,
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}
