package filesystem

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type GrepOptions struct {
	Pattern         string
	Path            string
	Glob            string
	CaseInsensitive bool
	ContextLines    int
	OutputMode      string
	Limit           int
	MaxLineLength   int
}

type GrepMatch struct {
	Path     string `json:"path"`
	LineNum  int    `json:"line_num"`
	LineText string `json:"line_text"`
	ModTimeUnix  int64  `json:"mod_time_unix"`
}

type GrepResult struct {
	Matches   []GrepMatch `json:"matches"`
	Count     int         `json:"count"`
	Truncated bool        `json:"truncated"`
	HasErrors bool        `json:"has_errors"`
	Output    string      `json:"output"`
}

func (m *Manager) Grep(ctx context.Context, opts GrepOptions) (*GrepResult, error) {
	searchPath := opts.Path
	if searchPath == "" {
		searchPath = m.baseDir
	}

	if !filepath.IsAbs(searchPath) {
		searchPath = filepath.Join(m.baseDir, searchPath)
	}

	var args []string
	args = append(args, "--color=never", "--no-messages")

	switch opts.OutputMode {
	case "files_only":
		args = append(args, "-l")
	case "count":
		args = append(args, "-c")
	default:
		args = append(args, "-nH")
		if opts.ContextLines > 0 {
			args = append(args, fmt.Sprintf("-C%d", opts.ContextLines))
		}
	}

	if opts.CaseInsensitive {
		args = append(args, "-i")
	}

	if opts.Glob != "" {
		args = append(args, "--glob", opts.Glob)
	}

	args = append(args, "--regexp", opts.Pattern, searchPath)

	cmd := exec.CommandContext(ctx, "rg", args...)
	output, err := cmd.Output()

	exitCode := 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			return nil, fmt.Errorf("failed to execute ripgrep: %w", err)
		}
	}

	outputStr := string(output)

	if exitCode == 1 || (exitCode == 2 && strings.TrimSpace(outputStr) == "") {
		return &GrepResult{
			Matches:   []GrepMatch{},
			Count:     0,
			Truncated: false,
			HasErrors: false,
			Output:    "No matches found",
		}, nil
	}

	hasErrors := exitCode == 2

	switch opts.OutputMode {
	case "files_only":
		return m.handleFilesOnlyMode(outputStr, hasErrors)
	case "count":
		return m.handleCountMode(outputStr, hasErrors)
	default:
		return m.handleContentMode(outputStr, opts.Limit, opts.MaxLineLength, hasErrors, opts.ContextLines > 0)
	}
}

func (m *Manager) handleFilesOnlyMode(output string, hasErrors bool) (*GrepResult, error) {
	lines := strings.Split(strings.TrimSpace(output), "\n")
	var files []string
	for _, line := range lines {
		if line != "" {
			files = append(files, line)
		}
	}

	if len(files) == 0 {
		return &GrepResult{
			Matches:   []GrepMatch{},
			Count:     0,
			Truncated: false,
			HasErrors: hasErrors,
			Output:    "No matches found",
		}, nil
	}

	return &GrepResult{
		Matches:   []GrepMatch{},
		Count:     len(files),
		Truncated: false,
		HasErrors: hasErrors,
		Output:    strings.Join(files, "\n"),
	}, nil
}

func (m *Manager) handleCountMode(output string, hasErrors bool) (*GrepResult, error) {
	return &GrepResult{
		Matches:   []GrepMatch{},
		Count:     0,
		Truncated: false,
		HasErrors: hasErrors,
		Output:    strings.TrimSpace(output),
	}, nil
}

func (m *Manager) handleContentMode(output string, limit int, maxLineLength int, hasErrors bool, hasContext bool) (*GrepResult, error) {
	if hasContext {
		return &GrepResult{
			Matches:   []GrepMatch{},
			Count:     0,
			Truncated: false,
			HasErrors: hasErrors,
			Output:    output,
		}, nil
	}

	matches, err := parseGrepOutput(output)
	if err != nil {
		return nil, err
	}

	sort.Slice(matches, func(i, j int) bool {
		return matches[i].ModTimeUnix > matches[j].ModTimeUnix
	})

	truncated := false
	if limit > 0 && len(matches) > limit {
		truncated = true
		matches = matches[:limit]
	}

	if len(matches) == 0 {
		return &GrepResult{
			Matches:   []GrepMatch{},
			Count:     0,
			Truncated: false,
			HasErrors: hasErrors,
			Output:    "No matches found",
		}, nil
	}

	formattedOutput := formatGrepOutput(matches, truncated, hasErrors, maxLineLength)

	return &GrepResult{
		Matches:   matches,
		Count:     len(matches),
		Truncated: truncated,
		HasErrors: hasErrors,
		Output:    formattedOutput,
	}, nil
}

func parseGrepOutput(output string) ([]GrepMatch, error) {
	var matches []GrepMatch

	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		firstColon := strings.Index(line, ":")
		if firstColon == -1 {
			continue
		}

		filePath := line[:firstColon]
		rest := line[firstColon+1:]

		secondColon := strings.Index(rest, ":")
		if secondColon == -1 {
			continue
		}

		lineNumStr := rest[:secondColon]
		lineText := rest[secondColon+1:]

		lineNum, err := strconv.Atoi(lineNumStr)
		if err != nil {
			continue
		}

		var modTime int64
		if info, err := os.Stat(filePath); err == nil {
			modTime = info.ModTime().Unix()
		}

		matches = append(matches, GrepMatch{
			Path:        filePath,
			LineNum:     lineNum,
			LineText:    lineText,
			ModTimeUnix: modTime,
		})
	}

	return matches, scanner.Err()
}

func formatGrepOutput(matches []GrepMatch, truncated, hasErrors bool, maxLineLength int) string {
	var output strings.Builder

	output.WriteString(fmt.Sprintf("Found %d matches\n", len(matches)))

	currentFile := ""
	for _, match := range matches {
		if currentFile != match.Path {
			if currentFile != "" {
				output.WriteString("\n")
			}
			currentFile = match.Path
			output.WriteString(fmt.Sprintf("%s:\n", match.Path))
		}

		lineText := match.LineText
		if maxLineLength > 0 && len(lineText) > maxLineLength {
			lineText = lineText[:maxLineLength] + "..."
		}
		output.WriteString(fmt.Sprintf("  Line %d: %s\n", match.LineNum, lineText))
	}

	if truncated {
		output.WriteString("\n(Results are truncated. Consider using a more specific path or pattern.)\n")
	}

	if hasErrors {
		output.WriteString("\n(Some paths were inaccessible and skipped)\n")
	}

	return output.String()
}
