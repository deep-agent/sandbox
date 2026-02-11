package web

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewFetcher(t *testing.T) {
	f := NewFetcher()
	if f == nil {
		t.Fatal("expected non-nil Fetcher")
	}
	if f.client == nil {
		t.Error("expected non-nil http client")
	}
	if f.converter == nil {
		t.Error("expected non-nil converter")
	}
	if f.cache == nil {
		t.Error("expected non-nil cache")
	}
	if f.cacheTTL != 15*time.Minute {
		t.Errorf("expected cacheTTL to be 15 minutes, got %v", f.cacheTTL)
	}
}

func TestFetcher_HTTPSchemeUpgrade(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("test content"))
	}))
	defer server.Close()

	f := NewFetcher()
	f.client = server.Client()

	httpURL := "http" + server.URL[5:]

	ctx := context.Background()
	result, err := f.Fetch(ctx, httpURL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Content != "test content" {
		t.Errorf("expected 'test content', got %q", result.Content)
	}
}

func TestFetcher_Cache(t *testing.T) {
	callCount := 0
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("cached content"))
	}))
	defer server.Close()

	f := NewFetcher()
	f.client = server.Client()
	ctx := context.Background()

	_, err := f.Fetch(ctx, server.URL)
	if err != nil {
		t.Fatalf("first fetch error: %v", err)
	}

	_, err = f.Fetch(ctx, server.URL)
	if err != nil {
		t.Fatalf("second fetch error: %v", err)
	}

	if callCount != 1 {
		t.Errorf("expected 1 server call due to cache, got %d", callCount)
	}
}

func TestFetcher_CacheExpiration(t *testing.T) {
	callCount := 0
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("content"))
	}))
	defer server.Close()

	f := NewFetcher()
	f.client = server.Client()
	f.cacheTTL = 10 * time.Millisecond

	ctx := context.Background()

	_, err := f.Fetch(ctx, server.URL)
	if err != nil {
		t.Fatalf("first fetch error: %v", err)
	}

	time.Sleep(20 * time.Millisecond)

	_, err = f.Fetch(ctx, server.URL)
	if err != nil {
		t.Fatalf("second fetch error: %v", err)
	}

	if callCount != 2 {
		t.Errorf("expected 2 server calls after cache expiration, got %d", callCount)
	}
}

func TestFetcher_FetchHTML(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte("<html><body><h1>Hello</h1><p>World</p></body></html>"))
	}))
	defer server.Close()

	f := NewFetcher()
	f.client = server.Client()
	ctx := context.Background()

	result, err := f.Fetch(ctx, server.URL)
	if err != nil {
		t.Fatalf("fetch error: %v", err)
	}

	if result.Content == "" {
		t.Error("expected non-empty content")
	}
}

func TestFetcher_FetchPlainText(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("plain text content"))
	}))
	defer server.Close()

	f := NewFetcher()
	f.client = server.Client()
	ctx := context.Background()

	result, err := f.Fetch(ctx, server.URL)
	if err != nil {
		t.Fatalf("fetch error: %v", err)
	}

	if result.Content != "plain text content" {
		t.Errorf("expected 'plain text content', got %q", result.Content)
	}
}

func TestFetcher_HTTPError(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Not Found", http.StatusNotFound)
	}))
	defer server.Close()

	f := NewFetcher()
	f.client = server.Client()
	ctx := context.Background()

	_, err := f.Fetch(ctx, server.URL)
	if err == nil {
		t.Error("expected error for 404 response")
	}
}

func TestFetcher_InvalidURL(t *testing.T) {
	f := NewFetcher()
	ctx := context.Background()

	_, err := f.Fetch(ctx, "://invalid")
	if err == nil {
		t.Error("expected error for invalid URL")
	}
}

func TestFetcher_Redirect(t *testing.T) {
	targetServer := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("redirected content"))
	}))
	defer targetServer.Close()

	redirectServer := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, targetServer.URL, http.StatusFound)
	}))
	defer redirectServer.Close()

	f := NewFetcher()
	f.client = redirectServer.Client()
	ctx := context.Background()

	result, err := f.Fetch(ctx, redirectServer.URL)
	if err != nil {
		t.Fatalf("fetch error: %v", err)
	}

	if result.RedirectURL == "" && result.Content == "" {
		t.Error("expected either redirect URL or content")
	}
}

func TestFetcher_ContextCancellation(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.Write([]byte("delayed"))
	}))
	defer server.Close()

	f := NewFetcher()
	f.client = server.Client()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := f.Fetch(ctx, server.URL)
	if err == nil {
		t.Error("expected error for cancelled context")
	}
}

func TestFetcher_CleanExpiredCache(t *testing.T) {
	f := NewFetcher()
	f.cacheTTL = 10 * time.Millisecond

	f.cacheMu.Lock()
	f.cache["https://example.com/old"] = cacheEntry{
		content:   "old content",
		timestamp: time.Now().Add(-time.Hour),
	}
	f.cache["https://example.com/new"] = cacheEntry{
		content:   "new content",
		timestamp: time.Now(),
	}
	f.cacheMu.Unlock()

	f.cleanExpiredCache()

	f.cacheMu.RLock()
	defer f.cacheMu.RUnlock()

	if _, ok := f.cache["https://example.com/old"]; ok {
		t.Error("expected old entry to be removed")
	}
	if _, ok := f.cache["https://example.com/new"]; !ok {
		t.Error("expected new entry to remain")
	}
}
