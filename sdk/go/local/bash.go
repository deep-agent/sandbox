package local

import (
	"context"
	"time"

	"github.com/deep-agent/sandbox/internal/services/bash"
	"github.com/deep-agent/sandbox/model"
)

func (c *Client) BashExec(req *model.BashExecRequest) (*model.BashExecResult, error) {
	ctx := context.Background()

	timeout := 30 * time.Second
	if req.TimeoutMS > 0 {
		timeout = time.Duration(req.TimeoutMS) * time.Millisecond
	}
	c.bashExecutor.SetTimeout(timeout)

	cwd := req.Cwd
	if cwd == "" {
		cwd = c.sandboxCtx.Workspace
	}

	if req.RunInBackground {
		result, err := c.bashExecutor.ExecuteBackground(ctx, req.Command, cwd, timeout)
		if err != nil {
			return nil, err
		}
		return &model.BashExecResult{
			Output:     result.Output,
			ExitCode:   result.ExitCode,
			OutputFile: result.OutputFile,
		}, nil
	}

	truncateOpts := &bash.TruncateOptions{
		MaxLines: 2000,
		MaxBytes: 100000,
	}

	result, err := c.bashExecutor.Execute(ctx, req.Command, cwd, truncateOpts)
	if err != nil {
		return nil, err
	}

	return &model.BashExecResult{
		Output:   result.Output,
		ExitCode: result.ExitCode,
	}, nil
}
