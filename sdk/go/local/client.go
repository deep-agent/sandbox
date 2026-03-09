package local

import (
	"context"
	"runtime"

	"github.com/deep-agent/sandbox/internal/services/bash"
	"github.com/deep-agent/sandbox/internal/services/browser"
	"github.com/deep-agent/sandbox/internal/services/filesystem"
	"github.com/deep-agent/sandbox/internal/services/jsonl"
	"github.com/deep-agent/sandbox/internal/services/web"
	sandbox "github.com/deep-agent/sandbox/sdk/go"
	"github.com/deep-agent/sandbox/types/model"
)

var _ sandbox.Sandbox = (*Client)(nil)

type Client struct {
	bashExecutor *bash.Executor
	fileManager  *filesystem.Manager
	jsonlService *jsonl.Service
	browserCtrl  *browser.Controller
	webFetcher   *web.Fetcher
	webSearcher  *web.Searcher
	sandboxCtx   *model.SandboxContext
}

type Option func(*Client)

func WithBrowserCDP(cdpURL string) Option {
	return func(c *Client) {
		c.browserCtrl = browser.NewController(cdpURL)
	}
}

func WithSandboxContext(ctx *model.SandboxContext) Option {
	return func(c *Client) {
		c.sandboxCtx = ctx
	}
}

func NewClient(workDir string, opts ...Option) *Client {
	c := &Client{
		bashExecutor: bash.NewExecutor(),
		fileManager:  filesystem.NewManager(),
		jsonlService: jsonl.NewService(),
		webFetcher:   web.NewFetcher(),
		webSearcher:  web.NewSearcher(),
		sandboxCtx: &model.SandboxContext{
			Workspace: workDir,
			OS:        runtime.GOOS,
			Arch:      runtime.GOARCH,
		},
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

func (c *Client) GetContext() (*model.SandboxContext, error) {
	return c.sandboxCtx, nil
}

func (c *Client) WebFetch(req *model.WebFetchRequest) (*model.WebFetchResult, error) {
	result, err := c.webFetcher.Fetch(context.Background(), req.URL)
	if err != nil {
		return nil, err
	}
	return &model.WebFetchResult{Content: result.Content}, nil
}

func (c *Client) WebSearch(req *model.WebSearchRequest) (*model.WebSearchResult, error) {
	opts := web.SearchOptions{
		Query:          req.Query,
		AllowedDomains: req.AllowedDomains,
		BlockedDomains: req.BlockedDomains,
		NumResults:     req.NumResults,
		Language:       req.Language,
	}
	resp, err := c.webSearcher.Search(context.Background(), opts)
	if err != nil {
		return nil, err
	}

	results := make([]model.WebSearchResultItem, len(resp.Results))
	for i, r := range resp.Results {
		results[i] = model.WebSearchResultItem{
			Title:   r.Title,
			URL:     r.URL,
			Snippet: r.Snippet,
		}
	}
	return &model.WebSearchResult{Results: results}, nil
}
