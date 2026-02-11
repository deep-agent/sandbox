package sandbox

import (
	"encoding/json"
	"fmt"

	"github.com/deep-agent/sandbox/model"
)

func (c *Client) GrepSearch(req *model.GrepRequest) (*model.GrepResult, error) {
	resp, err := c.doRequest("POST", "/v1/grep/search", req)
	if err != nil {
		return nil, err
	}

	var result model.GrepResult
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal result: %w", err)
	}

	return &result, nil
}
