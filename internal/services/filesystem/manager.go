package filesystem

import (
	"fmt"
	"path/filepath"
)

type Manager struct {
	baseDir string
}

func NewManager(baseDir string) *Manager {
	return &Manager{
		baseDir: baseDir,
	}
}

func (m *Manager) validatePath(path string) (string, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("invalid path: %w", err)
	}

	// if !strings.HasPrefix(absPath, m.baseDir) {
	// 	return "", fmt.Errorf("access denied: path outside workspace")
	// }

	return absPath, nil
}

func (m *Manager) GetBaseDir() string {
	return m.baseDir
}
