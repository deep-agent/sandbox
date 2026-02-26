package tools

import (
	"context"

	"github.com/deep-agent/sandbox/internal/services/filesystem"
	"github.com/deep-agent/sandbox/pkg/session"
	"github.com/mark3labs/mcp-go/mcp"
)

const (
	DefaultGrepLimit     = 100
	DefaultMaxLineLength = 2000
)

func GrepToolDef() mcp.Tool {
	return mcp.NewTool("Grep",
		mcp.WithDescription("- Fast content search tool using ripgrep\n- Searches file contents for patterns (regular expressions or literal strings)\n- Returns matching lines with file paths and line numbers\n- Use this tool when you need to find specific content within files\n- Supports various output modes and context lines"),
		mcp.WithString("pattern",
			mcp.Required(),
			mcp.Description("The pattern to search for (supports regex)"),
		),
		mcp.WithString("path",
			mcp.Description("The directory or file to search in. Defaults to workspace root."),
		),
		mcp.WithString("glob",
			mcp.Description("File glob pattern to filter files (e.g., '*.go', '*.{ts,tsx}')"),
		),
		mcp.WithBoolean("case_insensitive",
			mcp.Description("Perform case-insensitive search"),
		),
		mcp.WithNumber("context_lines",
			mcp.Description("Number of context lines to show before and after matches"),
		),
		mcp.WithString("output_mode",
			mcp.Description("Output mode: 'content' (default), 'files_only', or 'count'"),
		),
	)
}

func GrepHandler(homeDir string) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	manager := filesystem.NewManager()

	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		pattern, err := request.RequireString("pattern")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		searchPath := request.GetString("path", "")
		if searchPath == "" {
			searchPath = session.GetWorkspaceFromHeader(request.Header, homeDir)
		}

		opts := filesystem.GrepOptions{
			Pattern:         pattern,
			Path:            searchPath,
			Glob:            request.GetString("glob", ""),
			CaseInsensitive: request.GetBool("case_insensitive", false),
			ContextLines:    int(request.GetFloat("context_lines", 0)),
			OutputMode:      request.GetString("output_mode", "content"),
			Limit:           DefaultGrepLimit,
			MaxLineLength:   DefaultMaxLineLength,
		}

		result, err := manager.Grep(ctx, opts)
		if err != nil {
			return mcp.NewToolResultError("Error: " + err.Error()), nil
		}

		if result.Output == "" {
			return mcp.NewToolResultText("No matches found"), nil
		}

		return mcp.NewToolResultText(result.Output), nil
	}
}
