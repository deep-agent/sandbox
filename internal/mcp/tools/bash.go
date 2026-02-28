package tools

import (
	"context"
	"time"

	"github.com/deep-agent/sandbox/internal/services/bash"
	"github.com/deep-agent/sandbox/pkg/ctxutil"
	"github.com/mark3labs/mcp-go/mcp"
)

func BashToolDef() mcp.Tool {
	return mcp.NewTool("Bash",
		mcp.WithDescription("Executes a given bash command with optional timeout. Working directory persists between commands; shell state (everything else) does not. The shell environment is initialized from the user's profile (bash or zsh).\n\nIMPORTANT: This tool is for terminal operations like git, npm, docker, etc. DO NOT use it for file operations (reading, writing, editing, searching, finding files) - use the specialized tools for this instead.\n\nBefore executing the command, please follow these steps:\n\n1. Directory Verification:\n   - If the command will create new directories or files, first use `ls` to verify the parent directory exists and is the correct location\n   - For example, before running \"mkdir foo/bar\", first use `ls foo` to check that \"foo\" exists and is the intended parent directory\n\n2. Command Execution:\n   - Always quote file paths that contain spaces with double quotes (e.g., cd \"path with spaces/file.txt\")\n   - Examples of proper quoting:\n     - cd \"/Users/name/My Documents\" (correct)\n     - cd /Users/name/My Documents (incorrect - will fail)\n     - python \"/path/with spaces/script.py\" (correct)\n     - python /path/with spaces/script.py (incorrect - will fail)\n   - After ensuring proper quoting, execute the command.\n   - Capture the output of the command.\n\nUsage notes:\n  - The command argument is required.\n  - You can specify an optional timeout in milliseconds (up to 600000ms / 10 minutes). If not specified, commands will timeout after 120000ms (2 minutes).\n  - It is very helpful if you write a clear, concise description of what this command does. For simple commands, keep it brief (5-10 words). For complex commands (piped commands, obscure flags, or anything hard to understand at a glance), add enough context to clarify what it does.\n  - If the output exceeds 30000 characters, output will be truncated before being returned to you.\n  - Set run_in_background to true to run the command in background. Output will be written to a file and you can use Read tool to view it later."),
		mcp.WithString("command",
			mcp.Required(),
			mcp.Description("The command to execute"),
		),
		mcp.WithNumber("timeout_ms",
			mcp.Description("Optional timeout in milliseconds (max 600000)"),
		),
		mcp.WithBoolean("run_in_background",
			mcp.Description(
				"When [run_in_background] is set to `true`, the command will run until it completes, and during this period, the user won't be able to interact with the Agent. You MUST ensure this value is set according to the following rules:\n\nAssign [run_in_background] to `false` only if:\n1. Launching a web server or dev server.\n2. Starting a long-running process that runs continuously (e.g., system services, monitoring processes, database servers, or message queues).\n\nOtherwise, set [run_in_background] to `true`. For example, if the command will finish in a relatively short amount of time, or it's important to review the command's output before responding to the user, make the command blocking.\n. Output will be written to a file."),
		),
		mcp.WithString("description",
			mcp.Description("Clear, concise description of what this command does in active voice. Never use words like \"complex\" or \"risk\" in the description - just describe what it does.\n\nFor simple commands (git, npm, standard CLI tools), keep it brief (5-10 words):\n- ls → \"List files in current directory\"\n- git status → \"Show working tree status\"\n- npm install → \"Install package dependencies\"\n\nFor commands that are harder to parse at a glance (piped commands, obscure flags, etc.), add enough context to clarify what it does:\n- find . -name \"*.tmp\" -exec rm {} \\; → \"Find and delete all .tmp files recursively\"\n- git reset --hard origin/main → \"Discard all local changes and match remote main\"\n- curl -s url | jq '.data[]' → \"Fetch JSON from URL and extract data array elements\""),
		),
	)
}

func BashHandler() func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		command, err := request.RequireString("command")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		cwd := ctxutil.GetCwd(ctx)

		runInBackground := request.GetBool("run_in_background", false)
		executor := bash.NewExecutor()

		if runInBackground {
			timeout := 10 * time.Minute
			if to := request.GetFloat("timeout_ms", 0); to > 0 {
				timeout = time.Duration(to) * time.Millisecond
			}
			result, err := executor.ExecuteBackground(ctx, command, cwd, timeout)
			if err != nil {
				return mcp.NewToolResultError("Error: " + err.Error()), nil
			}
			output := result.Output + "\n\nOutput file: " + result.OutputFile + "\nUse Read tool to view the output."
			return mcp.NewToolResultText(output), nil
		}

		timeoutMS := 30000
		if to := request.GetFloat("timeout_ms", 0); to > 0 {
			timeoutMS = int(to)
			if timeoutMS > 60000 {
				timeoutMS = 60000
			}
		}
		executor.SetTimeout(time.Duration(timeoutMS) * time.Millisecond)

		result, err := executor.Execute(ctx, command, cwd, &bash.TruncateOptions{MaxLines: 2000, MaxBytes: 50 * 1024})
		if err != nil {
			return mcp.NewToolResultError("Error: " + err.Error()), nil
		}

		output := result.Output
		if len(output) > 30000 {
			output = output[:30000] + "\n... (output truncated)"
		}

		if result.ExitCode != 0 {
			return mcp.NewToolResultError(output), nil
		}

		if result.TimedOut {
			return mcp.NewToolResultError(output), nil
		}

		return mcp.NewToolResultText(output), nil
	}
}
