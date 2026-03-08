package web

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	md "github.com/JohannesKaufmann/html-to-markdown"
)

type FetchResult struct {
	Content     string
	RedirectURL string
}

const (
	DefaultCacheTTL      = 15 * time.Minute
	DefaultMaxCacheSize  = 100
	DefaultMaxCacheBytes = 50 * 1024 * 1024
)

type cacheEntry struct {
	content   string
	timestamp time.Time
	size      int
}

type Fetcher struct {
	client        *http.Client
	converter     *md.Converter
	cache         map[string]cacheEntry
	cacheMu       sync.RWMutex
	cacheTTL      time.Duration
	maxCacheSize  int
	maxCacheBytes int
	currentBytes  int
}

func NewFetcher() *Fetcher {
	return &Fetcher{
		client: &http.Client{
			Timeout: 30 * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= 10 {
					return fmt.Errorf("too many redirects")
				}
				return nil
			},
		},
		converter:     md.NewConverter("", true, nil),
		cache:         make(map[string]cacheEntry),
		cacheTTL:      DefaultCacheTTL,
		maxCacheSize:  DefaultMaxCacheSize,
		maxCacheBytes: DefaultMaxCacheBytes,
	}
}

func (f *Fetcher) Fetch(ctx context.Context, rawURL string) (*FetchResult, error) {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return nil, fmt.Errorf("invalid URL: %w", err)
	}

	if parsedURL.Scheme == "http" {
		parsedURL.Scheme = "https"
		rawURL = parsedURL.String()
	}

	f.cacheMu.RLock()
	if entry, ok := f.cache[rawURL]; ok {
		if time.Since(entry.timestamp) < f.cacheTTL {
			f.cacheMu.RUnlock()
			return &FetchResult{Content: entry.content}, nil
		}
	}
	f.cacheMu.RUnlock()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; SandboxBot/1.0)")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.Request.URL.Host != parsedURL.Host {
		return &FetchResult{
			RedirectURL: resp.Request.URL.String(),
		}, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error: %d %s", resp.StatusCode, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	contentType := resp.Header.Get("Content-Type")
	var content string

	if strings.Contains(contentType, "text/html") {
		content, err = f.converter.ConvertString(string(body))
		if err != nil {
			content = string(body)
		}
	} else {
		content = string(body)
	}

	f.cacheMu.Lock()
	now := time.Now()
	f.cleanExpiredCacheLocked(now)

	size := len(content)
	if f.maxCacheBytes <= 0 || size <= f.maxCacheBytes {
		if old, ok := f.cache[rawURL]; ok {
			f.currentBytes -= old.size
		}
		f.cache[rawURL] = cacheEntry{
			content:   content,
			timestamp: now,
			size:      size,
		}
		f.currentBytes += size
		f.ensureCapacityLocked()
	}
	if f.currentBytes < 0 {
		f.currentBytes = 0
	}
	f.cacheMu.Unlock()

	return &FetchResult{Content: content}, nil
}

func (f *Fetcher) cleanExpiredCache() {
	f.cacheMu.Lock()
	defer f.cacheMu.Unlock()

	f.cleanExpiredCacheLocked(time.Now())
	if f.currentBytes < 0 {
		f.currentBytes = 0
	}
}

func (f *Fetcher) cleanExpiredCacheLocked(now time.Time) {
	for key, entry := range f.cache {
		if now.Sub(entry.timestamp) > f.cacheTTL {
			delete(f.cache, key)
			f.currentBytes -= entry.size
		}
	}
}

func (f *Fetcher) ensureCapacityLocked() {
	for {
		tooManyEntries := f.maxCacheSize > 0 && len(f.cache) > f.maxCacheSize
		tooManyBytes := f.maxCacheBytes > 0 && f.currentBytes > f.maxCacheBytes
		if !tooManyEntries && !tooManyBytes {
			return
		}

		var oldestKey string
		var oldestTime time.Time
		found := false
		for key, entry := range f.cache {
			if !found || entry.timestamp.Before(oldestTime) {
				oldestKey = key
				oldestTime = entry.timestamp
				found = true
			}
		}
		if !found {
			return
		}

		entry := f.cache[oldestKey]
		delete(f.cache, oldestKey)
		f.currentBytes -= entry.size
	}
}
