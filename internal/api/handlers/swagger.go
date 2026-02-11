package handlers

import (
	"context"
	_ "embed"
	"net/http"

	"github.com/cloudwego/hertz/pkg/app"
)

//go:embed swagger-ui.html
var swaggerUIHTML []byte

//go:embed openapi.json
var openapiJSON []byte

type SwaggerHandler struct{}

func NewSwaggerHandler() *SwaggerHandler {
	return &SwaggerHandler{}
}

func (h *SwaggerHandler) SwaggerUI(ctx context.Context, c *app.RequestContext) {
	c.SetContentType("text/html; charset=utf-8")
	c.SetStatusCode(http.StatusOK)
	c.Write(swaggerUIHTML)
}

func (h *SwaggerHandler) OpenAPISpec(ctx context.Context, c *app.RequestContext) {
	c.SetContentType("application/json")
	c.SetStatusCode(http.StatusOK)
	c.Write(openapiJSON)
}
