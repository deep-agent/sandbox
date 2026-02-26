package filesystem

import (
	"fmt"
	"path/filepath"
)

type Manager struct{}

func NewManager() *Manager {
	return &Manager{}
}

func (m *Manager) validatePath(path string) (string, error) {
	if path == "" {
		return "", fmt.Errorf("invalid path: path cannot be empty")
	}

	cleanPath := filepath.Clean(path)

	return cleanPath, nil
}
