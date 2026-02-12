package local

import (
	"fmt"

	"github.com/deep-agent/sandbox/internal/services/browser"
	"github.com/deep-agent/sandbox/model"
)

func (c *Client) ensureBrowser() error {
	if c.browserCtrl == nil {
		return fmt.Errorf("browser controller not initialized, use WithBrowserCDP option")
	}
	return nil
}

func (c *Client) BrowserGetInfo() (*model.BrowserInfo, error) {
	if err := c.ensureBrowser(); err != nil {
		return nil, err
	}
	return c.browserCtrl.GetInfo()
}

func (c *Client) BrowserNavigate(req *model.BrowserNavigateRequest) error {
	if err := c.ensureBrowser(); err != nil {
		return err
	}
	return c.browserCtrl.Navigate(req.URL)
}

func (c *Client) BrowserScreenshot(req *model.BrowserScreenshotRequest) (*model.BrowserScreenshotResult, error) {
	if err := c.ensureBrowser(); err != nil {
		return nil, err
	}

	opts := &browser.ScreenshotOptions{
		Format:  req.Format,
		Quality: req.Quality,
		Full:    req.Full,
	}

	screenshot, err := c.browserCtrl.Screenshot(opts)
	if err != nil {
		return nil, err
	}

	return &model.BrowserScreenshotResult{
		Screenshot: screenshot,
	}, nil
}

func (c *Client) BrowserClick(req *model.BrowserClickRequest) error {
	if err := c.ensureBrowser(); err != nil {
		return err
	}
	return c.browserCtrl.Click(req.Selector)
}

func (c *Client) BrowserType(req *model.BrowserTypeRequest) error {
	if err := c.ensureBrowser(); err != nil {
		return err
	}
	return c.browserCtrl.Type(req.Selector, req.Text)
}

func (c *Client) BrowserEvaluate(req *model.BrowserEvaluateRequest) (*model.BrowserEvaluateResult, error) {
	if err := c.ensureBrowser(); err != nil {
		return nil, err
	}

	result, err := c.browserCtrl.Evaluate(req.Expression)
	if err != nil {
		return nil, err
	}

	return &model.BrowserEvaluateResult{
		Result: result,
	}, nil
}

func (c *Client) BrowserScroll(req *model.BrowserScrollRequest) error {
	if err := c.ensureBrowser(); err != nil {
		return err
	}
	return c.browserCtrl.Scroll(req.X, req.Y)
}

func (c *Client) BrowserGetHTML(req *model.BrowserGetHTMLRequest) (*model.BrowserGetHTMLResult, error) {
	if err := c.ensureBrowser(); err != nil {
		return nil, err
	}

	html, err := c.browserCtrl.GetHTML(req.Selector)
	if err != nil {
		return nil, err
	}

	return &model.BrowserGetHTMLResult{
		HTML: html,
	}, nil
}

func (c *Client) BrowserWaitVisible(req *model.BrowserWaitVisibleRequest) error {
	if err := c.ensureBrowser(); err != nil {
		return err
	}
	return c.browserCtrl.WaitVisible(req.Selector)
}

func (c *Client) BrowserGetCurrentURL() (*model.BrowserURLResult, error) {
	if err := c.ensureBrowser(); err != nil {
		return nil, err
	}

	url, err := c.browserCtrl.GetCurrentURL()
	if err != nil {
		return nil, err
	}

	return &model.BrowserURLResult{
		URL: url,
	}, nil
}

func (c *Client) BrowserGetTitle() (*model.BrowserTitleResult, error) {
	if err := c.ensureBrowser(); err != nil {
		return nil, err
	}

	title, err := c.browserCtrl.GetTitle()
	if err != nil {
		return nil, err
	}

	return &model.BrowserTitleResult{
		Title: title,
	}, nil
}

func (c *Client) BrowserGetPageInfo() (*model.BrowserPageInfo, error) {
	if err := c.ensureBrowser(); err != nil {
		return nil, err
	}

	pageInfo, err := c.browserCtrl.GetPageInfo()
	if err != nil {
		return nil, err
	}

	return &model.BrowserPageInfo{
		URL:    pageInfo.URL,
		Title:  pageInfo.Title,
		Width:  pageInfo.Width,
		Height: pageInfo.Height,
	}, nil
}

func (c *Client) BrowserPDF() (*model.BrowserPDFResult, error) {
	if err := c.ensureBrowser(); err != nil {
		return nil, err
	}

	pdf, err := c.browserCtrl.PDF()
	if err != nil {
		return nil, err
	}

	return &model.BrowserPDFResult{
		PDF: pdf,
	}, nil
}
