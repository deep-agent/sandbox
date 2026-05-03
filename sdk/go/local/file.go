package local

import (
	"bytes"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/deep-agent/sandbox/internal/services/filesystem"
	"github.com/deep-agent/sandbox/types/model"
)

func (c *Client) FileRead(req *model.FileReadRequest) (*model.FileReadResult, error) {
	var content string
	var err error

	if req.Base64 {
		content, err = c.fileManager.ReadFileBase64(req.File)
	} else {
		content, err = c.fileManager.ReadFile(req.File)
	}

	if err != nil {
		return nil, err
	}

	return &model.FileReadResult{
		Content: content,
	}, nil
}

func (c *Client) FileWrite(req *model.FileWriteRequest) error {
	writeOpts := filesystem.WriteOptions{Mode: os.FileMode(req.Mode), Atomic: req.Atomic}
	if req.Base64 {
		return c.fileManager.WriteFileBase64(req.File, req.Content, writeOpts)
	}
	return c.fileManager.WriteFile(req.File, req.Content, writeOpts)
}

func (c *Client) FileList(req *model.FileListRequest) (*model.FileListResult, error) {
	files, err := c.fileManager.ListDir(req.Path)
	if err != nil {
		return nil, err
	}

	return &model.FileListResult{
		Files: files,
	}, nil
}

func (c *Client) FileDelete(req *model.FileDeleteRequest) error {
	return c.fileManager.DeleteFile(req.Path)
}

func (c *Client) FileMove(req *model.FileMoveRequest) error {
	return c.fileManager.MoveFile(req.Source, req.Destination)
}

func (c *Client) FileCopy(req *model.FileCopyRequest) error {
	return c.fileManager.CopyFile(req.Source, req.Destination)
}

func (c *Client) MkDir(req *model.MkDirRequest) error {
	return c.fileManager.MkDir(req.Path)
}

func (c *Client) FileExists(path string) (*model.FileExistsResult, error) {
	exists := c.fileManager.Exists(path)
	return &model.FileExistsResult{
		Exists: exists,
	}, nil
}

func (c *Client) FileUpload(filename string, reader io.Reader, destPath string) (*model.FileUploadResult, error) {
	if strings.HasSuffix(destPath, "/") {
		destPath = filepath.Join(destPath, filename)
	}

	content, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read upload content: %w", err)
	}

	dir := filepath.Dir(destPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(destPath, content, 0644); err != nil {
		return nil, fmt.Errorf("failed to write file: %w", err)
	}

	return &model.FileUploadResult{
		File: destPath,
		Size: int64(len(content)),
	}, nil
}

func (c *Client) FileDownload(filePath string) (io.ReadCloser, string, error) {
	info, err := os.Stat(filePath)
	if err != nil {
		return nil, "", err
	}
	if info.IsDir() {
		return nil, "", fmt.Errorf("path is a directory, not a file")
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, "", fmt.Errorf("failed to read file: %w", err)
	}

	contentType := localDetectContentType(filePath, content)
	return io.NopCloser(bytes.NewReader(content)), contentType, nil
}

func (c *Client) FileCreateTemp(req *model.FileCreateTempRequest) (*model.FileCreateTempResult, error) {
	var (
		file string
		err  error
	)
	if req.Base64 {
		file, err = c.fileManager.CreateTempFileBase64(req.Dir, req.Pattern, req.Content, os.FileMode(req.Mode))
	} else {
		file, err = c.fileManager.CreateTempFile(req.Dir, req.Pattern, []byte(req.Content), os.FileMode(req.Mode))
	}
	if err != nil {
		return nil, err
	}
	return &model.FileCreateTempResult{File: file}, nil
}

func (c *Client) FileGlob(req *model.FileGlobRequest) (*model.FileGlobResult, error) {
	result, err := c.fileManager.Glob(filesystem.GlobOptions{Path: req.Path, Pattern: req.Pattern, Limit: req.Limit})
	if err != nil {
		return nil, err
	}
	return &model.FileGlobResult{Files: result.Files, Count: result.Count, Truncated: result.Truncated, Output: result.Output}, nil
}

func (c *Client) FileEvalSymlinks(req *model.FileEvalSymlinksRequest) (*model.FileEvalSymlinksResult, error) {
	resolved, err := c.fileManager.EvalSymlinks(req.Path)
	if err != nil {
		return nil, err
	}
	return &model.FileEvalSymlinksResult{ResolvedPath: resolved}, nil
}

func (c *Client) FileAppend(req *model.FileAppendRequest) error {
	return c.fileManager.AppendFile(req.File, req.Content)
}

func (c *Client) FileStat(req *model.FileStatRequest) (*model.FileStatResult, error) {
	return c.fileManager.Stat(req.Path)
}

func localDetectContentType(filePath string, content []byte) string {
	ext := strings.ToLower(filepath.Ext(filePath))
	if ct := mime.TypeByExtension(ext); ct != "" {
		return ct
	}

	textExts := map[string]string{
		".go": "text/plain; charset=utf-8", ".py": "text/plain; charset=utf-8",
		".rs": "text/plain; charset=utf-8", ".rb": "text/plain; charset=utf-8",
		".java": "text/plain; charset=utf-8", ".c": "text/plain; charset=utf-8",
		".h": "text/plain; charset=utf-8", ".cpp": "text/plain; charset=utf-8",
		".sh": "text/plain; charset=utf-8", ".yml": "text/plain; charset=utf-8",
		".yaml": "text/plain; charset=utf-8", ".toml": "text/plain; charset=utf-8",
		".log": "text/plain; charset=utf-8", ".md": "text/plain; charset=utf-8",
		".txt": "text/plain; charset=utf-8",
	}
	if ct, ok := textExts[ext]; ok {
		return ct
	}

	return http.DetectContentType(content)
}
