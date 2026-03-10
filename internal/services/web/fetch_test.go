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
		_, _ = w.Write([]byte("test content"))
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
		_, _ = w.Write([]byte("cached content"))
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
		_, _ = w.Write([]byte("content"))
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

func TestFetcher_CacheMaxSizeEviction(t *testing.T) {
	callCount := 0
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte(r.URL.Path))
	}))
	defer server.Close()

	f := NewFetcher()
	f.client = server.Client()
	f.cacheTTL = time.Hour
	f.maxCacheSize = 2
	f.maxCacheBytes = 1024 * 1024

	ctx := context.Background()
	urlA := server.URL + "/a"
	urlB := server.URL + "/b"
	urlC := server.URL + "/c"

	if _, err := f.Fetch(ctx, urlA); err != nil {
		t.Fatalf("fetch a error: %v", err)
	}
	time.Sleep(2 * time.Millisecond)
	if _, err := f.Fetch(ctx, urlB); err != nil {
		t.Fatalf("fetch b error: %v", err)
	}
	time.Sleep(2 * time.Millisecond)
	if _, err := f.Fetch(ctx, urlC); err != nil {
		t.Fatalf("fetch c error: %v", err)
	}

	f.cacheMu.RLock()
	if len(f.cache) != 2 {
		f.cacheMu.RUnlock()
		t.Fatalf("expected cache size 2, got %d", len(f.cache))
	}
	if _, ok := f.cache[urlA]; ok {
		f.cacheMu.RUnlock()
		t.Fatalf("expected urlA to be evicted")
	}
	if _, ok := f.cache[urlB]; !ok {
		f.cacheMu.RUnlock()
		t.Fatalf("expected urlB to remain")
	}
	if _, ok := f.cache[urlC]; !ok {
		f.cacheMu.RUnlock()
		t.Fatalf("expected urlC to remain")
	}
	f.cacheMu.RUnlock()

	if _, err := f.Fetch(ctx, urlA); err != nil {
		t.Fatalf("fetch a again error: %v", err)
	}
	if callCount != 4 {
		t.Fatalf("expected 4 server calls, got %d", callCount)
	}
}

func TestFetcher_CacheMaxBytesEviction(t *testing.T) {
	callCount := 0
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.Header().Set("Content-Type", "text/plain")
		switch r.URL.Path {
		case "/a":
			_, _ = w.Write([]byte("aaaaaa"))
		case "/b":
			_, _ = w.Write([]byte("bbbbbb"))
		case "/big":
			_, _ = w.Write([]byte("xxxxxxxxxxxxxxxxxxxx"))
		default:
			_, _ = w.Write([]byte("ok"))
		}
	}))
	defer server.Close()

	f := NewFetcher()
	f.client = server.Client()
	f.cacheTTL = time.Hour
	f.maxCacheSize = 100
	f.maxCacheBytes = 10

	ctx := context.Background()
	urlA := server.URL + "/a"
	urlB := server.URL + "/b"
	urlBig := server.URL + "/big"

	if _, err := f.Fetch(ctx, urlA); err != nil {
		t.Fatalf("fetch a error: %v", err)
	}
	time.Sleep(2 * time.Millisecond)
	if _, err := f.Fetch(ctx, urlB); err != nil {
		t.Fatalf("fetch b error: %v", err)
	}

	f.cacheMu.RLock()
	if _, ok := f.cache[urlA]; ok {
		f.cacheMu.RUnlock()
		t.Fatalf("expected urlA to be evicted by bytes limit")
	}
	if _, ok := f.cache[urlB]; !ok {
		f.cacheMu.RUnlock()
		t.Fatalf("expected urlB to remain")
	}
	if f.currentBytes != 6 {
		f.cacheMu.RUnlock()
		t.Fatalf("expected currentBytes=6, got %d", f.currentBytes)
	}
	f.cacheMu.RUnlock()

	if _, err := f.Fetch(ctx, urlBig); err != nil {
		t.Fatalf("fetch big error: %v", err)
	}

	f.cacheMu.RLock()
	if _, ok := f.cache[urlBig]; ok {
		f.cacheMu.RUnlock()
		t.Fatalf("expected oversized entry not to be cached")
	}
	if _, ok := f.cache[urlB]; !ok {
		f.cacheMu.RUnlock()
		t.Fatalf("expected urlB to remain after big fetch")
	}
	if f.currentBytes != 6 {
		f.cacheMu.RUnlock()
		t.Fatalf("expected currentBytes=6 after big fetch, got %d", f.currentBytes)
	}
	f.cacheMu.RUnlock()

	if callCount != 3 {
		t.Fatalf("expected 3 server calls, got %d", callCount)
	}
}

func TestFetcher_FetchHTML(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte("<html><body><h1>Hello</h1><p>World</p></body></html>"))
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
		_, _ = w.Write([]byte("plain text content"))
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
		_, _ = w.Write([]byte("redirected content"))
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
		_, _ = w.Write([]byte("delayed"))
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
