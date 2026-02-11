package sandbox

import (
	"encoding/json"
	"fmt"

	"github.com/deep-agent/sandbox/model"
)

func (c *Client) BashExec(req *model.BashExecRequest) (*model.BashExecResult, error) {
	resp, err := c.doRequest("POST", "/v1/bash/exec", req)
	if err != nil {
		return nil, err
	}

	var result model.BashExecResult
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal result: %w", err)
	}

	return &result, nil
}
