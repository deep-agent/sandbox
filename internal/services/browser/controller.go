package browser

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/deep-agent/sandbox/types/consts"
	"github.com/deep-agent/sandbox/types/model"
)

type Controller struct {
	cdpURL  string
	timeout time.Duration

	allocCtx    context.Context
	allocCancel context.CancelFunc
	browserCtx  context.Context
	ctxCancel   context.CancelFunc
}

type ScreenshotOptions struct {
	Format  string `json:"format"`
	Quality int    `json:"quality"`
	Full    bool   `json:"full"`
}

type PageInfo struct {
	URL    string `json:"url"`
	Title  string `json:"title"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

func NewController(cdpURL string) *Controller {
	allocCtx, allocCancel := chromedp.NewRemoteAllocator(context.Background(), cdpURL)
	browserCtx, ctxCancel := chromedp.NewContext(allocCtx)

	return &Controller{
		cdpURL:      cdpURL,
		timeout:     consts.DefaultBrowserTimeout,
		allocCtx:    allocCtx,
		allocCancel: allocCancel,
		browserCtx:  browserCtx,
		ctxCancel:   ctxCancel,
	}
}

func (c *Controller) Close() {
	if c.ctxCancel != nil {
		c.ctxCancel()
	}
	if c.allocCancel != nil {
		c.allocCancel()
	}
}

func (c *Controller) GetInfo() (*model.BrowserInfo, error) {
	u, err := url.Parse(c.cdpURL)
	if err != nil {
		return &model.BrowserInfo{
			CDPURL: c.cdpURL,
			Status: "disconnected",
		}, nil
	}
	resp, err := http.Get(fmt.Sprintf("http://%s/json/version", u.Host))
	if err != nil {
		return &model.BrowserInfo{
			CDPURL: c.cdpURL,
			Status: "disconnected",
		}, nil
	}
	defer resp.Body.Close()

	return &model.BrowserInfo{
		CDPURL:    c.cdpURL,
		WebSocket: c.cdpURL,
		Status:    "connected",
	}, nil
}

func (c *Controller) run(actions ...chromedp.Action) error {
	// Use a timeout context but do NOT defer cancel() — canceling a child context
	// of the chromedp browser context would cause chromedp to close the tab.
	// The context will be garbage-collected after the timeout expires.
	ctx, _ := context.WithTimeout(c.browserCtx, c.timeout)
	return chromedp.Run(ctx, actions...)
}

func (c *Controller) Navigate(url string) error {
	return c.run(chromedp.Navigate(url))
}

func (c *Controller) Screenshot(opts *ScreenshotOptions) (string, error) {
	var buf []byte
	var action chromedp.Action

	if opts != nil && opts.Full {
		action = chromedp.FullScreenshot(&buf, 90)
	} else {
		action = chromedp.CaptureScreenshot(&buf)
	}

	if err := c.run(action); err != nil {
		return "", fmt.Errorf("failed to capture screenshot: %w", err)
	}

	return base64.StdEncoding.EncodeToString(buf), nil
}

func (c *Controller) GetCurrentURL() (string, error) {
	var url string
	if err := c.run(chromedp.Location(&url)); err != nil {
		return "", fmt.Errorf("failed to get current URL: %w", err)
	}

	return url, nil
}

func (c *Controller) GetTitle() (string, error) {
	var title string
	if err := c.run(chromedp.Title(&title)); err != nil {
		return "", fmt.Errorf("failed to get title: %w", err)
	}

	return title, nil
}

func (c *Controller) Click(selector string) error {
	return c.run(chromedp.Click(selector, chromedp.ByQuery))
}

func (c *Controller) Type(selector, text string) error {
	return c.run(
		chromedp.Click(selector, chromedp.ByQuery),
		chromedp.SendKeys(selector, text, chromedp.ByQuery),
	)
}

func (c *Controller) Evaluate(expression string) (interface{}, error) {
	var result interface{}
	if err := c.run(chromedp.Evaluate(expression, &result)); err != nil {
		return nil, fmt.Errorf("failed to evaluate: %w", err)
	}

	return result, nil
}

func (c *Controller) GetHTML(selector string) (string, error) {
	var html string
	if err := c.run(chromedp.OuterHTML(selector, &html)); err != nil {
		return "", fmt.Errorf("failed to get HTML: %w", err)
	}

	return html, nil
}

func (c *Controller) WaitVisible(selector string) error {
	return c.run(chromedp.WaitReady(selector))
}

func (c *Controller) Scroll(x, y int64) error {
	return c.run(chromedp.Evaluate(fmt.Sprintf("window.scrollTo(%d, %d)", x, y), nil))
}

func (c *Controller) GetPageInfo() (*PageInfo, error) {
	var url, title string
	if err := c.run(
		chromedp.Location(&url),
		chromedp.Title(&title),
	); err != nil {
		return nil, fmt.Errorf("failed to get page info: %w", err)
	}

	return &PageInfo{
		URL:   url,
		Title: title,
	}, nil
}

func (c *Controller) PDF() (string, error) {
	var buf []byte
	if err := c.run(chromedp.ActionFunc(func(ctx context.Context) error {
		var err error
		buf, _, err = page.PrintToPDF().WithPrintBackground(true).Do(ctx)
		return err
	})); err != nil {
		return "", fmt.Errorf("failed to generate PDF: %w", err)
	}

	return base64.StdEncoding.EncodeToString(buf), nil
}

func (c *Controller) GetCookies() (string, error) {
	var cookies interface{}
	if err := c.run(chromedp.Evaluate("document.cookie", &cookies)); err != nil {
		return "", fmt.Errorf("failed to get cookies: %w", err)
	}

	result, _ := json.Marshal(cookies)
	return string(result), nil
}
