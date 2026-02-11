package sandbox

import (
	"encoding/json"
	"fmt"

	"github.com/deep-agent/sandbox/model"
)

func (c *Client) FileRead(req *model.FileReadRequest) (*model.FileReadResult, error) {
	resp, err := c.doRequest("POST", "/v1/file/read", req)
	if err != nil {
		return nil, err
	}

	var result model.FileReadResult
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal result: %w", err)
	}

	return &result, nil
}

func (c *Client) FileWrite(req *model.FileWriteRequest) error {
	_, err := c.doRequest("POST", "/v1/file/write", req)
	return err
}

func (c *Client) FileList(req *model.FileListRequest) (*model.FileListResult, error) {
	resp, err := c.doRequest("POST", "/v1/file/list", req)
	if err != nil {
		return nil, err
	}

	var result model.FileListResult
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal result: %w", err)
	}

	return &result, nil
}

func (c *Client) FileDelete(req *model.FileDeleteRequest) error {
	_, err := c.doRequest("POST", "/v1/file/delete", req)
	return err
}

func (c *Client) FileMove(req *model.FileMoveRequest) error {
	_, err := c.doRequest("POST", "/v1/file/move", req)
	return err
}

func (c *Client) FileCopy(req *model.FileCopyRequest) error {
	_, err := c.doRequest("POST", "/v1/file/copy", req)
	return err
}

func (c *Client) MkDir(req *model.MkDirRequest) error {
	_, err := c.doRequest("POST", "/v1/file/mkdir", req)
	return err
}

func (c *Client) FileExists(path string) (*model.FileExistsResult, error) {
	resp, err := c.doRequest("GET", "/v1/file/exists?path="+path, nil)
	if err != nil {
		return nil, err
	}

	var result model.FileExistsResult
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal result: %w", err)
	}

	return &result, nil
}
