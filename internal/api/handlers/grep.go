package handlers

import (
	"context"
	"net/http"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/deep-agent/sandbox/internal/services/filesystem"
	"github.com/deep-agent/sandbox/types/model"
)

type GrepHandler struct {
	manager *filesystem.Manager
}

func NewGrepHandler(manager *filesystem.Manager) *GrepHandler {
	return &GrepHandler{manager: manager}
}

func (h *GrepHandler) Search(ctx context.Context, c *app.RequestContext) {
	var req model.GrepRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Code:    400,
			Message: "invalid request: " + err.Error(),
		})
		return
	}

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

	result, err := h.manager.Grep(ctx, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{
			Code:    500,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{
		Code: 0,
		Data: model.GrepResult{
			Output:    result.Output,
			Truncated: result.Truncated,
		},
	})
}
