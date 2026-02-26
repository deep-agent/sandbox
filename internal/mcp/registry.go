package mcp

import (
	"github.com/deep-agent/sandbox/internal/mcp/tools"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type ToolConfig struct {
	HomeDir string
	CDPURL  string
}

type Registry struct {
	config ToolConfig
}

func NewRegistry(cfg ToolConfig) *Registry {
	return &Registry{
		config: cfg,
	}
}

func (r *Registry) RegisterAll(addTool func(tool mcp.Tool, handler server.ToolHandlerFunc)) {
	addTool(tools.BashToolDef(), tools.BashHandler(r.config.HomeDir))

	addTool(tools.GlobToolDef(), tools.GlobHandler(r.config.HomeDir))
	addTool(tools.GrepToolDef(), tools.GrepHandler(r.config.HomeDir))
	addTool(tools.ReadToolDef(), tools.ReadHandler())
	addTool(tools.WriteToolDef(), tools.WriteHandler())
	addTool(tools.EditToolDef(), tools.EditHandler())

	addTool(tools.BrowserNavigateToolDef(), tools.BrowserNavigateHandler(r.config.CDPURL))
	addTool(tools.BrowserScreenshotToolDef(), tools.BrowserScreenshotHandler(r.config.CDPURL))
	addTool(tools.BrowserClickToolDef(), tools.BrowserClickHandler(r.config.CDPURL))
	addTool(tools.BrowserTypeToolDef(), tools.BrowserTypeHandler(r.config.CDPURL))
	addTool(tools.BrowserGetURLToolDef(), tools.BrowserGetURLHandler(r.config.CDPURL))
	addTool(tools.BrowserGetTitleToolDef(), tools.BrowserGetTitleHandler(r.config.CDPURL))
	addTool(tools.BrowserGetHTMLToolDef(), tools.BrowserGetHTMLHandler(r.config.CDPURL))
	addTool(tools.BrowserEvaluateToolDef(), tools.BrowserEvaluateHandler(r.config.CDPURL))
	addTool(tools.BrowserScrollToolDef(), tools.BrowserScrollHandler(r.config.CDPURL))
	addTool(tools.BrowserWaitVisibleToolDef(), tools.BrowserWaitVisibleHandler(r.config.CDPURL))
	addTool(tools.BrowserGetPageInfoToolDef(), tools.BrowserGetPageInfoHandler(r.config.CDPURL))
	addTool(tools.BrowserPDFToolDef(), tools.BrowserPDFHandler(r.config.CDPURL))

	addTool(tools.WebFetchToolDef(), tools.WebFetchHandler())
	addTool(tools.WebSearchToolDef(), tools.WebSearchHandler())
}
