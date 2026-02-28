package local

import (
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
	if req.Base64 {
		return c.fileManager.WriteFileBase64(req.File, req.Content)
	}
	return c.fileManager.WriteFile(req.File, req.Content)
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
