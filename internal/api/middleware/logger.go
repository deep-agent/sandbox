package middleware

import (
	"context"
	"log"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/deep-agent/sandbox/types/consts"
)

func Logger() app.HandlerFunc {
	return func(ctx context.Context, c *app.RequestContext) {
		start := time.Now()
		path := string(c.Request.URI().Path())
		method := string(c.Request.Method())
		query := string(c.Request.URI().QueryString())
		body := c.Request.Body()
		sessionID := string(c.Request.Header.Peek(consts.HeaderSessionID))

		log.Printf("[REQ][SessionID:%s] %s %s query=%s body=%s", sessionID, method, path, query, truncate(string(body), 1024))

		c.Next(ctx)

		latency := time.Since(start)
		status := c.Response.StatusCode()
		respBody := c.Response.Body()

		log.Printf("[RESP][SessionID:%s] %s %s status=%d latency=%v body=%s", sessionID, method, path, status, latency, truncate(string(respBody), 1024))
		log.Printf("================================\n")
	}
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "...(truncated)"
}
