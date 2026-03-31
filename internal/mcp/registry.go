package mcp

import (
	"github.com/deep-agent/sandbox/internal/mcp/tools"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type Registry struct{}

func NewRegistry() *Registry {
	return &Registry{}
}

func (r *Registry) RegisterAll(addTool func(tool mcp.Tool, handler server.ToolHandlerFunc)) {
	addTool(tools.BashToolDef(), tools.BashHandler())

	addTool(tools.GlobToolDef(), tools.GlobHandler())
	addTool(tools.GrepToolDef(), tools.GrepHandler())
	addTool(tools.ReadToolDef(), tools.ReadHandler())
	addTool(tools.WriteToolDef(), tools.WriteHandler())
	addTool(tools.EditToolDef(), tools.EditHandler())

	addTool(tools.WebFetchToolDef(), tools.WebFetchHandler())
	addTool(tools.WebSearchToolDef(), tools.WebSearchHandler())
}
