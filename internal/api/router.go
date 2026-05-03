package api

import (
	"context"
	"fmt"
	"time"

	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/deep-agent/sandbox/internal/api/handlers"
	"github.com/deep-agent/sandbox/internal/api/middleware"
	"github.com/deep-agent/sandbox/internal/config"
	"github.com/deep-agent/sandbox/internal/services/filesystem"
	"github.com/deep-agent/sandbox/internal/services/jsonl"
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
	fileManager := filesystem.NewManager()
	webFetcher := web.NewFetcher()
	webSearcher := web.NewSearcher()

	sandboxHandler := handlers.NewSandboxHandler(r.cfg)
	bashHandler := handlers.NewBashHandler()
	fileHandler := handlers.NewFileHandler(fileManager)
	grepHandler := handlers.NewGrepHandler(fileManager)
	webHandler := handlers.NewWebHandler(webFetcher, webSearcher)
	jsonlService := jsonl.NewService()
	jsonlHandler := handlers.NewJSONLHandler(jsonlService)
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
	r.server.Use(middleware.Context())

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
			fileGroup.POST("/glob", fileHandler.Glob)
			fileGroup.POST("/delete", fileHandler.DeleteFile)
			fileGroup.POST("/move", fileHandler.MoveFile)
			fileGroup.POST("/copy", fileHandler.CopyFile)
			fileGroup.POST("/mkdir", fileHandler.MkDir)
			fileGroup.GET("/exists", fileHandler.Exists)
			fileGroup.POST("/upload", fileHandler.Upload)
			fileGroup.GET("/download", fileHandler.Download)
			fileGroup.POST("/create-temp", fileHandler.CreateTemp)
			fileGroup.POST("/eval-symlinks", fileHandler.EvalSymlinks)
			fileGroup.POST("/append", fileHandler.Append)
			fileGroup.POST("/stat", fileHandler.Stat)
			fileGroup.GET("/temp-dir", fileHandler.TempDir)
			fileGroup.GET("/home-dir", fileHandler.UserHomeDir)
		}

		grepGroup := v1.Group("/grep")
		{
			grepGroup.POST("/search", grepHandler.Search)
		}

		webGroup := v1.Group("/web")
		{
			webGroup.POST("/fetch", webHandler.Fetch)
			webGroup.POST("/search", webHandler.Search)
		}

		jsonlGroup := v1.Group("/jsonl")
		{
			jsonlGroup.POST("/count", jsonlHandler.CountLines)
			jsonlGroup.POST("/read", jsonlHandler.ReadLines)
			jsonlGroup.POST("/append", jsonlHandler.AppendLine)
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
