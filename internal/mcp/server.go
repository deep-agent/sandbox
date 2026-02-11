package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type ToolHandlerFunc = server.ToolHandlerFunc

type Server struct {
	mcpServer  *server.MCPServer
	httpServer *server.StreamableHTTPServer
	port       int
}

func NewServer(name, version string, port int) *Server {
	s := server.NewMCPServer(
		name,
		version,
		server.WithToolCapabilities(true),
	)

	return &Server{
		mcpServer: s,
		port:      port,
	}
}

func (s *Server) AddTool(tool mcp.Tool, handler server.ToolHandlerFunc) {
	wrappedHandler := loggingMiddleware(tool.Name, handler)
	s.mcpServer.AddTool(tool, wrappedHandler)
}

func loggingMiddleware(toolName string, next server.ToolHandlerFunc) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		start := time.Now()

		argsJSON, _ := json.Marshal(request.Params.Arguments)
		log.Printf("[%s] Request: %s", toolName, string(argsJSON))

		result, err := next(ctx, request)

		duration := time.Since(start)
		responseText := getResultText(result)

		if err != nil {
			log.Printf("[%s] Error after %v: %v \n=============", toolName, duration, err)
		} else if result != nil && result.IsError {
			log.Printf("[%s] Failed after %v\nResponse: %s \n=============", toolName, duration, responseText)
		} else {
			log.Printf("[%s] Success after %v\nResponse: %s \n=============", toolName, duration, responseText)
		}

		return result, err
	}
}

func getResultText(result *mcp.CallToolResult) string {
	if result == nil || len(result.Content) == 0 {
		return ""
	}
	for _, c := range result.Content {
		if textContent, ok := c.(mcp.TextContent); ok {
			text := textContent.Text
			// if len(text) > 200 {
			// 	return text[:200] + "..."
			// }
			return text
		}
	}
	return ""
}

func (s *Server) Start() error {
	s.httpServer = server.NewStreamableHTTPServer(s.mcpServer)

	addr := fmt.Sprintf(":%d", s.port)
	log.Printf("Starting MCP Streamable HTTP server on %s", addr)

	return s.httpServer.Start(addr)
}

func (s *Server) StartWithMux(mux *http.ServeMux, path string) {
	s.httpServer = server.NewStreamableHTTPServer(s.mcpServer,
		server.WithEndpointPath(path),
	)
	mux.Handle(path, s.httpServer)
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if s.httpServer != nil {
		s.httpServer.ServeHTTP(w, r)
	}
}

func (s *Server) Shutdown(ctx context.Context) error {
	if s.httpServer != nil {
		return s.httpServer.Shutdown(ctx)
	}
	return nil
}
