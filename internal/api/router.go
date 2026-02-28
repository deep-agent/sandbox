package api

import (
	"context"
	"fmt"
	"time"

	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/deep-agent/sandbox/internal/api/handlers"
	"github.com/deep-agent/sandbox/internal/api/middleware"
	"github.com/deep-agent/sandbox/internal/config"
	"github.com/deep-agent/sandbox/internal/services/bash"
	"github.com/deep-agent/sandbox/internal/services/browser"
	"github.com/deep-agent/sandbox/internal/services/filesystem"
	"github.com/deep-agent/sandbox/internal/services/web"
	"github.com/hertz-contrib/cors"
)

type Router struct {
	server          *server.Hertz
	cfg             *config.Config
	terminalHandler *handlers.TerminalHandler
}

func NewRouter(cfg *config.Config) *Router {
	h := server.Default(server.WithHostPorts(
		fmt.Sprintf(":%d", cfg.SandboxServerPort)))

	return &Router{
		server:          h,
		cfg:             cfg,
		terminalHandler: handlers.NewTerminalHandler(cfg.Workspace),
	}
}

func (r *Router) Setup() {
	bashExecutor := bash.NewExecutor()
	fileManager := filesystem.NewManager()
	browserController := browser.NewController(fmt.Sprintf("ws://localhost:%d", r.cfg.BrowserCDPPort))
	webFetcher := web.NewFetcher()
	webSearcher := web.NewSearcher()

	sandboxHandler := handlers.NewSandboxHandler(r.cfg)
	bashHandler := handlers.NewBashHandler(bashExecutor)
	fileHandler := handlers.NewFileHandler(fileManager)
	grepHandler := handlers.NewGrepHandler(fileManager)
	browserHandler := handlers.NewBrowserHandler(browserController)
	webHandler := handlers.NewWebHandler(webFetcher, webSearcher)
	swaggerHandler := handlers.NewSwaggerHandler()
	wsHandler := handlers.NewWSHandler()

	r.server.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Content-Type", "Authorization", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: false,
		MaxAge:           24 * time.Hour,
	}))
	r.server.Use(middleware.Logger())

	r.server.GET("/health", sandboxHandler.Health)
	r.server.GET("/docs", swaggerHandler.SwaggerUI)

	v1 := r.server.Group("/v1")
	v1.GET("/openapi.json", swaggerHandler.OpenAPISpec)
	v1.Use(middleware.Auth())
	{
		v1.GET("/sandbox", sandboxHandler.GetContext)

		bashGroup := v1.Group("/bash")
		{
			bashGroup.POST("/exec", bashHandler.ExecCommand)
			bashGroup.POST("/exec/stream", bashHandler.ExecCommandStream)
		}

		fileGroup := v1.Group("/file")
		{
			fileGroup.POST("/read", fileHandler.ReadFile)
			fileGroup.POST("/write", fileHandler.WriteFile)
			fileGroup.POST("/list", fileHandler.ListDir)
			fileGroup.POST("/delete", fileHandler.DeleteFile)
			fileGroup.POST("/move", fileHandler.MoveFile)
			fileGroup.POST("/copy", fileHandler.CopyFile)
			fileGroup.POST("/mkdir", fileHandler.MkDir)
			fileGroup.GET("/exists", fileHandler.Exists)
		}

		grepGroup := v1.Group("/grep")
		{
			grepGroup.POST("/search", grepHandler.Search)
		}

		browserGroup := v1.Group("/browser")
		{
			browserGroup.GET("/info", browserHandler.GetInfo)
			browserGroup.POST("/navigate", browserHandler.Navigate)
			browserGroup.POST("/screenshot", browserHandler.Screenshot)
			browserGroup.POST("/click", browserHandler.Click)
			browserGroup.POST("/type", browserHandler.Type)
			browserGroup.POST("/evaluate", browserHandler.Evaluate)
			browserGroup.GET("/url", browserHandler.GetCurrentURL)
			browserGroup.GET("/title", browserHandler.GetTitle)
			browserGroup.POST("/scroll", browserHandler.Scroll)
			browserGroup.POST("/html", browserHandler.GetHTML)
			browserGroup.POST("/wait", browserHandler.WaitVisible)
			browserGroup.GET("/page", browserHandler.GetPageInfo)
			browserGroup.POST("/pdf", browserHandler.PDF)
		}

		webGroup := v1.Group("/web")
		{
			webGroup.POST("/fetch", webHandler.Fetch)
			webGroup.POST("/search", webHandler.Search)
		}

		v1.GET("/terminal/ws", r.terminalHandler.HandleWebSocket)
		v1.GET("/ws", wsHandler.HandleWebSocket)
	}
}

func (r *Router) Run() error {
	return r.server.Run()
}

func (r *Router) Shutdown(ctx context.Context) error {
	return r.server.Shutdown(ctx)
}
