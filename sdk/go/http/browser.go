package http

import (
	"encoding/json"
	"fmt"

	"github.com/deep-agent/sandbox/model"
)

func (c *Client) BrowserGetInfo() (*model.BrowserInfo, error) {
	resp, err := c.doRequest("GET", "/v1/browser/info", nil)
	if err != nil {
		return nil, err
	}

	var result model.BrowserInfo
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal result: %w", err)
	}

	return &result, nil
}

func (c *Client) BrowserNavigate(req *model.BrowserNavigateRequest) error {
	_, err := c.doRequest("POST", "/v1/browser/navigate", req)
	return err
}

func (c *Client) BrowserScreenshot(req *model.BrowserScreenshotRequest) (*model.BrowserScreenshotResult, error) {
	resp, err := c.doRequest("POST", "/v1/browser/screenshot", req)
	if err != nil {
		return nil, err
	}

	var result model.BrowserScreenshotResult
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal result: %w", err)
	}

	return &result, nil
}

func (c *Client) BrowserClick(req *model.BrowserClickRequest) error {
	_, err := c.doRequest("POST", "/v1/browser/click", req)
	return err
}

func (c *Client) BrowserType(req *model.BrowserTypeRequest) error {
	_, err := c.doRequest("POST", "/v1/browser/type", req)
	return err
}

func (c *Client) BrowserEvaluate(req *model.BrowserEvaluateRequest) (*model.BrowserEvaluateResult, error) {
	resp, err := c.doRequest("POST", "/v1/browser/evaluate", req)
	if err != nil {
		return nil, err
	}

	var result model.BrowserEvaluateResult
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal result: %w", err)
	}

	return &result, nil
}

func (c *Client) BrowserScroll(req *model.BrowserScrollRequest) error {
	_, err := c.doRequest("POST", "/v1/browser/scroll", req)
	return err
}

func (c *Client) BrowserGetHTML(req *model.BrowserGetHTMLRequest) (*model.BrowserGetHTMLResult, error) {
	resp, err := c.doRequest("POST", "/v1/browser/html", req)
	if err != nil {
		return nil, err
	}

	var result model.BrowserGetHTMLResult
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal result: %w", err)
	}

	return &result, nil
}

func (c *Client) BrowserWaitVisible(req *model.BrowserWaitVisibleRequest) error {
	_, err := c.doRequest("POST", "/v1/browser/wait-visible", req)
	return err
}

func (c *Client) BrowserGetCurrentURL() (*model.BrowserURLResult, error) {
	resp, err := c.doRequest("GET", "/v1/browser/url", nil)
	if err != nil {
		return nil, err
	}

	var result model.BrowserURLResult
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal result: %w", err)
	}

	return &result, nil
}

func (c *Client) BrowserGetTitle() (*model.BrowserTitleResult, error) {
	resp, err := c.doRequest("GET", "/v1/browser/title", nil)
	if err != nil {
		return nil, err
	}

	var result model.BrowserTitleResult
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal result: %w", err)
	}

	return &result, nil
}

func (c *Client) BrowserGetPageInfo() (*model.BrowserPageInfo, error) {
	resp, err := c.doRequest("GET", "/v1/browser/page-info", nil)
	if err != nil {
		return nil, err
	}

	var result model.BrowserPageInfo
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal result: %w", err)
	}

	return &result, nil
}

func (c *Client) BrowserPDF() (*model.BrowserPDFResult, error) {
	resp, err := c.doRequest("POST", "/v1/browser/pdf", nil)
	if err != nil {
		return nil, err
	}

	var result model.BrowserPDFResult
	if err := json.Unmarshal(resp.Data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal result: %w", err)
	}

	return &result, nil
}
