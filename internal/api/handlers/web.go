package handlers

import (
	"context"
	"net/http"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/deep-agent/sandbox/internal/services/web"
	"github.com/deep-agent/sandbox/model"
)

type WebHandler struct {
	fetcher  *web.Fetcher
	searcher *web.Searcher
}

func NewWebHandler(fetcher *web.Fetcher, searcher *web.Searcher) *WebHandler {
	return &WebHandler{
		fetcher:  fetcher,
		searcher: searcher,
	}
}

func (h *WebHandler) Fetch(ctx context.Context, c *app.RequestContext) {
	var req model.WebFetchRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Code:    400,
			Message: "invalid request: " + err.Error(),
		})
		return
	}

	result, err := h.fetcher.Fetch(ctx, req.URL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{
			Code:    500,
			Message: err.Error(),
		})
		return
	}

	if result.RedirectURL != "" {
		c.JSON(http.StatusOK, model.Response{
			Code:    302,
			Message: "Redirect detected",
			Data: map[string]string{
				"redirect_url": result.RedirectURL,
			},
		})
		return
	}

	c.JSON(http.StatusOK, model.Response{
		Code: 0,
		Data: model.WebFetchResult{
			Content: result.Content,
		},
	})
}

func (h *WebHandler) Search(ctx context.Context, c *app.RequestContext) {
	var req model.WebSearchRequest
	if err := c.BindAndValidate(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Code:    400,
			Message: "invalid request: " + err.Error(),
		})
		return
	}

	opts := web.SearchOptions{
		Query:          req.Query,
		AllowedDomains: req.AllowedDomains,
		BlockedDomains: req.BlockedDomains,
		NumResults:     req.NumResults,
		Language:       req.Language,
	}

	result, err := h.searcher.Search(ctx, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.Response{
			Code:    500,
			Message: err.Error(),
		})
		return
	}

	items := make([]model.WebSearchResultItem, 0, len(result.Results))
	for _, r := range result.Results {
		items = append(items, model.WebSearchResultItem{
			Title:   r.Title,
			URL:     r.URL,
			Snippet: r.Snippet,
		})
	}

	c.JSON(http.StatusOK, model.Response{
		Code: 0,
		Data: model.WebSearchResult{
			Results: items,
		},
	})
}
