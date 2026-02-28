package middleware

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/deep-agent/sandbox/pkg/ctxutil"
	"github.com/deep-agent/sandbox/types/consts"
)

func Context() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		sessionID := string(c.Request.Header.Peek(consts.HeaderSessionID))
		cwd := string(c.Request.Header.Peek(consts.HeaderWorkspace))

		if sessionID != "" {
			ctx = ctxutil.WithSessionID(ctx, sessionID)
		}
		if cwd != "" {
			ctx = ctxutil.WithCwd(ctx, cwd)
		}

		c.Next(ctx)
	}
}
