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

		format := "[HTTP] sid=%s method=%s path=%s"
		args := []any{sessionID, method, path}
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
		log.Println("================================")
		log.Printf(format, args...)
		log.Println("================================")
	}
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "...(truncated)"
}
