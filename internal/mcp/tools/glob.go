package tools

import (
	"context"
	"strings"

	"github.com/deep-agent/sandbox/internal/services/filesystem"
	"github.com/deep-agent/sandbox/pkg/session"
	"github.com/mark3labs/mcp-go/mcp"
)

const DefaultGlobLimit = 100

func GlobToolDef() mcp.Tool {
	return mcp.NewTool("Glob",
		mcp.WithDescription("- Fast file pattern matching tool that works with any codebase size\n- Supports glob patterns like \"**/*.js\" or \"src/**/*.ts\"\n- Returns matching file paths sorted by modification time\n- Use this tool when you need to find files by name patterns\n- When you are doing an open ended search that may require multiple rounds of globbing and grepping, use the Agent tool instead\n- You can call multiple tools in a single response. It is always better to speculatively perform multiple searches in parallel if they are potentially useful."),
		mcp.WithString("pattern",
			mcp.Required(),
			mcp.Description("The glob pattern to match files against"),
		),
		mcp.WithString("path",
			mcp.Description("The directory to search in. If not specified, the current working directory will be used. IMPORTANT: Omit this field to use the default directory. DO NOT enter \"undefined\" or \"null\" - simply omit it for the default behavior. Must be a valid directory path if provided."),
		),
	)
}

func GlobHandler(homeDir string) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		pattern, err := request.RequireString("pattern")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		searchPath := request.GetString("path", "")
		if searchPath == "" {
			searchPath = session.GetWorkspaceFromHeader(request.Header, homeDir)
		}

		fileManager := filesystem.NewManager()
		result, err := fileManager.Glob(filesystem.GlobOptions{
			Path:    searchPath,
			Pattern: pattern,
			Limit:   DefaultGlobLimit,
		})
		if err != nil {
			return mcp.NewToolResultError("Error: " + err.Error()), nil
		}

		if len(result.Files) == 0 {
			return mcp.NewToolResultText("No files found matching pattern: " + pattern), nil
		}

		return mcp.NewToolResultText(strings.Join(result.Files, "\n")), nil
	}
}
