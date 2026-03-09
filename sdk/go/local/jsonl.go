package local

import (
	"github.com/deep-agent/sandbox/types/model"
)

func (c *Client) JSONLCountLines(req *model.JSONLCountRequest) (*model.JSONLCountResult, error) {
	count, err := c.jsonlService.CountLines(req.File)
	if err != nil {
		return nil, err
	}

	return &model.JSONLCountResult{Lines: count}, nil
}

func (c *Client) JSONLReadLines(req *model.JSONLReadRequest) (*model.JSONLReadResult, error) {
	lines, err := c.jsonlService.ReadLines(req.File, req.StartLine, req.Count)
	if err != nil {
		return nil, err
	}

	return &model.JSONLReadResult{Lines: lines}, nil
}

func (c *Client) JSONLAppendLine(req *model.JSONLAppendRequest) error {
	return c.jsonlService.AppendLine(req.File, req.JSONString)
}
