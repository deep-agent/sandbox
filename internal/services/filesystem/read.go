package filesystem

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"os"
	"strings"
)

func (m *Manager) ReadFile(path string) (string, error) {
	absPath, err := m.validatePath(path)
	if err != nil {
		return "", err
	}

	content, err := os.ReadFile(absPath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	return string(content), nil
}

func (m *Manager) ReadFileBase64(path string) (string, error) {
	absPath, err := m.validatePath(path)
	if err != nil {
		return "", err
	}

	content, err := os.ReadFile(absPath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	return base64.StdEncoding.EncodeToString(content), nil
}

type ReadOptions struct {
	Offset         int
	Limit          int
	MaxLineLength  int
	WithLineNumber bool
}

type ReadResult struct {
	Content   string
	LinesRead int
	TotalLine int
}

func (m *Manager) ReadFileWithOptions(path string, opts ReadOptions) (*ReadResult, error) {
	if opts.Offset < 1 {
		opts.Offset = 1
	}
	if opts.Limit < 1 {
		opts.Limit = 2000
	}
	if opts.MaxLineLength < 1 {
		opts.MaxLineLength = 2000
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	var result strings.Builder
	scanner := bufio.NewScanner(file)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	lineNum := 0
	linesRead := 0

	for scanner.Scan() {
		lineNum++
		if lineNum < opts.Offset {
			continue
		}
		if linesRead >= opts.Limit {
			break
		}

		line := scanner.Text()
		if len(line) > opts.MaxLineLength {
			line = line[:opts.MaxLineLength] + "..."
		}

		if opts.WithLineNumber {
			result.WriteString(fmt.Sprintf("%6d\t%s\n", lineNum, line))
		} else {
			result.WriteString(line)
			result.WriteString("\n")
		}
		linesRead++
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	return &ReadResult{
		Content:   result.String(),
		LinesRead: linesRead,
		TotalLine: lineNum,
	}, nil
}
