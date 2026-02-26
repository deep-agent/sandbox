package tools

import (
	"context"

	"github.com/deep-agent/sandbox/internal/services/filesystem"
	"github.com/mark3labs/mcp-go/mcp"
)

func EditToolDef() mcp.Tool {
	return mcp.NewTool("Edit",
		mcp.WithDescription("Performs exact string replacements in files.\n\nUsage:\n- You must use your `Read` tool at least once in the conversation before editing. This tool will error if you attempt an edit without reading the file. \n- When editing text from Read tool output, ensure you preserve the exact indentation (tabs/spaces) as it appears AFTER the line number prefix. The line number prefix format is: spaces + line number + tab. Everything after that tab is the actual file content to match. Never include any part of the line number prefix in the old_string or new_string.\n- ALWAYS prefer editing existing files in the codebase. NEVER write new files unless explicitly required.\n- Only use emojis if the user explicitly requests it. Avoid adding emojis to files unless asked.\n- The edit will FAIL if `old_string` is not unique in the file. Either provide a larger string with more surrounding context to make it unique or use `replace_all` to change every instance of `old_string`.\n- Use `replace_all` for replacing and renaming strings across the file. This parameter is useful if you want to rename a variable for instance."),
		mcp.WithString("file_path",
			mcp.Required(),
			mcp.Description("The absolute path to the file to modify"),
		),
		mcp.WithString("old_string",
			mcp.Required(),
			mcp.Description("The text to replace"),
		),
		mcp.WithString("new_string",
			mcp.Required(),
			mcp.Description("The text to replace it with (must be different from old_string)"),
		),
		mcp.WithBoolean("replace_all",
			mcp.Description("Replace all occurences of old_string (default false)"),
		),
	)
}

func EditHandler() func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		filePath, err := request.RequireString("file_path")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		oldString, err := request.RequireString("old_string")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		newString, err := request.RequireString("new_string")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		replaceAll := request.GetBool("replace_all", false)

		fileManager := filesystem.NewManager()
		err = fileManager.EditFile(filePath, oldString, newString, filesystem.EditOptions{
			ReplaceAll: replaceAll,
		})
		if err != nil {
			return mcp.NewToolResultError("Error: " + err.Error()), nil
		}

		return mcp.NewToolResultText("File edited successfully"), nil
	}
}
