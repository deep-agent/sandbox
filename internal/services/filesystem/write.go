package filesystem

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var ErrInvalidBase64 = errors.New("invalid base64 content")

type WriteOptions struct {
	Mode   os.FileMode
	Atomic bool
}

func (m *Manager) WriteFile(path string, content string, opts ...WriteOptions) error {
	absPath, err := m.validatePath(path)
	if err != nil {
		return err
	}

	return m.writeBytes(absPath, []byte(content), resolveWriteOptions(opts...))
}

func (m *Manager) WriteFileBase64(path string, contentBase64 string, opts ...WriteOptions) error {
	absPath, err := m.validatePath(path)
	if err != nil {
		return err
	}

	content, err := base64.StdEncoding.DecodeString(contentBase64)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidBase64, err)
	}

	return m.writeBytes(absPath, content, resolveWriteOptions(opts...))
}

func resolveWriteOptions(opts ...WriteOptions) WriteOptions {
	if len(opts) == 0 {
		return WriteOptions{}
	}
	return opts[0]
}

func (m *Manager) writeBytes(absPath string, content []byte, opts WriteOptions) error {
	dir := filepath.Dir(absPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if opts.Atomic {
		mode := opts.Mode
		if mode == 0 {
			if info, err := os.Stat(absPath); err == nil {
				mode = info.Mode().Perm()
			} else {
				mode = 0644
			}
		}
		return atomicWriteFile(absPath, content, mode)
	}

	mode := opts.Mode
	if mode == 0 {
		mode = 0644
	}
	if err := os.WriteFile(absPath, content, mode); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

func (m *Manager) CreateTempFile(dir, pattern string, content []byte, mode os.FileMode) (string, error) {
	if dir == "" {
		dir = os.TempDir()
	}

	absDir, err := m.validatePath(dir)
	if err != nil {
		return "", err
	}
	if pattern == "" {
		pattern = "sandbox-*"
	}

	if err := os.MkdirAll(absDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	tmp, err := os.CreateTemp(absDir, pattern)
	if err != nil {
		return "", fmt.Errorf("failed to create temp file: %w", err)
	}
	defer tmp.Close()

	if len(content) > 0 {
		if _, err := tmp.Write(content); err != nil {
			_ = os.Remove(tmp.Name())
			return "", fmt.Errorf("failed to write temp file: %w", err)
		}
	}
	if err := tmp.Sync(); err != nil {
		_ = os.Remove(tmp.Name())
		return "", fmt.Errorf("failed to sync temp file: %w", err)
	}
	if mode != 0 {
		if err := tmp.Chmod(mode); err != nil {
			_ = os.Remove(tmp.Name())
			return "", fmt.Errorf("failed to chmod temp file: %w", err)
		}
	}

	return tmp.Name(), nil
}

func (m *Manager) CreateTempFileBase64(dir, pattern, contentBase64 string, mode os.FileMode) (string, error) {
	content, err := base64.StdEncoding.DecodeString(contentBase64)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrInvalidBase64, err)
	}
	return m.CreateTempFile(dir, pattern, content, mode)
}

func atomicWriteFile(path string, data []byte, mode os.FileMode) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".atomic-*")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmp.Name()
	cleanup := func() {
		_ = tmp.Close()
		_ = os.Remove(tmpPath)
	}

	if _, err := tmp.Write(data); err != nil {
		cleanup()
		return fmt.Errorf("failed to write temp file: %w", err)
	}
	if err := tmp.Sync(); err != nil {
		cleanup()
		return fmt.Errorf("failed to sync temp file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		cleanup()
		return fmt.Errorf("failed to close temp file: %w", err)
	}
	if err := os.Chmod(tmpPath, mode); err != nil {
		cleanup()
		return fmt.Errorf("failed to chmod temp file: %w", err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		cleanup()
		return fmt.Errorf("failed to rename temp file: %w", err)
	}

	if d, err := os.Open(dir); err == nil {
		_ = d.Sync()
		_ = d.Close()
	}

	return nil
}

type EditOptions struct {
	ReplaceAll bool
}

func (m *Manager) EditFile(path string, oldString, newString string, opts EditOptions) error {
	if oldString == newString {
		return fmt.Errorf("old_string and new_string must be different")
	}

	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	fileContent := string(content)

	search, err := FindReplacement(fileContent, oldString, opts.ReplaceAll)
	if err != nil {
		return err
	}

	if search == "" {
		return fmt.Errorf("old_string not found in file")
	}

	if opts.ReplaceAll {
		fileContent = strings.ReplaceAll(fileContent, search, newString)
	} else {
		index := strings.Index(fileContent, search)
		lastIndex := strings.LastIndex(fileContent, search)
		if index != lastIndex {
			return fmt.Errorf("found multiple matches for old_string. Provide more surrounding lines in old_string to identify the correct match")
		}
		fileContent = fileContent[:index] + newString + fileContent[index+len(search):]
	}

	if err := os.WriteFile(path, []byte(fileContent), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}
