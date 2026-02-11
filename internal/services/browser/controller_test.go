package browser

import (
	"testing"
	"time"
)

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
