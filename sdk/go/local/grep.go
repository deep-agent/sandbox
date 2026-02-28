package local

import (
	"context"

	"github.com/deep-agent/sandbox/internal/services/filesystem"
	"github.com/deep-agent/sandbox/types/model"
)

func (c *Client) GrepSearch(req *model.GrepRequest) (*model.GrepResult, error) {
	ctx := context.Background()

	opts := filesystem.GrepOptions{
		Pattern:         req.Pattern,
		Path:            req.Path,
		Glob:            req.Glob,
		CaseInsensitive: req.CaseInsensitive,
		ContextLines:    req.ContextLines,
		OutputMode:      req.OutputMode,
		Limit:           req.Limit,
		MaxLineLength:   req.MaxLineLength,
	}

	result, err := c.fileManager.Grep(ctx, opts)
	if err != nil {
		return nil, err
	}

	return &model.GrepResult{
		Output:    result.Output,
		Truncated: result.Truncated,
	}, nil
}
