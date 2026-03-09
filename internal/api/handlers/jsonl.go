package handlers

import (
	"context"
	"net/http"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/deep-agent/sandbox/internal/services/jsonl"
	"github.com/deep-agent/sandbox/types/model"
)

type JSONLHandler struct {
	service *jsonl.Service
}

func NewJSONLHandler(service *jsonl.Service) *JSONLHandler {
	return &JSONLHandler{service: service}
}

func (h *JSONLHandler) CountLines(ctx context.Context, c *app.RequestContext) {
	var req model.JSONLCountRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Code:    400,
			Message: "invalid request: " + err.Error(),
		})
		return
	}

	count, err := h.service.CountLines(req.File)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{
			Code:    500,
			Message: "failed to count lines: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{
		Code: 0,
		Data: model.JSONLCountResult{Lines: count},
	})
}

func (h *JSONLHandler) ReadLines(ctx context.Context, c *app.RequestContext) {
	var req model.JSONLReadRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Code:    400,
			Message: "invalid request: " + err.Error(),
		})
		return
	}

	lines, err := h.service.ReadLines(req.File, req.StartLine, req.Count)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{
			Code:    500,
			Message: "failed to read lines: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{
		Code: 0,
		Data: model.JSONLReadResult{Lines: lines},
	})
}

func (h *JSONLHandler) AppendLine(ctx context.Context, c *app.RequestContext) {
	var req model.JSONLAppendRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Code:    400,
			Message: "invalid request: " + err.Error(),
		})
		return
	}

	if err := h.service.AppendLine(req.File, req.JSONString); err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{
			Code:    500,
			Message: "failed to append line: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{
		Code:    0,
		Message: "success",
	})
}
