package middleware

import (
	"context"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/deep-agent/sandbox/pkg/ctxutil"
	"github.com/deep-agent/sandbox/pkg/logger"
)

func Logger() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		start := time.Now()
		path := string(c.Request.URI().Path())
		method := string(c.Request.Method())
		query := string(c.Request.URI().QueryString())
		body := c.Request.Body()
		logID := ctxutil.GetLogIDFromCtx(ctx)

		format := "[HTTP] log_id=%s method=%s path=%s"
		args := []any{logID, method, path}
		if query != "" {
			format += " query=%q"
			args = append(args, query)
		}
		if len(body) > 0 {
			format += "\nreq_body=%s"
			args = append(args, truncate(string(body), 1024))
		}

		c.Next(ctx)

		latency := time.Since(start)
		status := c.Response.StatusCode()
		respBody := c.Response.Body()

		format += " status=%d latency=%v"
		args = append(args, status, latency)
		if len(respBody) > 0 {
			format += "\nresp_body=%s"
			args = append(args, truncate(string(respBody), 1024))
		}
		logger.Println("================================")
		logger.Printf(format, args...)
		logger.Println("================================")
	}
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	// Truncate at rune boundary to avoid splitting multi-byte UTF-8 characters.
	truncated := []rune(s)
	if len(truncated) <= maxLen {
		return s
	}
	return string(truncated[:maxLen]) + "...(truncated)"
}
