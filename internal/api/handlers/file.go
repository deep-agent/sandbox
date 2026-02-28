package handlers

import (
	"context"
	"net/http"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/deep-agent/sandbox/internal/services/filesystem"
	"github.com/deep-agent/sandbox/types/model"
)

type FileHandler struct {
	manager *filesystem.Manager
}

func NewFileHandler(manager *filesystem.Manager) *FileHandler {
	return &FileHandler{manager: manager}
}

func (h *FileHandler) ReadFile(ctx context.Context, c *app.RequestContext) {
	var req model.FileReadRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Code:    400,
			Message: "invalid request: " + err.Error(),
		})
		return
	}

	var content string
	var err error

	if req.Base64 {
		content, err = h.manager.ReadFileBase64(req.File)
	} else {
		content, err = h.manager.ReadFile(req.File)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{
			Code:    500,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{
		Code: 0,
		Data: model.FileReadResult{Content: content},
	})
}

func (h *FileHandler) WriteFile(ctx context.Context, c *app.RequestContext) {
	var req model.FileWriteRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Code:    400,
			Message: "invalid request: " + err.Error(),
		})
		return
	}

	var err error
	if req.Base64 {
		err = h.manager.WriteFileBase64(req.File, req.Content)
	} else {
		err = h.manager.WriteFile(req.File, req.Content)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{
			Code:    500,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{
		Code:    0,
		Message: "success",
	})
}

func (h *FileHandler) ListDir(ctx context.Context, c *app.RequestContext) {
	var req model.FileListRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Code:    400,
			Message: "invalid request: " + err.Error(),
		})
		return
	}

	files, err := h.manager.ListDir(req.Path)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{
			Code:    500,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{
		Code: 0,
		Data: model.FileListResult{Files: files},
	})
}

func (h *FileHandler) DeleteFile(ctx context.Context, c *app.RequestContext) {
	var req model.FileDeleteRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Code:    400,
			Message: "invalid request: " + err.Error(),
		})
		return
	}

	if err := h.manager.DeleteFile(req.Path); err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{
			Code:    500,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{
		Code:    0,
		Message: "success",
	})
}

func (h *FileHandler) MoveFile(ctx context.Context, c *app.RequestContext) {
	var req model.FileMoveRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Code:    400,
			Message: "invalid request: " + err.Error(),
		})
		return
	}

	if err := h.manager.MoveFile(req.Source, req.Destination); err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{
			Code:    500,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{
		Code:    0,
		Message: "success",
	})
}

func (h *FileHandler) CopyFile(ctx context.Context, c *app.RequestContext) {
	var req model.FileCopyRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Code:    400,
			Message: "invalid request: " + err.Error(),
		})
		return
	}

	if err := h.manager.CopyFile(req.Source, req.Destination); err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{
			Code:    500,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{
		Code:    0,
		Message: "success",
	})
}

func (h *FileHandler) MkDir(ctx context.Context, c *app.RequestContext) {
	var req model.MkDirRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Code:    400,
			Message: "invalid request: " + err.Error(),
		})
		return
	}

	if err := h.manager.MkDir(req.Path); err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{
			Code:    500,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{
		Code:    0,
		Message: "success",
	})
}

func (h *FileHandler) Exists(ctx context.Context, c *app.RequestContext) {
	path := c.Query("path")
	if path == "" {
		c.JSON(http.StatusBadRequest, model.Response{
			Code:    400,
			Message: "path is required",
		})
		return
	}

	exists := h.manager.Exists(path)
	c.JSON(http.StatusOK, model.Response{
		Code: 0,
		Data: model.FileExistsResult{Exists: exists},
	})
}
