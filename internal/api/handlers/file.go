package handlers

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/deep-agent/sandbox/internal/services/filesystem"
	"github.com/deep-agent/sandbox/types/consts"
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
			Code:    consts.CodeBadRequest,
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
		if errors.Is(err, os.ErrNotExist) {
			c.JSON(http.StatusNotFound, model.Response{
				Code:    consts.CodeNotFound,
				Message: "file not found: " + req.File,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, model.Response{
			Code:    consts.CodeInternal,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{
		Code: consts.CodeSuccess,
		Data: model.FileReadResult{Content: content},
	})
}

func (h *FileHandler) WriteFile(ctx context.Context, c *app.RequestContext) {
	var req model.FileWriteRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Code:    consts.CodeBadRequest,
			Message: "invalid request: " + err.Error(),
		})
		return
	}

	var err error
	writeOpts := filesystem.WriteOptions{Mode: os.FileMode(req.Mode), Atomic: req.Atomic}
	if req.Base64 {
		err = h.manager.WriteFileBase64(req.File, req.Content, writeOpts)
	} else {
		err = h.manager.WriteFile(req.File, req.Content, writeOpts)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{
			Code:    consts.CodeInternal,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{
		Code:    consts.CodeSuccess,
		Message: "success",
	})
}

func (h *FileHandler) ListDir(ctx context.Context, c *app.RequestContext) {
	var req model.FileListRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Code:    consts.CodeBadRequest,
			Message: "invalid request: " + err.Error(),
		})
		return
	}

	files, err := h.manager.ListDir(req.Path)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{
			Code:    consts.CodeInternal,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{
		Code: consts.CodeSuccess,
		Data: model.FileListResult{Files: files},
	})
}

func (h *FileHandler) DeleteFile(ctx context.Context, c *app.RequestContext) {
	var req model.FileDeleteRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Code:    consts.CodeBadRequest,
			Message: "invalid request: " + err.Error(),
		})
		return
	}

	if err := h.manager.DeleteFile(req.Path); err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{
			Code:    consts.CodeInternal,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{
		Code:    consts.CodeSuccess,
		Message: "success",
	})
}

func (h *FileHandler) MoveFile(ctx context.Context, c *app.RequestContext) {
	var req model.FileMoveRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Code:    consts.CodeBadRequest,
			Message: "invalid request: " + err.Error(),
		})
		return
	}

	if err := h.manager.MoveFile(req.Source, req.Destination); err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{
			Code:    consts.CodeInternal,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{
		Code:    consts.CodeSuccess,
		Message: "success",
	})
}

func (h *FileHandler) CopyFile(ctx context.Context, c *app.RequestContext) {
	var req model.FileCopyRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Code:    consts.CodeBadRequest,
			Message: "invalid request: " + err.Error(),
		})
		return
	}

	if err := h.manager.CopyFile(req.Source, req.Destination); err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{
			Code:    consts.CodeInternal,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{
		Code:    consts.CodeSuccess,
		Message: "success",
	})
}

func (h *FileHandler) MkDir(ctx context.Context, c *app.RequestContext) {
	var req model.MkDirRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Code:    consts.CodeBadRequest,
			Message: "invalid request: " + err.Error(),
		})
		return
	}

	if err := h.manager.MkDir(req.Path); err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{
			Code:    consts.CodeInternal,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{
		Code:    consts.CodeSuccess,
		Message: "success",
	})
}

func (h *FileHandler) Exists(ctx context.Context, c *app.RequestContext) {
	path := c.Query("path")
	if path == "" {
		c.JSON(http.StatusBadRequest, model.Response{
			Code:    consts.CodeBadRequest,
			Message: "path is required",
		})
		return
	}

	exists := h.manager.Exists(path)
	c.JSON(http.StatusOK, model.Response{
		Code: consts.CodeSuccess,
		Data: model.FileExistsResult{Exists: exists},
	})
}

func (h *FileHandler) Upload(ctx context.Context, c *app.RequestContext) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Code:    consts.CodeBadRequest,
			Message: "file is required: " + err.Error(),
		})
		return
	}

	destPath := string(c.FormValue("path"))
	if destPath == "" {
		c.JSON(http.StatusBadRequest, model.Response{
			Code:    consts.CodeBadRequest,
			Message: "path is required",
		})
		return
	}

	// If path is a directory, append the original filename
	if strings.HasSuffix(destPath, "/") {
		destPath = filepath.Join(destPath, file.Filename)
	}

	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{
			Code:    consts.CodeInternal,
			Message: "failed to open uploaded file: " + err.Error(),
		})
		return
	}
	defer src.Close()

	content, err := io.ReadAll(src)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{
			Code:    consts.CodeInternal,
			Message: "failed to read uploaded file: " + err.Error(),
		})
		return
	}

	dir := filepath.Dir(destPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{
			Code:    consts.CodeInternal,
			Message: "failed to create directory: " + err.Error(),
		})
		return
	}

	if err := os.WriteFile(destPath, content, 0644); err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{
			Code:    consts.CodeInternal,
			Message: "failed to write file: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{
		Code: consts.CodeSuccess,
		Data: model.FileUploadResult{
			File: destPath,
			Size: int64(len(content)),
		},
	})
}

func (h *FileHandler) Download(ctx context.Context, c *app.RequestContext) {
	filePath := c.Query("file")
	if filePath == "" {
		c.JSON(http.StatusBadRequest, model.Response{
			Code:    consts.CodeBadRequest,
			Message: "file is required",
		})
		return
	}

	info, err := os.Stat(filePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			c.JSON(http.StatusNotFound, model.Response{
				Code:    consts.CodeNotFound,
				Message: "file not found: " + filePath,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, model.Response{
			Code:    consts.CodeInternal,
			Message: err.Error(),
		})
		return
	}

	if info.IsDir() {
		c.JSON(http.StatusBadRequest, model.Response{
			Code:    consts.CodeBadRequest,
			Message: "path is a directory, not a file",
		})
		return
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{
			Code:    consts.CodeInternal,
			Message: "failed to read file: " + err.Error(),
		})
		return
	}

	// Determine content type for inline display
	contentType := detectContentType(filePath, content)
	c.Response.Header.Set("Content-Type", contentType)

	// For text and images, use inline disposition so the browser renders them directly
	if isInlineContentType(contentType) {
		c.Response.Header.Set("Content-Disposition", fmt.Sprintf("inline; filename=\"%s\"", filepath.Base(filePath)))
	} else {
		c.Response.Header.Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filepath.Base(filePath)))
	}

	c.Response.Header.Set("Content-Length", fmt.Sprintf("%d", len(content)))
	c.SetStatusCode(http.StatusOK)
	c.Response.SetBody(content)
}

func (h *FileHandler) CreateTemp(ctx context.Context, c *app.RequestContext) {
	var req model.FileCreateTempRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{Code: consts.CodeBadRequest, Message: "invalid request: " + err.Error()})
		return
	}

	var (
		file string
		err  error
	)
	if req.Base64 {
		file, err = h.manager.CreateTempFileBase64(req.Dir, req.Pattern, req.Content, os.FileMode(req.Mode))
	} else {
		file, err = h.manager.CreateTempFile(req.Dir, req.Pattern, []byte(req.Content), os.FileMode(req.Mode))
	}
	if err != nil {
		if errors.Is(err, filesystem.ErrInvalidBase64) {
			c.JSON(http.StatusBadRequest, model.Response{Code: consts.CodeBadRequest, Message: err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, model.Response{Code: consts.CodeInternal, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, model.Response{Code: consts.CodeSuccess, Data: model.FileCreateTempResult{File: file}})
}

func (h *FileHandler) Glob(ctx context.Context, c *app.RequestContext) {
	var req model.FileGlobRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{Code: consts.CodeBadRequest, Message: "invalid request: " + err.Error()})
		return
	}

	result, err := h.manager.GlobWithContext(ctx, filesystem.GlobOptions{Path: req.Path, Pattern: req.Pattern, Limit: req.Limit})
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Code: consts.CodeInternal, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, model.Response{Code: consts.CodeSuccess, Data: model.FileGlobResult{Files: result.Files, Count: result.Count, Truncated: result.Truncated, Output: result.Output}})
}

func (h *FileHandler) EvalSymlinks(ctx context.Context, c *app.RequestContext) {
	var req model.FileEvalSymlinksRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{Code: consts.CodeBadRequest, Message: "invalid request: " + err.Error()})
		return
	}

	resolved, err := h.manager.EvalSymlinks(req.Path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			c.JSON(http.StatusNotFound, model.Response{Code: consts.CodeNotFound, Message: err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, model.Response{Code: consts.CodeInternal, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, model.Response{Code: consts.CodeSuccess, Data: model.FileEvalSymlinksResult{ResolvedPath: resolved}})
}

func (h *FileHandler) Append(ctx context.Context, c *app.RequestContext) {
	var req model.FileAppendRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{Code: consts.CodeBadRequest, Message: "invalid request: " + err.Error()})
		return
	}

	if err := h.manager.AppendFile(req.File, req.Content); err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Code: consts.CodeInternal, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, model.Response{Code: consts.CodeSuccess, Message: "success"})
}

func (h *FileHandler) Stat(ctx context.Context, c *app.RequestContext) {
	var req model.FileStatRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{Code: consts.CodeBadRequest, Message: "invalid request: " + err.Error()})
		return
	}

	result, err := h.manager.Stat(req.Path)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{Code: consts.CodeInternal, Message: err.Error()})
		return
	}

	c.JSON(http.StatusOK, model.Response{Code: consts.CodeSuccess, Data: result})
}

func detectContentType(filePath string, content []byte) string {
	// Try extension-based detection first for accuracy
	ext := strings.ToLower(filepath.Ext(filePath))
	if ct := mime.TypeByExtension(ext); ct != "" {
		return ct
	}

	// Common text file extensions that mime.TypeByExtension might miss
	textExts := map[string]string{
		".go":   "text/plain; charset=utf-8",
		".py":   "text/plain; charset=utf-8",
		".rs":   "text/plain; charset=utf-8",
		".rb":   "text/plain; charset=utf-8",
		".java": "text/plain; charset=utf-8",
		".c":    "text/plain; charset=utf-8",
		".h":    "text/plain; charset=utf-8",
		".cpp":  "text/plain; charset=utf-8",
		".hpp":  "text/plain; charset=utf-8",
		".sh":   "text/plain; charset=utf-8",
		".yml":  "text/plain; charset=utf-8",
		".yaml": "text/plain; charset=utf-8",
		".toml": "text/plain; charset=utf-8",
		".ini":  "text/plain; charset=utf-8",
		".conf": "text/plain; charset=utf-8",
		".cfg":  "text/plain; charset=utf-8",
		".log":  "text/plain; charset=utf-8",
		".md":   "text/plain; charset=utf-8",
		".txt":  "text/plain; charset=utf-8",
	}
	if ct, ok := textExts[ext]; ok {
		return ct
	}

	// Fall back to http.DetectContentType
	return http.DetectContentType(content)
}

func isInlineContentType(contentType string) bool {
	ct := strings.ToLower(contentType)
	return strings.HasPrefix(ct, "text/") ||
		strings.HasPrefix(ct, "image/") ||
		ct == "application/pdf" ||
		ct == "application/json" ||
		strings.HasPrefix(ct, "video/") ||
		strings.HasPrefix(ct, "audio/")
}
