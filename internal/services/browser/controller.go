package browser

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/chromedp/cdproto/input"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/cdproto/target"
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

type TabInfo struct {
	TabID string `json:"tab_id"`
	URL   string `json:"url"`
	Title string `json:"title"`
}

type InteractiveElement struct {
	Tag         string `json:"tag"`
	Text        string `json:"text,omitempty"`
	Selector    string `json:"selector,omitempty"`
	Placeholder string `json:"placeholder,omitempty"`
	Href        string `json:"href,omitempty"`
}

type BrowserState struct {
	URL                 string               `json:"url"`
	Title               string               `json:"title"`
	Tabs                []TabInfo            `json:"tabs"`
	Viewport            *ViewportInfo        `json:"viewport,omitempty"`
	InteractiveElements []InteractiveElement `json:"interactive_elements"`
	Screenshot          string               `json:"screenshot,omitempty"`
}

type ViewportInfo struct {
	Width  int `json:"width"`
	Height int `json:"height"`
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

func (c *Controller) NavigateNewTab(targetURL string) error {
	return c.run(chromedp.ActionFunc(func(ctx context.Context) error {
		_, err := target.CreateTarget(targetURL).Do(ctx)
		return err
	}))
}

func (c *Controller) ClickCoordinate(x, y int) error {
	return c.run(chromedp.ActionFunc(func(ctx context.Context) error {
		fx := float64(x)
		fy := float64(y)
		if err := input.DispatchMouseEvent(input.MousePressed, fx, fy).
			WithButton(input.Left).
			WithClickCount(1).
			Do(ctx); err != nil {
			return err
		}
		return input.DispatchMouseEvent(input.MouseReleased, fx, fy).
			WithButton(input.Left).
			WithClickCount(1).
			Do(ctx)
	}))
}

func (c *Controller) ScrollByDirection(direction string) error {
	var deltaY int
	if direction == "up" {
		deltaY = -500
	} else {
		deltaY = 500
	}
	return c.run(chromedp.Evaluate(fmt.Sprintf("window.scrollBy(0, %d)", deltaY), nil))
}

func (c *Controller) GoBack() error {
	return c.run(chromedp.Evaluate("window.history.back()", nil))
}

func (c *Controller) GetState(includeScreenshot bool) (*BrowserState, error) {
	var pageURL, title string
	var elementsJSON string

	js := `(function() {
		var selectors = ['a', 'button', 'input', 'select', 'textarea', '[role="button"]', '[onclick]'];
		var seen = new Set();
		var results = [];
		selectors.forEach(function(sel) {
			document.querySelectorAll(sel).forEach(function(el) {
				if (seen.has(el)) return;
				seen.add(el);
				var rect = el.getBoundingClientRect();
				if (rect.width === 0 && rect.height === 0) return;
				var text = (el.textContent || '').trim().substring(0, 100);
				var entry = {tag: el.tagName.toLowerCase(), text: text};
				if (el.id) entry.selector = '#' + el.id;
				else if (el.name) entry.selector = el.tagName.toLowerCase() + '[name="' + el.name + '"]';
				else if (el.className && typeof el.className === 'string') entry.selector = el.tagName.toLowerCase() + '.' + el.className.trim().split(/\s+/).join('.');
				if (el.placeholder) entry.placeholder = el.placeholder;
				if (el.href) entry.href = el.href;
				results.push(entry);
			});
		});
		return JSON.stringify(results);
	})()`

	if err := c.run(
		chromedp.Location(&pageURL),
		chromedp.Title(&title),
		chromedp.Evaluate(js, &elementsJSON),
	); err != nil {
		return nil, fmt.Errorf("failed to get browser state: %w", err)
	}

	state := &BrowserState{
		URL:   pageURL,
		Title: title,
	}

	if elementsJSON != "" {
		var elements []InteractiveElement
		if err := json.Unmarshal([]byte(elementsJSON), &elements); err == nil {
			state.InteractiveElements = elements
		}
	}
	if state.InteractiveElements == nil {
		state.InteractiveElements = []InteractiveElement{}
	}

	if includeScreenshot {
		b64, err := c.Screenshot(nil)
		if err == nil {
			state.Screenshot = b64
		}
	}

	// Get tabs
	tabs, err := c.ListTabs()
	if err == nil {
		state.Tabs = tabs
	}

	return state, nil
}

func (c *Controller) ListTabs() ([]TabInfo, error) {
	var tabs []TabInfo
	err := c.run(chromedp.ActionFunc(func(ctx context.Context) error {
		targets, err := target.GetTargets().Do(ctx)
		if err != nil {
			return err
		}
		for _, t := range targets {
			if t.Type != "page" {
				continue
			}
			id := string(t.TargetID)
			tabID := id
			if len(id) > 4 {
				tabID = id[len(id)-4:]
			}
			tabs = append(tabs, TabInfo{
				TabID: tabID,
				URL:   t.URL,
				Title: t.Title,
			})
		}
		return nil
	}))
	if err != nil {
		return nil, fmt.Errorf("failed to list tabs: %w", err)
	}
	if tabs == nil {
		tabs = []TabInfo{}
	}
	return tabs, nil
}

func (c *Controller) findTargetByTabID(tabID string) (target.ID, error) {
	var found target.ID
	err := c.run(chromedp.ActionFunc(func(ctx context.Context) error {
		targets, err := target.GetTargets().Do(ctx)
		if err != nil {
			return err
		}
		for _, t := range targets {
			if t.Type != "page" {
				continue
			}
			id := string(t.TargetID)
			suffix := id
			if len(id) > 4 {
				suffix = id[len(id)-4:]
			}
			if suffix == tabID {
				found = t.TargetID
				return nil
			}
		}
		return fmt.Errorf("tab %q not found", tabID)
	}))
	return found, err
}

func (c *Controller) SwitchTab(tabID string) error {
	targetID, err := c.findTargetByTabID(tabID)
	if err != nil {
		return err
	}
	return c.run(chromedp.ActionFunc(func(ctx context.Context) error {
		return target.ActivateTarget(targetID).Do(ctx)
	}))
}

func (c *Controller) CloseTab(tabID string) error {
	targetID, err := c.findTargetByTabID(tabID)
	if err != nil {
		return err
	}
	// Create a new chromedp context attached to the specific target,
	// then cancel it to close the tab cleanly.
	newCtx, cancel := chromedp.NewContext(c.allocCtx, chromedp.WithTargetID(targetID))
	defer cancel()
	// Ensure the context is initialized (attaches to the target)
	if err := chromedp.Run(newCtx); err != nil {
		// If we can't attach, the target may already be gone
		return nil
	}
	// Cancel the context, which tells chromedp to close the target
	cancel()
	return nil
}

// findFullTargetID resolves a short tab ID (last 4 chars) to the full target ID string.
// Used by CloseTab's HTTP approach.
func (c *Controller) findFullTargetID(tabID string) (string, error) {
	var fullID string
	err := c.run(chromedp.ActionFunc(func(ctx context.Context) error {
		targets, err := target.GetTargets().Do(ctx)
		if err != nil {
			return err
		}
		for _, t := range targets {
			if t.Type != "page" {
				continue
			}
			id := string(t.TargetID)
			suffix := id
			if len(id) > 4 {
				suffix = id[len(id)-4:]
			}
			if suffix == tabID {
				fullID = id
				return nil
			}
		}
		return fmt.Errorf("tab %q not found", tabID)
	}))
	return fullID, err
}
