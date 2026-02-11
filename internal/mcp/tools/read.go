package tools

import (
	"context"

	"github.com/deep-agent/sandbox/internal/services/filesystem"
	"github.com/mark3labs/mcp-go/mcp"
)

func ReadToolDef() mcp.Tool {
	return mcp.NewTool("Read",
		mcp.WithDescription("Reads a file from the local filesystem. You can access any file directly by using this tool.\nAssume this tool is able to read all files on the machine. If the User provides a path to a file assume that path is valid. It is okay to read a file that does not exist; an error will be returned.\n\nUsage:\n- The file_path parameter must be an absolute path, not a relative path\n- By default, it reads up to 2000 lines starting from the beginning of the file\n- You can optionally specify a line offset and limit (especially handy for long files), but it's recommended to read the whole file by not providing these parameters\n- Any lines longer than 2000 characters will be truncated\n- Results are returned using cat -n format, with line numbers starting at 1\n- This tool allows Claude Code to read images (eg PNG, JPG, etc). When reading an image file the contents are presented visually as Claude Code is a multimodal LLM.\n- This tool can read PDF files (.pdf). PDFs are processed page by page, extracting both text and visual content for analysis.\n- This tool can read Jupyter notebooks (.ipynb files) and returns all cells with their outputs, combining code, text, and visualizations.\n- This tool can only read files, not directories. To read a directory, use an ls command via the Bash tool.\n- You can call multiple tools in a single response. It is always better to speculatively read multiple potentially useful files in parallel.\n- You will regularly be asked to read screenshots. If the user provides a path to a screenshot, ALWAYS use this tool to view the file at the path. This tool will work with all temporary file paths.\n- If you read a file that exists but has empty contents you will receive a system reminder warning in place of file contents."),
		mcp.WithString("file_path",
			mcp.Required(),
			mcp.Description("The absolute path to the file to read"),
		),
		mcp.WithNumber("offset",
			mcp.Description("The line number to start reading from. Only provide if the file is too large to read at once"),
		),
		mcp.WithNumber("limit",
			mcp.Description("The number of lines to read. Only provide if the file is too large to read at once."),
		),
	)
}

func ReadHandler(workspace string) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		filePath, err := request.RequireString("file_path")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		offset := int(request.GetFloat("offset", 1))
		limit := int(request.GetFloat("limit", 2000))

		fileManager := filesystem.NewManager(workspace)
		result, err := fileManager.ReadFileWithOptions(filePath, filesystem.ReadOptions{
			Offset:         offset,
			Limit:          limit,
			MaxLineLength:  20000,
			WithLineNumber: true,
		})
		if err != nil {
			return mcp.NewToolResultError("Error: " + err.Error()), nil
		}

		if result.Content == "" {
			return mcp.NewToolResultText("<system-reminder>Warning: File exists but has empty contents</system-reminder>"), nil
		}

		return mcp.NewToolResultText(result.Content), nil
	}
}
