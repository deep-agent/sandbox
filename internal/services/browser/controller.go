package browser

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/deep-agent/sandbox/types/model"
)

type Controller struct {
	cdpURL  string
	timeout time.Duration
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
	return &Controller{
		cdpURL:  cdpURL,
		timeout: 30 * time.Second,
	}
}

func (c *Controller) GetInfo() (*model.BrowserInfo, error) {
	resp, err := http.Get(fmt.Sprintf("http://localhost:%s/json/version",
		c.cdpURL[len("ws://localhost:"):]))
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

func (c *Controller) createContext() (context.Context, context.CancelFunc) {
	allocCtx, _ := chromedp.NewRemoteAllocator(context.Background(), c.cdpURL)
	ctx, _ := chromedp.NewContext(allocCtx)
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	return ctx, cancel
}

func (c *Controller) Navigate(url string) error {
	ctx, cancel := c.createContext()
	defer cancel()

	return chromedp.Run(ctx, chromedp.Navigate(url))
}

func (c *Controller) Screenshot(opts *ScreenshotOptions) (string, error) {
	ctx, cancel := c.createContext()
	defer cancel()

	var buf []byte
	var action chromedp.Action

	if opts != nil && opts.Full {
		action = chromedp.FullScreenshot(&buf, 90)
	} else {
		action = chromedp.CaptureScreenshot(&buf)
	}

	if err := chromedp.Run(ctx, action); err != nil {
		return "", fmt.Errorf("failed to capture screenshot: %w", err)
	}

	return base64.StdEncoding.EncodeToString(buf), nil
}

func (c *Controller) GetCurrentURL() (string, error) {
	ctx, cancel := c.createContext()
	defer cancel()

	var url string
	if err := chromedp.Run(ctx, chromedp.Location(&url)); err != nil {
		return "", fmt.Errorf("failed to get current URL: %w", err)
	}

	return url, nil
}

func (c *Controller) GetTitle() (string, error) {
	ctx, cancel := c.createContext()
	defer cancel()

	var title string
	if err := chromedp.Run(ctx, chromedp.Title(&title)); err != nil {
		return "", fmt.Errorf("failed to get title: %w", err)
	}

	return title, nil
}

func (c *Controller) Click(selector string) error {
	ctx, cancel := c.createContext()
	defer cancel()

	return chromedp.Run(ctx, chromedp.Click(selector, chromedp.NodeVisible))
}

func (c *Controller) Type(selector, text string) error {
	ctx, cancel := c.createContext()
	defer cancel()

	return chromedp.Run(ctx,
		chromedp.Click(selector, chromedp.NodeVisible),
		chromedp.SendKeys(selector, text),
	)
}

func (c *Controller) Evaluate(expression string) (interface{}, error) {
	ctx, cancel := c.createContext()
	defer cancel()

	var result interface{}
	if err := chromedp.Run(ctx, chromedp.Evaluate(expression, &result)); err != nil {
		return nil, fmt.Errorf("failed to evaluate: %w", err)
	}

	return result, nil
}

func (c *Controller) GetHTML(selector string) (string, error) {
	ctx, cancel := c.createContext()
	defer cancel()

	var html string
	if err := chromedp.Run(ctx, chromedp.OuterHTML(selector, &html)); err != nil {
		return "", fmt.Errorf("failed to get HTML: %w", err)
	}

	return html, nil
}

func (c *Controller) WaitVisible(selector string) error {
	ctx, cancel := c.createContext()
	defer cancel()

	return chromedp.Run(ctx, chromedp.WaitVisible(selector))
}

func (c *Controller) Scroll(x, y int64) error {
	ctx, cancel := c.createContext()
	defer cancel()

	return chromedp.Run(ctx, chromedp.Evaluate(fmt.Sprintf("window.scrollTo(%d, %d)", x, y), nil))
}

func (c *Controller) GetPageInfo() (*PageInfo, error) {
	ctx, cancel := c.createContext()
	defer cancel()

	var url, title string
	if err := chromedp.Run(ctx,
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
	ctx, cancel := c.createContext()
	defer cancel()

	var buf []byte
	if err := chromedp.Run(ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		var err error
		buf, _, err = page.PrintToPDF().WithPrintBackground(true).Do(ctx)
		return err
	})); err != nil {
		return "", fmt.Errorf("failed to generate PDF: %w", err)
	}

	return base64.StdEncoding.EncodeToString(buf), nil
}

func (c *Controller) GetCookies() (string, error) {
	ctx, cancel := c.createContext()
	defer cancel()

	var cookies interface{}
	if err := chromedp.Run(ctx, chromedp.Evaluate("document.cookie", &cookies)); err != nil {
		return "", fmt.Errorf("failed to get cookies: %w", err)
	}

	result, _ := json.Marshal(cookies)
	return string(result), nil
}
