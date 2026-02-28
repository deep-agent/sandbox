package http

import (
	"context"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/client/transport"
)

func (c *Client) NewStreamableHttpClient(ctx context.Context, mcpURL string) (*client.Client, error) {
	return client.NewStreamableHttpClient(mcpURL,
		transport.WithHTTPHeaders(map[string]string{
			"X-Session-ID": c.sessionID,
			"X-Workspace":  c.cwd,
		}),
	)
}
