package http

import (
	"context"
	"fmt"
	"net/url"
	"time"

	emcp "github.com/cloudwego/eino-ext/components/tool/mcp"
	"github.com/cloudwego/eino/components/tool"
	sandbox "github.com/deep-agent/sandbox/sdk/go"
	"github.com/deep-agent/sandbox/types/consts"
	mcpclient "github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/client/transport"
	"github.com/mark3labs/mcp-go/mcp"
)

const (
	mcpClientName      = "sandbox-mcp-client"
	mcpClientVersion   = "1.0.0"
	mcpDefaultInitWait = 15 * time.Second
)

func (c *Client) MCPTools(ctx context.Context, opts ...sandbox.MCPOption) ([]tool.BaseTool, error) {
	cfg := &sandbox.MCPConfig{}
	for _, opt := range opts {
		opt(cfg)
	}

	mcpURL, err := url.JoinPath(c.baseURL, "mcp")
	if err != nil {
		return nil, fmt.Errorf("build mcp URL: %w", err)
	}

	headers := make(map[string]string, len(cfg.ExtraHeaders)+1)
	for k, v := range cfg.ExtraHeaders {
		headers[k] = v
	}
	if c.cwd != "" {
		headers[consts.HeaderWorkspace] = c.cwd
	}

	cli, err := mcpclient.NewStreamableHttpClient(mcpURL, transport.WithHTTPHeaders(headers))
	if err != nil {
		return nil, fmt.Errorf("create MCP client: %w", err)
	}

	wait := cfg.InitTimeout
	if wait <= 0 {
		wait = mcpDefaultInitWait
	}
	initCtx, cancel := context.WithTimeout(ctx, wait)
	defer cancel()

	if err := cli.Start(initCtx); err != nil {
		return nil, fmt.Errorf("start MCP client (server at %s): %w", mcpURL, err)
	}

	initReq := mcp.InitializeRequest{}
	initReq.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initReq.Params.ClientInfo = mcp.Implementation{
		Name:    mcpClientName,
		Version: mcpClientVersion,
	}
	if _, err := cli.Initialize(initCtx, initReq); err != nil {
		return nil, fmt.Errorf("initialize MCP client: %w", err)
	}

	tools, err := emcp.GetTools(ctx, &emcp.Config{
		Cli:          cli,
		ToolNameList: cfg.ToolNames,
	})
	if err != nil {
		return nil, fmt.Errorf("list MCP tools: %w", err)
	}
	return tools, nil
}
