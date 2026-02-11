package tools

import (
	"context"

	"github.com/deep-agent/sandbox/internal/services/filesystem"
	"github.com/mark3labs/mcp-go/mcp"
)

func WriteToolDef() mcp.Tool {
	return mcp.NewTool("Write",
		mcp.WithDescription("Writes a file to the local filesystem.\n\nUsage:\n- This tool will overwrite the existing file if there is one at the provided path.\n- If this is an existing file, you MUST use the Read tool first to read the file's contents. This tool will fail if you did not read the file first.\n- ALWAYS prefer editing existing files in the codebase. NEVER write new files unless explicitly required.\n- NEVER proactively create documentation files (*.md) or README files. Only create documentation files if explicitly requested by the User.\n- Only use emojis if the user explicitly requests it. Avoid writing emojis to files unless asked."),
		mcp.WithString("file_path",
			mcp.Required(),
			mcp.Description("The absolute path to the file to write (must be absolute, not relative)"),
		),
		mcp.WithString("content",
			mcp.Required(),
			mcp.Description("The content to write to the file"),
		),
	)
}

func WriteHandler(workspace string) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		filePath, err := request.RequireString("file_path")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		content, err := request.RequireString("content")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		fileManager := filesystem.NewManager(workspace)
		if err := fileManager.WriteFile(filePath, content); err != nil {
			return mcp.NewToolResultError("Error: " + err.Error()), nil
		}

		return mcp.NewToolResultText("File written successfully: " + filePath), nil
	}
}
