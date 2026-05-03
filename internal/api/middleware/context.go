package middleware

import (
	"context"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/deep-agent/sandbox/pkg/ctxutil"
	"github.com/deep-agent/sandbox/types/consts"
)

func Context() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		logID := string(c.Request.Header.Peek(consts.HeaderLogID))
		cwd := string(c.Request.Header.Peek(consts.HeaderWorkspace))

		if logID == "" {
			logID = ctxutil.NewLogID()
		}
		ctx = ctxutil.WithLogID(ctx, logID)

		if cwd != "" {
			ctx = ctxutil.WithCwd(ctx, cwd)
		}

		c.Next(ctx)
	}
}
