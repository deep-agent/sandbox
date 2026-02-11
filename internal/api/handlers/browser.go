package handlers

import (
	"context"
	"net/http"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/deep-agent/sandbox/internal/services/browser"
	"github.com/deep-agent/sandbox/model"
)

type BrowserHandler struct {
	controller *browser.Controller
}

func NewBrowserHandler(controller *browser.Controller) *BrowserHandler {
	return &BrowserHandler{controller: controller}
}

func (h *BrowserHandler) GetInfo(ctx context.Context, c *app.RequestContext) {
	info, err := h.controller.GetInfo()
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{
			Code:    500,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{
		Code: 0,
		Data: info,
	})
}

func (h *BrowserHandler) Navigate(ctx context.Context, c *app.RequestContext) {
	var req model.BrowserNavigateRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Code:    400,
			Message: "invalid request: " + err.Error(),
		})
		return
	}

	if err := h.controller.Navigate(req.URL); err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{
			Code:    500,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{
		Code:    0,
		Message: "success",
	})
}

func (h *BrowserHandler) Screenshot(ctx context.Context, c *app.RequestContext) {
	var req model.BrowserScreenshotRequest
	c.BindAndValidate(&req)

	opts := &browser.ScreenshotOptions{
		Format:  req.Format,
		Quality: req.Quality,
		Full:    req.Full,
	}

	screenshot, err := h.controller.Screenshot(opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{
			Code:    500,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{
		Code: 0,
		Data: model.BrowserScreenshotResult{Screenshot: screenshot},
	})
}

func (h *BrowserHandler) Click(ctx context.Context, c *app.RequestContext) {
	var req model.BrowserClickRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Code:    400,
			Message: "invalid request: " + err.Error(),
		})
		return
	}

	if err := h.controller.Click(req.Selector); err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{
			Code:    500,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{
		Code:    0,
		Message: "success",
	})
}

func (h *BrowserHandler) Type(ctx context.Context, c *app.RequestContext) {
	var req model.BrowserTypeRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Code:    400,
			Message: "invalid request: " + err.Error(),
		})
		return
	}

	if err := h.controller.Type(req.Selector, req.Text); err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{
			Code:    500,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{
		Code:    0,
		Message: "success",
	})
}

func (h *BrowserHandler) Evaluate(ctx context.Context, c *app.RequestContext) {
	var req model.BrowserEvaluateRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Code:    400,
			Message: "invalid request: " + err.Error(),
		})
		return
	}

	result, err := h.controller.Evaluate(req.Expression)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{
			Code:    500,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{
		Code: 0,
		Data: model.BrowserEvaluateResult{Result: result},
	})
}

func (h *BrowserHandler) GetCurrentURL(ctx context.Context, c *app.RequestContext) {
	url, err := h.controller.GetCurrentURL()
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{
			Code:    500,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{
		Code: 0,
		Data: model.BrowserURLResult{URL: url},
	})
}

func (h *BrowserHandler) GetTitle(ctx context.Context, c *app.RequestContext) {
	title, err := h.controller.GetTitle()
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{
			Code:    500,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{
		Code: 0,
		Data: model.BrowserTitleResult{Title: title},
	})
}

func (h *BrowserHandler) Scroll(ctx context.Context, c *app.RequestContext) {
	var req model.BrowserScrollRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Code:    400,
			Message: "invalid request: " + err.Error(),
		})
		return
	}

	if err := h.controller.Scroll(req.X, req.Y); err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{
			Code:    500,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{
		Code:    0,
		Message: "success",
	})
}

func (h *BrowserHandler) GetHTML(ctx context.Context, c *app.RequestContext) {
	var req model.BrowserGetHTMLRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Code:    400,
			Message: "invalid request: " + err.Error(),
		})
		return
	}

	html, err := h.controller.GetHTML(req.Selector)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{
			Code:    500,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{
		Code: 0,
		Data: model.BrowserGetHTMLResult{HTML: html},
	})
}

func (h *BrowserHandler) WaitVisible(ctx context.Context, c *app.RequestContext) {
	var req model.BrowserWaitVisibleRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Code:    400,
			Message: "invalid request: " + err.Error(),
		})
		return
	}

	if err := h.controller.WaitVisible(req.Selector); err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{
			Code:    500,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{
		Code:    0,
		Message: "success",
	})
}

func (h *BrowserHandler) GetPageInfo(ctx context.Context, c *app.RequestContext) {
	info, err := h.controller.GetPageInfo()
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{
			Code:    500,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{
		Code: 0,
		Data: model.BrowserPageInfo{
			URL:    info.URL,
			Title:  info.Title,
			Width:  info.Width,
			Height: info.Height,
		},
	})
}

func (h *BrowserHandler) PDF(ctx context.Context, c *app.RequestContext) {
	pdf, err := h.controller.PDF()
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{
			Code:    500,
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{
		Code: 0,
		Data: model.BrowserPDFResult{PDF: pdf},
	})
}
