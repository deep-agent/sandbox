package http

import (
	"encoding/json"
	"fmt"

	"github.com/deep-agent/sandbox/types/model"
)

func (c *Client) JSONLCountLines(req *model.JSONLCountRequest) (*model.JSONLCountResult, error) {
	resp, err := c.doRequest("POST", "/v1/jsonl/count", req)
	if err != nil {
		return nil, err
	}

	var result model.JSONLCountResult
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal result: %w", err)
	}

	return &result, nil
}

func (c *Client) JSONLReadLines(req *model.JSONLReadRequest) (*model.JSONLReadResult, error) {
	resp, err := c.doRequest("POST", "/v1/jsonl/read", req)
	if err != nil {
		return nil, err
	}

	var result model.JSONLReadResult
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal result: %w", err)
	}

	return &result, nil
}

func (c *Client) JSONLAppendLine(req *model.JSONLAppendRequest) error {
	_, err := c.doRequest("POST", "/v1/jsonl/append", req)
	return err
}
