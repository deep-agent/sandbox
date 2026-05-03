package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"

	"github.com/deep-agent/sandbox/types/consts"
	"github.com/deep-agent/sandbox/types/model"
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

func (c *Client) FileUpload(filename string, reader io.Reader, destPath string) (*model.FileUploadResult, error) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return nil, fmt.Errorf("failed to create form file: %w", err)
	}
	if _, err := io.Copy(part, reader); err != nil {
		return nil, fmt.Errorf("failed to copy file content: %w", err)
	}

	if err := writer.WriteField("path", destPath); err != nil {
		return nil, fmt.Errorf("failed to write path field: %w", err)
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed to close multipart writer: %w", err)
	}

	req, err := http.NewRequest("POST", c.baseURL+"/v1/file/upload", &body)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	if c.cwd != "" {
		req.Header.Set(consts.HeaderWorkspace, c.cwd)
	}
	if c.tokenProvider != nil {
		token, err := c.tokenProvider()
		if err != nil {
			return nil, fmt.Errorf("failed to get token: %w", err)
		}
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	var apiResp response
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if apiResp.Code != 0 {
		return nil, newAPIError(apiResp.Code, apiResp.Message)
	}

	var result model.FileUploadResult
	if err := json.Unmarshal(apiResp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal result: %w", err)
	}

	return &result, nil
}

func (c *Client) FileDownload(filePath string) (io.ReadCloser, string, error) {
	req, err := http.NewRequest("GET", c.baseURL+"/v1/file/download?file="+filePath, nil)
	if err != nil {
		return nil, "", fmt.Errorf("failed to create request: %w", err)
	}

	if c.cwd != "" {
		req.Header.Set(consts.HeaderWorkspace, c.cwd)
	}
	if c.tokenProvider != nil {
		token, err := c.tokenProvider()
		if err != nil {
			return nil, "", fmt.Errorf("failed to get token: %w", err)
		}
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("request failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, "", fmt.Errorf("download failed with status: %d", resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	return resp.Body, contentType, nil
}

func (c *Client) FileCreateTemp(req *model.FileCreateTempRequest) (*model.FileCreateTempResult, error) {
	resp, err := c.doRequest("POST", "/v1/file/create-temp", req)
	if err != nil {
		return nil, err
	}

	var result model.FileCreateTempResult
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal result: %w", err)
	}

	return &result, nil
}

func (c *Client) FileGlob(req *model.FileGlobRequest) (*model.FileGlobResult, error) {
	resp, err := c.doRequest("POST", "/v1/file/glob", req)
	if err != nil {
		return nil, err
	}

	var result model.FileGlobResult
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal result: %w", err)
	}

	return &result, nil
}

func (c *Client) FileEvalSymlinks(req *model.FileEvalSymlinksRequest) (*model.FileEvalSymlinksResult, error) {
	resp, err := c.doRequest("POST", "/v1/file/eval-symlinks", req)
	if err != nil {
		return nil, err
	}

	var result model.FileEvalSymlinksResult
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal result: %w", err)
	}

	return &result, nil
}

func (c *Client) FileAppend(req *model.FileAppendRequest) error {
	_, err := c.doRequest("POST", "/v1/file/append", req)
	return err
}

func (c *Client) FileStat(req *model.FileStatRequest) (*model.FileStatResult, error) {
	resp, err := c.doRequest("POST", "/v1/file/stat", req)
	if err != nil {
		return nil, err
	}

	var result model.FileStatResult
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal result: %w", err)
	}

	return &result, nil
}

func (c *Client) TempDir() (*model.TempDirResult, error) {
	resp, err := c.doRequest("GET", "/v1/file/temp-dir", nil)
	if err != nil {
		return nil, err
	}

	var result model.TempDirResult
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal result: %w", err)
	}

	return &result, nil
}

func (c *Client) UserHomeDir() (*model.UserHomeDirResult, error) {
	resp, err := c.doRequest("GET", "/v1/file/home-dir", nil)
	if err != nil {
		return nil, err
	}

	var result model.UserHomeDirResult
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal result: %w", err)
	}

	return &result, nil
}
