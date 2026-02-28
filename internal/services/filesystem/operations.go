package filesystem

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/deep-agent/sandbox/types/model"
)

func (m *Manager) ListDir(path string) ([]model.FileInfo, error) {
	absPath, err := m.validatePath(path)
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to list directory: %w", err)
	}

	var files []model.FileInfo
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}

		files = append(files, model.FileInfo{
			Name:        entry.Name(),
			Path:        filepath.Join(path, entry.Name()),
			Size:        info.Size(),
			IsDir:       entry.IsDir(),
			Mode:        info.Mode().String(),
			ModTimeUnix: info.ModTime().Unix(),
		})
	}

	return files, nil
}

func (m *Manager) DeleteFile(path string) error {
	absPath, err := m.validatePath(path)
	if err != nil {
		return err
	}

	if err := os.RemoveAll(absPath); err != nil {
		return fmt.Errorf("failed to delete: %w", err)
	}

	return nil
}

func (m *Manager) MoveFile(src, dst string) error {
	absSrc, err := m.validatePath(src)
	if err != nil {
		return err
	}

	absDst, err := m.validatePath(dst)
	if err != nil {
		return err
	}

	if err := os.Rename(absSrc, absDst); err != nil {
		return fmt.Errorf("failed to move file: %w", err)
	}

	return nil
}

func (m *Manager) CopyFile(src, dst string) error {
	absSrc, err := m.validatePath(src)
	if err != nil {
		return err
	}

	absDst, err := m.validatePath(dst)
	if err != nil {
		return err
	}

	srcFile, err := os.Open(absSrc)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	dir := filepath.Dir(absDst)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	dstFile, err := os.Create(absDst)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	return nil
}

func (m *Manager) Exists(path string) bool {
	absPath, err := m.validatePath(path)
	if err != nil {
		return false
	}

	_, err = os.Stat(absPath)
	return err == nil
}

func (m *Manager) MkDir(path string) error {
	absPath, err := m.validatePath(path)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(absPath, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	return nil
}
