package local

import (
	"context"
	"os/exec"
	"runtime"
	"strings"

	"github.com/deep-agent/sandbox/internal/services/bash"
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
	webFetcher   *web.Fetcher
	webSearcher  *web.Searcher
	sandboxCtx   *model.SandboxContext
	cwd          string
}

type Option func(*Client)

func WithSandboxContext(ctx *model.SandboxContext) Option {
	return func(c *Client) {
		c.sandboxCtx = ctx
	}
}

func WithCwd(cwd string) Option {
	return func(c *Client) {
		c.cwd = cwd
	}
}

func NewClient(opts ...Option) *Client {
	fileManager := filesystem.NewManager()
	home, err := fileManager.UserHomeDir()
	if err != nil {
		home = "/home"
	}
	c := &Client{
		bashExecutor: bash.NewExecutor(),
		fileManager:  fileManager,
		jsonlService: jsonl.NewService(),
		webFetcher:   web.NewFetcher(),
		webSearcher:  web.NewSearcher(),
		sandboxCtx: &model.SandboxContext{
			HomeDir:   home,
			OS:        runtime.GOOS,
			Arch:      runtime.GOARCH,
			OSVersion: getOSVersion(),
		},
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

func getOSVersion() string {
	out, err := exec.Command("uname", "-sr").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
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
