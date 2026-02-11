package web

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type SearchResult struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	Snippet string `json:"snippet"`
}

type SearchResponse struct {
	Results []SearchResult
}

type mcpRequest struct {
	JSONRPC string         `json:"jsonrpc"`
	ID      int            `json:"id"`
	Method  string         `json:"method"`
	Params  mcpCallParams  `json:"params"`
}

type mcpCallParams struct {
	Name      string            `json:"name"`
	Arguments mcpSearchArgs     `json:"arguments"`
}

type mcpSearchArgs struct {
	Query                string `json:"query"`
	NumResults           int    `json:"numResults,omitempty"`
	Livecrawl            string `json:"livecrawl,omitempty"`
	Type                 string `json:"type,omitempty"`
	ContextMaxCharacters int    `json:"contextMaxCharacters,omitempty"`
}

type mcpResponse struct {
	JSONRPC string `json:"jsonrpc"`
	Result  struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
	} `json:"result"`
}

type exaSearchResult struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	Summary string `json:"summary"`
	Text    string `json:"text"`
}

type exaSearchResponse struct {
	Results []exaSearchResult `json:"results"`
}

type Searcher struct {
	client  *http.Client
	baseURL string
}

func NewSearcher() *Searcher {
	return &Searcher{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: "https://mcp.exa.ai",
	}
}

type SearchOptions struct {
	Query          string
	AllowedDomains []string
	BlockedDomains []string
	NumResults     int
	Language       string
}

func (s *Searcher) Search(ctx context.Context, opts SearchOptions) (*SearchResponse, error) {
	query := opts.Query

	for _, domain := range opts.AllowedDomains {
		query += fmt.Sprintf(" site:%s", domain)
	}

	for _, domain := range opts.BlockedDomains {
		query += fmt.Sprintf(" -site:%s", domain)
	}

	numResults := opts.NumResults
	if numResults <= 0 {
		numResults = 8
	}
	if numResults > 20 {
		numResults = 20
	}

	reqBody := mcpRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "tools/call",
		Params: mcpCallParams{
			Name: "web_search_exa",
			Arguments: mcpSearchArgs{
				Query:                query,
				NumResults:           numResults,
				Livecrawl:            "fallback",
				Type:                 "auto",
				ContextMaxCharacters: 10000,
			},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.baseURL+"/mcp", bytes.NewReader(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute search: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("search API error: %d %s", resp.StatusCode, resp.Status)
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")
			var mcpResp mcpResponse
			if err := json.Unmarshal([]byte(data), &mcpResp); err != nil {
				continue
			}

			if len(mcpResp.Result.Content) > 0 {
				text := mcpResp.Result.Content[0].Text
				return parseExaResults(text)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	return &SearchResponse{Results: []SearchResult{}}, nil
}

func parseExaResults(text string) (*SearchResponse, error) {
	var exaResp exaSearchResponse
	if err := json.Unmarshal([]byte(text), &exaResp); err != nil {
		results := []SearchResult{{
			Title:   "Search Results",
			URL:     "",
			Snippet: text,
		}}
		return &SearchResponse{Results: results}, nil
	}

	results := make([]SearchResult, 0, len(exaResp.Results))
	for _, r := range exaResp.Results {
		snippet := r.Summary
		if snippet == "" {
			snippet = r.Text
			if len(snippet) > 500 {
				snippet = snippet[:500] + "..."
			}
		}
		results = append(results, SearchResult{
			Title:   r.Title,
			URL:     r.URL,
			Snippet: snippet,
		})
	}

	return &SearchResponse{Results: results}, nil
}
