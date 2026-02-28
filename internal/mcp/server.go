package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/deep-agent/sandbox/pkg/ctxutil"
	"github.com/deep-agent/sandbox/types/consts"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type ToolHandlerFunc = server.ToolHandlerFunc

type Middleware func(server.ToolHandlerFunc) server.ToolHandlerFunc

type Server struct {
	mcpServer   *server.MCPServer
	httpServer  *server.StreamableHTTPServer
	port        int
	middlewares []Middleware
}

func NewServer(name, version string, port int) *Server {
	mcpServer := server.NewMCPServer(
		name,
		version,
		server.WithToolCapabilities(true),
	)

	s := &Server{
		mcpServer: mcpServer,
		port:      port,
	}

	s.Use(loggingMiddleware)
	s.Use(contextMiddleware)

	return s
}

func (s *Server) Use(middleware Middleware) {
	s.middlewares = append(s.middlewares, middleware)
}

func (s *Server) AddTool(tool mcp.Tool, handler server.ToolHandlerFunc) {
	wrapped := s.wrapHandler(handler)
	s.mcpServer.AddTool(tool, wrapped)
}

func (s *Server) wrapHandler(handler server.ToolHandlerFunc) server.ToolHandlerFunc {
	for i := len(s.middlewares) - 1; i >= 0; i-- {
		handler = s.middlewares[i](handler)
	}
	return handler
}

func contextMiddleware(next server.ToolHandlerFunc) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		sessionID := request.Header.Get(consts.HeaderSessionID)
		cwd := request.Header.Get(consts.HeaderWorkspace)
		log.Printf("[contextMiddleware] sessionID=%s, cwd=%s", sessionID, cwd)
		if sessionID != "" {
			ctx = ctxutil.WithSessionID(ctx, sessionID)
		}
		if cwd != "" {
			ctx = ctxutil.WithCwd(ctx, cwd)
		}
		return next(ctx, request)
	}
}

func loggingMiddleware(next server.ToolHandlerFunc) server.ToolHandlerFunc {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		start := time.Now()

		toolName := request.Params.Name
		sessionID := ctxutil.GetSessionIDFromCtx(ctx)
		argsJSON, _ := json.Marshal(request.Params.Arguments)
		log.Printf("[%s][SessionID:%s] Request: %s", toolName, sessionID, string(argsJSON))

		result, err := next(ctx, request)

		duration := time.Since(start)
		responseText := getResultText(result)

		if err != nil {
			log.Printf("[%s][SessionID:%s] Error after %v: %v \n=============", toolName, sessionID, duration, err)
		} else if result != nil && result.IsError {
			log.Printf("[%s][SessionID:%s] Failed after %v\nResponse: %s \n=============", toolName, sessionID, duration, responseText)
		} else {
			log.Printf("[%s][SessionID:%s] Success after %v\nResponse: %s \n=============", toolName, sessionID, duration, responseText)
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
