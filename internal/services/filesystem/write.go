package filesystem

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func (m *Manager) WriteFile(path string, content string) error {
	absPath, err := m.validatePath(path)
	if err != nil {
		return err
	}

	dir := filepath.Dir(absPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(absPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

func (m *Manager) WriteFileBase64(path string, contentBase64 string) error {
	absPath, err := m.validatePath(path)
	if err != nil {
		return err
	}

	content, err := base64.StdEncoding.DecodeString(contentBase64)
	if err != nil {
		return fmt.Errorf("invalid base64 content: %w", err)
	}

	dir := filepath.Dir(absPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(absPath, content, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
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
