package browser

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"
)

// getCDPURL discovers the WebSocket debugger URL from a Chrome instance
// running with --remote-debugging-port on the given port.
func getCDPURL(port int) (string, error) {
	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/json/version", port))
	if err != nil {
		return "", fmt.Errorf("Chrome CDP not available on port %d: %w", port, err)
	}
	defer resp.Body.Close()

	var result struct {
		WebSocketDebuggerURL string `json:"webSocketDebuggerUrl"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode CDP response: %w", err)
	}
	return result.WebSocketDebuggerURL, nil
}

func TestNewController(t *testing.T) {
	cdpURL := "ws://localhost:9222"
	c := NewController(cdpURL)

	if c == nil {
		t.Fatal("NewController returned nil")
	}
	if c.cdpURL != cdpURL {
		t.Errorf("cdpURL = %q, want %q", c.cdpURL, cdpURL)
	}
	if c.timeout != 30*time.Second {
		t.Errorf("timeout = %v, want %v", c.timeout, 30*time.Second)
	}
}

func TestScreenshotOptions(t *testing.T) {
	opts := &ScreenshotOptions{
		Format:  "png",
		Quality: 90,
		Full:    true,
	}

	if opts.Format != "png" {
		t.Errorf("Format = %q, want %q", opts.Format, "png")
	}
	if opts.Quality != 90 {
		t.Errorf("Quality = %d, want %d", opts.Quality, 90)
	}
	if !opts.Full {
		t.Error("Full should be true")
	}
}

func TestPageInfo(t *testing.T) {
	info := &PageInfo{
		URL:    "https://example.com",
		Title:  "Example",
		Width:  1920,
		Height: 1080,
	}

	if info.URL != "https://example.com" {
		t.Errorf("URL = %q, want %q", info.URL, "https://example.com")
	}
	if info.Title != "Example" {
		t.Errorf("Title = %q, want %q", info.Title, "Example")
	}
	if info.Width != 1920 {
		t.Errorf("Width = %d, want %d", info.Width, 1920)
	}
	if info.Height != 1080 {
		t.Errorf("Height = %d, want %d", info.Height, 1080)
	}
}

func TestController_GetInfo_Disconnected(t *testing.T) {
	c := NewController("ws://localhost:99999")

	info, err := c.GetInfo()
	if err != nil {
		t.Fatalf("GetInfo() error = %v", err)
	}

	if info == nil {
		t.Fatal("GetInfo() returned nil info")
	}
	if info.Status != "disconnected" {
		t.Errorf("Status = %q, want %q", info.Status, "disconnected")
	}
	if info.CDPURL != "ws://localhost:99999" {
		t.Errorf("CDPURL = %q, want %q", info.CDPURL, "ws://localhost:99999")
	}
}

func TestController_GetInfo_InvalidPort(t *testing.T) {
	c := NewController("ws://localhost:0")

	info, err := c.GetInfo()
	if err != nil {
		t.Fatalf("GetInfo() error = %v", err)
	}

	if info.Status != "disconnected" {
		t.Errorf("Status = %q, want %q", info.Status, "disconnected")
	}
}

func TestScreenshotOptions_DefaultValues(t *testing.T) {
	opts := &ScreenshotOptions{}

	if opts.Format != "" {
		t.Errorf("default Format = %q, want empty", opts.Format)
	}
	if opts.Quality != 0 {
		t.Errorf("default Quality = %d, want 0", opts.Quality)
	}
	if opts.Full {
		t.Error("default Full should be false")
	}
}

func TestPageInfo_ZeroValues(t *testing.T) {
	info := &PageInfo{}

	if info.URL != "" {
		t.Errorf("default URL = %q, want empty", info.URL)
	}
	if info.Title != "" {
		t.Errorf("default Title = %q, want empty", info.Title)
	}
	if info.Width != 0 {
		t.Errorf("default Width = %d, want 0", info.Width)
	}
	if info.Height != 0 {
		t.Errorf("default Height = %d, want 0", info.Height)
	}
}

// TestController_CDP tests all browser controller capabilities against a real
// Chrome instance. Requires Chrome running with --remote-debugging-port=9222.
// Skipped automatically if Chrome CDP is not available.
func TestController_CDP(t *testing.T) {
	cdpURL, err := getCDPURL(9222)
	if err != nil {
		t.Skipf("Skipping CDP tests: %v", err)
	}

	c := NewController(cdpURL)
	defer c.Close()

	t.Run("GetInfo", func(t *testing.T) {
		info, err := c.GetInfo()
		if err != nil {
			t.Fatalf("GetInfo failed: %v", err)
		}
		if info.Status != "connected" {
			t.Fatalf("expected connected, got %s", info.Status)
		}
		t.Logf("status=%s", info.Status)
	})

	t.Run("Navigate", func(t *testing.T) {
		if err := c.Navigate("https://www.example.com"); err != nil {
			t.Fatalf("Navigate failed: %v", err)
		}
	})

	t.Run("GetCurrentURL", func(t *testing.T) {
		url, err := c.GetCurrentURL()
		if err != nil {
			t.Fatalf("GetCurrentURL failed: %v", err)
		}
		if !strings.Contains(url, "example.com") {
			t.Fatalf("expected example.com in URL, got %s", url)
		}
		t.Logf("url=%s", url)
	})

	t.Run("GetTitle", func(t *testing.T) {
		title, err := c.GetTitle()
		if err != nil {
			t.Fatalf("GetTitle failed: %v", err)
		}
		if title == "" {
			t.Fatal("returned empty title")
		}
		t.Logf("title=%s", title)
	})

	t.Run("GetPageInfo", func(t *testing.T) {
		info, err := c.GetPageInfo()
		if err != nil {
			t.Fatalf("GetPageInfo failed: %v", err)
		}
		if info.URL == "" || info.Title == "" {
			t.Fatalf("empty fields: url=%s title=%s", info.URL, info.Title)
		}
		t.Logf("url=%s title=%s", info.URL, info.Title)
	})

	t.Run("Screenshot_Viewport", func(t *testing.T) {
		b64, err := c.Screenshot(nil)
		if err != nil {
			t.Fatalf("Screenshot failed: %v", err)
		}
		data, _ := base64.StdEncoding.DecodeString(b64)
		if len(data) < 100 {
			t.Fatal("screenshot too small")
		}
		t.Logf("%d bytes", len(data))
	})

	t.Run("Screenshot_Full", func(t *testing.T) {
		b64, err := c.Screenshot(&ScreenshotOptions{Full: true})
		if err != nil {
			t.Fatalf("Screenshot(full) failed: %v", err)
		}
		data, _ := base64.StdEncoding.DecodeString(b64)
		if len(data) < 100 {
			t.Fatal("screenshot too small")
		}
		t.Logf("%d bytes", len(data))
	})

	t.Run("GetHTML", func(t *testing.T) {
		html, err := c.GetHTML("body")
		if err != nil {
			t.Fatalf("GetHTML failed: %v", err)
		}
		if !strings.Contains(html, "Example Domain") {
			t.Fatalf("expected 'Example Domain' in HTML, got: %s", html[:min(200, len(html))])
		}
		t.Logf("%d chars", len(html))
	})

	t.Run("WaitVisible", func(t *testing.T) {
		if err := c.WaitVisible("h1"); err != nil {
			t.Fatalf("WaitVisible failed: %v", err)
		}
	})

	t.Run("Evaluate", func(t *testing.T) {
		result, err := c.Evaluate("document.title")
		if err != nil {
			t.Fatalf("Evaluate failed: %v", err)
		}
		title := fmt.Sprintf("%v", result)
		if !strings.Contains(title, "Example Domain") {
			t.Fatalf("expected 'Example Domain', got: %v", result)
		}
		t.Logf("result=%v", result)
	})

	t.Run("Scroll", func(t *testing.T) {
		if err := c.Scroll(0, 100); err != nil {
			t.Fatalf("Scroll failed: %v", err)
		}
	})

	t.Run("Click", func(t *testing.T) {
		if err := c.Click("a"); err != nil {
			t.Fatalf("Click failed: %v", err)
		}
	})

	t.Run("Type", func(t *testing.T) {
		if err := c.Navigate(`data:text/html,<html><body><input id="test" type="text"></body></html>`); err != nil {
			t.Fatalf("Navigate to test page failed: %v", err)
		}
		if err := c.Type("#test", "hello world"); err != nil {
			t.Fatalf("Type failed: %v", err)
		}
		result, err := c.Evaluate(`document.getElementById("test").value`)
		if err != nil {
			t.Fatalf("Evaluate after Type failed: %v", err)
		}
		t.Logf("input value=%v", result)
	})

	t.Run("GetCookies", func(t *testing.T) {
		_ = c.Navigate("https://www.example.com")
		cookies, err := c.GetCookies()
		if err != nil {
			t.Fatalf("GetCookies failed: %v", err)
		}
		t.Logf("cookies=%s", cookies)
	})

	t.Run("PDF", func(t *testing.T) {
		b64, err := c.PDF()
		if err != nil {
			t.Fatalf("PDF failed: %v", err)
		}
		data, _ := base64.StdEncoding.DecodeString(b64)
		if len(data) < 100 {
			t.Fatal("PDF too small")
		}
		t.Logf("%d bytes", len(data))
	})
}
