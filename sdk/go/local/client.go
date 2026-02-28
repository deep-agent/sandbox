package local

import (
	"runtime"

	"github.com/deep-agent/sandbox/internal/services/bash"
	"github.com/deep-agent/sandbox/internal/services/browser"
	"github.com/deep-agent/sandbox/internal/services/filesystem"
	"github.com/deep-agent/sandbox/model"
	sandbox "github.com/deep-agent/sandbox/sdk/go"
)

var _ sandbox.Sandbox = (*Client)(nil)

type Client struct {
	bashExecutor *bash.Executor
	fileManager  *filesystem.Manager
	browserCtrl  *browser.Controller
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
