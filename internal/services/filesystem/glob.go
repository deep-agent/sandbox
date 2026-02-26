package filesystem

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

type GlobOptions struct {
	Path    string
	Pattern string
	Limit   int
}

type FileInfo struct {
	Path    string `json:"path"`
	ModTime int64  `json:"mtime"`
}

type GlobResult struct {
	Files     []string `json:"files"`
	Count     int      `json:"count"`
	Truncated bool     `json:"truncated"`
	Output    string   `json:"output"`
}

func (m *Manager) Glob(opts GlobOptions) (*GlobResult, error) {
	return m.GlobWithContext(context.Background(), opts)
}

func (m *Manager) GlobWithContext(ctx context.Context, opts GlobOptions) (*GlobResult, error) {
	searchPath := opts.Path
	if searchPath == "" {
		searchPath = "."
	}

	result, err := m.globWithRipgrep(ctx, searchPath, opts.Pattern, opts.Limit)
	if err != nil {
		return m.globWithWalk(searchPath, opts.Pattern, opts.Limit)
	}

	return result, nil
}

func (m *Manager) globWithRipgrep(ctx context.Context, searchPath, pattern string, limit int) (*GlobResult, error) {
	effectiveSearchPath := searchPath
	effectivePattern := pattern

	if strings.Contains(pattern, "/") && !strings.HasPrefix(pattern, "**/") {
		parts := strings.SplitN(pattern, "/", 2)
		if len(parts) == 2 {
			prefix := parts[0]
			if !strings.Contains(prefix, "*") && !strings.Contains(prefix, "?") {
				effectiveSearchPath = filepath.Join(searchPath, prefix)
				effectivePattern = parts[1]
			}
		}
	}

	args := []string{
		"--files",
		"--hidden",
		"--follow",
		"--no-messages",
		"--glob", effectivePattern,
		effectiveSearchPath,
	}

	cmd := exec.CommandContext(ctx, "rg", args...)
	output, err := cmd.Output()

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if exitErr.ExitCode() == 1 {
				return &GlobResult{
					Files:     []string{},
					Count:     0,
					Truncated: false,
					Output:    "No files found",
				}, nil
			}
		}
		return nil, fmt.Errorf("ripgrep failed: %w", err)
	}

	type fileWithTime struct {
		path    string
		modTime int64
	}

	var files []fileWithTime
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		filePath := scanner.Text()
		if filePath == "" {
			continue
		}

		var modTime int64
		if info, err := os.Stat(filePath); err == nil {
			modTime = info.ModTime().Unix()
		}

		files = append(files, fileWithTime{
			path:    filePath,
			modTime: modTime,
		})
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].modTime > files[j].modTime
	})

	truncated := false
	if limit > 0 && len(files) > limit {
		truncated = true
		files = files[:limit]
	}

	result := &GlobResult{
		Files:     make([]string, len(files)),
		Count:     len(files),
		Truncated: truncated,
	}

	for i, f := range files {
		result.Files[i] = f.path
	}

	result.Output = formatGlobOutput(result.Files, truncated)

	return result, nil
}

func (m *Manager) globWithWalk(searchPath, pattern string, limit int) (*GlobResult, error) {
	type fileWithTime struct {
		path    string
		modTime int64
	}

	var files []fileWithTime
	truncated := false

	err := filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			return nil
		}

		relPath, _ := filepath.Rel(searchPath, path)
		matched, _ := filepath.Match(pattern, filepath.Base(path))
		if !matched {
			matched = matchGlobPattern(pattern, relPath)
		}

		if matched {
			if limit > 0 && len(files) >= limit {
				truncated = true
				return filepath.SkipAll
			}
			files = append(files, fileWithTime{
				path:    path,
				modTime: info.ModTime().Unix(),
			})
		}
		return nil
	})

	if err != nil && err != filepath.SkipAll {
		return nil, fmt.Errorf("failed to glob: %w", err)
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].modTime > files[j].modTime
	})

	result := &GlobResult{
		Files:     make([]string, len(files)),
		Count:     len(files),
		Truncated: truncated,
	}

	for i, f := range files {
		result.Files[i] = f.path
	}

	result.Output = formatGlobOutput(result.Files, truncated)

	return result, nil
}

func formatGlobOutput(files []string, truncated bool) string {
	var output strings.Builder

	if len(files) == 0 {
		output.WriteString("No files found")
		return output.String()
	}

	for i, f := range files {
		if i > 0 {
			output.WriteString("\n")
		}
		output.WriteString(f)
	}

	if truncated {
		output.WriteString("\n\n(Results are truncated. Consider using a more specific path or pattern.)")
	}

	return output.String()
}

func matchGlobPattern(pattern, path string) bool {
	if strings.Contains(pattern, "**") {
		parts := strings.Split(pattern, "**")
		if len(parts) == 2 {
			prefix := strings.TrimSuffix(parts[0], "/")
			suffix := strings.TrimPrefix(parts[1], "/")

			if prefix != "" && !strings.HasPrefix(path, prefix) {
				return false
			}

			if suffix != "" {
				matched, _ := filepath.Match(suffix, filepath.Base(path))
				return matched
			}
			return true
		}
	}

	matched, _ := filepath.Match(pattern, path)
	return matched
}
