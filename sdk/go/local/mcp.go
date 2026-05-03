package local

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/bytedance/sonic"
	etool "github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	sbxmcp "github.com/deep-agent/sandbox/internal/mcp"
	"github.com/deep-agent/sandbox/pkg/ctxutil"
	sandbox "github.com/deep-agent/sandbox/sdk/go"
	"github.com/eino-contrib/jsonschema"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func (c *Client) MCPTools(_ context.Context, opts ...sandbox.MCPOption) ([]etool.BaseTool, error) {
	cfg := &sandbox.MCPConfig{}
	for _, opt := range opts {
		opt(cfg)
	}

	cwd := c.cwd
	if cwd == "" && c.sandboxCtx != nil {
		cwd = c.sandboxCtx.HomeDir
	}

	want := make(map[string]struct{}, len(cfg.ToolNames))
	for _, n := range cfg.ToolNames {
		want[n] = struct{}{}
	}

	var (
		tools   []etool.BaseTool
		collect = func(def mcp.Tool, handler server.ToolHandlerFunc) {
			if len(want) > 0 {
				if _, ok := want[def.Name]; !ok {
					return
				}
			}
			info, err := toolInfoFromMCP(def)
			if err != nil {
				panic(fmt.Errorf("mcp: build tool info for %q: %w", def.Name, err))
			}
			tools = append(tools, &mcpTool{
				def:     def,
				info:    info,
				handler: handler,
				cwd:     cwd,
			})
		}
	)

	sbxmcp.NewRegistry().RegisterAll(collect)
	return tools, nil
}

func toolInfoFromMCP(t mcp.Tool) (*schema.ToolInfo, error) {
	raw, err := sonic.Marshal(t.InputSchema)
	if err != nil {
		return nil, fmt.Errorf("marshal input schema: %w", err)
	}
	js := &jsonschema.Schema{}
	if err := sonic.Unmarshal(raw, js); err != nil {
		return nil, fmt.Errorf("unmarshal input schema: %w", err)
	}
	return &schema.ToolInfo{
		Name:        t.Name,
		Desc:        t.Description,
		ParamsOneOf: schema.NewParamsOneOfByJSONSchema(js),
	}, nil
}

type mcpTool struct {
	def     mcp.Tool
	info    *schema.ToolInfo
	handler server.ToolHandlerFunc
	cwd     string
}

func (t *mcpTool) Info(_ context.Context) (*schema.ToolInfo, error) {
	return t.info, nil
}

func (t *mcpTool) InvokableRun(ctx context.Context, argumentsInJSON string, _ ...etool.Option) (string, error) {
	args := map[string]any{}
	if argumentsInJSON != "" {
		if err := json.Unmarshal([]byte(argumentsInJSON), &args); err != nil {
			return "", fmt.Errorf("unmarshal arguments: %w", err)
		}
	}

	req := mcp.CallToolRequest{}
	req.Request.Method = "tools/call"
	req.Params.Name = t.def.Name
	req.Params.Arguments = args

	if t.cwd != "" {
		ctx = ctxutil.WithCwd(ctx, t.cwd)
	}

	result, err := t.handler(ctx, req)
	if err != nil {
		return "", fmt.Errorf("call %s: %w", t.def.Name, err)
	}

	out, err := sonic.MarshalString(result)
	if err != nil {
		return "", fmt.Errorf("marshal result: %w", err)
	}
	if result != nil && result.IsError {
		return "", fmt.Errorf("%s returned error: %s", t.def.Name, out)
	}
	return out, nil
}
