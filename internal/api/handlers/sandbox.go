package handlers

import (
	"context"
	"net/http"
	"runtime"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/deep-agent/sandbox/internal/config"
	"github.com/deep-agent/sandbox/model"
	"github.com/deep-agent/sandbox/pkg/session"
)

type SandboxHandler struct {
	cfg *config.Config
}

func NewSandboxHandler(cfg *config.Config) *SandboxHandler {
	return &SandboxHandler{cfg: cfg}
}

func (h *SandboxHandler) GetContext(ctx context.Context, c *app.RequestContext) {
	sessionID := session.GetSessionIDFromHertz(&c.Request.Header)
	sandboxCtx := model.SandboxContext{
		Workspace: session.GetSessionWorkspace(h.cfg.Workspace, sessionID),
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
	}

	c.JSON(http.StatusOK, model.Response{
		Code: 0,
		Data: sandboxCtx,
	})
}

func (h *SandboxHandler) Health(ctx context.Context, c *app.RequestContext) {
	c.JSON(http.StatusOK, map[string]interface{}{
		"status": "ok",
	})
}
