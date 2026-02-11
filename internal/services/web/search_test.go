package web

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewSearcher(t *testing.T) {
	s := NewSearcher()
	if s == nil {
		t.Fatal("expected non-nil Searcher")
	}
	if s.client == nil {
		t.Error("expected non-nil http client")
	}
	if s.baseURL != "https://mcp.exa.ai" {
		t.Errorf("expected baseURL to be 'https://mcp.exa.ai', got %q", s.baseURL)
	}
	if s.client.Timeout != 30*time.Second {
		t.Errorf("expected timeout to be 30 seconds, got %v", s.client.Timeout)
	}
}

func TestParseExaResults_ValidJSON(t *testing.T) {
	exaResp := exaSearchResponse{
		Results: []exaSearchResult{
			{
				Title:   "Test Title 1",
				URL:     "https://example.com/1",
				Summary: "Test summary 1",
				Text:    "Full text 1",
			},
			{
				Title:   "Test Title 2",
				URL:     "https://example.com/2",
				Summary: "",
				Text:    "Short text",
			},
		},
	}

	jsonData, err := json.Marshal(exaResp)
	if err != nil {
		t.Fatalf("failed to marshal test data: %v", err)
	}

	result, err := parseExaResults(string(jsonData))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(result.Results))
	}

	if result.Results[0].Title != "Test Title 1" {
		t.Errorf("expected title 'Test Title 1', got %q", result.Results[0].Title)
	}
	if result.Results[0].URL != "https://example.com/1" {
		t.Errorf("expected URL 'https://example.com/1', got %q", result.Results[0].URL)
	}
	if result.Results[0].Snippet != "Test summary 1" {
		t.Errorf("expected snippet 'Test summary 1', got %q", result.Results[0].Snippet)
	}

	if result.Results[1].Snippet != "Short text" {
		t.Errorf("expected snippet 'Short text' when summary is empty, got %q", result.Results[1].Snippet)
	}
}

func TestParseExaResults_InvalidJSON(t *testing.T) {
	result, err := parseExaResults("not valid json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Results) != 1 {
		t.Fatalf("expected 1 result for invalid JSON, got %d", len(result.Results))
	}

	if result.Results[0].Title != "Search Results" {
		t.Errorf("expected title 'Search Results', got %q", result.Results[0].Title)
	}
	if result.Results[0].Snippet != "not valid json" {
		t.Errorf("expected snippet to contain raw text, got %q", result.Results[0].Snippet)
	}
}

func TestParseExaResults_LongText(t *testing.T) {
	longText := ""
	for i := 0; i < 600; i++ {
		longText += "a"
	}

	exaResp := exaSearchResponse{
		Results: []exaSearchResult{
			{
				Title:   "Long Text Result",
				URL:     "https://example.com/long",
				Summary: "",
				Text:    longText,
			},
		},
	}

	jsonData, err := json.Marshal(exaResp)
	if err != nil {
		t.Fatalf("failed to marshal test data: %v", err)
	}

	result, err := parseExaResults(string(jsonData))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Results[0].Snippet) != 503 {
		t.Errorf("expected snippet length 503 (500 + '...'), got %d", len(result.Results[0].Snippet))
	}
}

func TestParseExaResults_EmptyResults(t *testing.T) {
	exaResp := exaSearchResponse{
		Results: []exaSearchResult{},
	}

	jsonData, err := json.Marshal(exaResp)
	if err != nil {
		t.Fatalf("failed to marshal test data: %v", err)
	}

	result, err := parseExaResults(string(jsonData))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Results) != 0 {
		t.Errorf("expected 0 results, got %d", len(result.Results))
	}
}

func TestSearchOptions_DefaultValues(t *testing.T) {
	opts := SearchOptions{
		Query: "test query",
	}

	if opts.Query != "test query" {
		t.Errorf("expected query 'test query', got %q", opts.Query)
	}
	if opts.NumResults != 0 {
		t.Errorf("expected default NumResults 0, got %d", opts.NumResults)
	}
	if len(opts.AllowedDomains) != 0 {
		t.Errorf("expected empty AllowedDomains, got %v", opts.AllowedDomains)
	}
	if len(opts.BlockedDomains) != 0 {
		t.Errorf("expected empty BlockedDomains, got %v", opts.BlockedDomains)
	}
}

func TestSearchOptions_WithDomains(t *testing.T) {
	opts := SearchOptions{
		Query:          "test",
		AllowedDomains: []string{"example.com", "test.org"},
		BlockedDomains: []string{"blocked.com"},
		NumResults:     5,
		Language:       "en",
	}

	if len(opts.AllowedDomains) != 2 {
		t.Errorf("expected 2 allowed domains, got %d", len(opts.AllowedDomains))
	}
	if len(opts.BlockedDomains) != 1 {
		t.Errorf("expected 1 blocked domain, got %d", len(opts.BlockedDomains))
	}
	if opts.NumResults != 5 {
		t.Errorf("expected NumResults 5, got %d", opts.NumResults)
	}
	if opts.Language != "en" {
		t.Errorf("expected language 'en', got %q", opts.Language)
	}
}

func TestSearcher_Search_MockServer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST method, got %s", r.Method)
		}
		if r.URL.Path != "/mcp" {
			t.Errorf("expected path '/mcp', got %s", r.URL.Path)
		}

		contentType := r.Header.Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("expected Content-Type 'application/json', got %s", contentType)
		}

		exaResults := exaSearchResponse{
			Results: []exaSearchResult{
				{
					Title:   "Mock Result",
					URL:     "https://mock.example.com",
					Summary: "Mock summary",
				},
			},
		}
		resultsJSON, _ := json.Marshal(exaResults)

		mcpResp := mcpResponse{}
		mcpResp.JSONRPC = "2.0"
		mcpResp.Result.Content = []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		}{
			{
				Type: "text",
				Text: string(resultsJSON),
			},
		}
		respJSON, _ := json.Marshal(mcpResp)

		w.Header().Set("Content-Type", "text/event-stream")
		fmt.Fprintf(w, "data: %s\n\n", respJSON)
	}))
	defer server.Close()

	s := NewSearcher()
	s.baseURL = server.URL

	ctx := context.Background()
	opts := SearchOptions{
		Query:      "test query",
		NumResults: 5,
	}

	result, err := s.Search(ctx, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result.Results))
	}

	if result.Results[0].Title != "Mock Result" {
		t.Errorf("expected title 'Mock Result', got %q", result.Results[0].Title)
	}
}

func TestSearcher_Search_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}))
	defer server.Close()

	s := NewSearcher()
	s.baseURL = server.URL

	ctx := context.Background()
	opts := SearchOptions{
		Query: "test",
	}

	_, err := s.Search(ctx, opts)
	if err == nil {
		t.Error("expected error for 500 response")
	}
}

func TestSearcher_Search_EmptyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	s := NewSearcher()
	s.baseURL = server.URL

	ctx := context.Background()
	opts := SearchOptions{
		Query: "test",
	}

	result, err := s.Search(ctx, opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Results) != 0 {
		t.Errorf("expected 0 results for empty response, got %d", len(result.Results))
	}
}

func TestSearcher_Search_NumResultsLimit(t *testing.T) {
	var receivedNumResults int

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req mcpRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}
		receivedNumResults = req.Params.Arguments.NumResults

		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	s := NewSearcher()
	s.baseURL = server.URL

	tests := []struct {
		input    int
		expected int
	}{
		{0, 8},
		{-1, 8},
		{5, 5},
		{20, 20},
		{25, 20},
	}

	for _, tc := range tests {
		ctx := context.Background()
		opts := SearchOptions{
			Query:      "test",
			NumResults: tc.input,
		}

		s.Search(ctx, opts)

		if receivedNumResults != tc.expected {
			t.Errorf("for input %d, expected numResults %d, got %d", tc.input, tc.expected, receivedNumResults)
		}
	}
}

func TestSearcher_Search_QueryWithDomains(t *testing.T) {
	var receivedQuery string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req mcpRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("failed to decode request: %v", err)
		}
		receivedQuery = req.Params.Arguments.Query

		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	s := NewSearcher()
	s.baseURL = server.URL

	ctx := context.Background()
	opts := SearchOptions{
		Query:          "test query",
		AllowedDomains: []string{"example.com"},
		BlockedDomains: []string{"blocked.com"},
	}

	s.Search(ctx, opts)

	expectedQuery := "test query site:example.com -site:blocked.com"
	if receivedQuery != expectedQuery {
		t.Errorf("expected query %q, got %q", expectedQuery, receivedQuery)
	}
}

func TestSearcher_Search_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	s := NewSearcher()
	s.baseURL = server.URL

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	opts := SearchOptions{
		Query: "test",
	}

	_, err := s.Search(ctx, opts)
	if err == nil {
		t.Error("expected error for cancelled context")
	}
}
