package sandbox

import (
	"testing"
	"time"

	"github.com/deep-agent/sandbox/model"
)

func TestBrowserGetInfo(t *testing.T) {
	client := newTestClient()

	info, err := client.BrowserGetInfo()
	if err != nil {
		t.Fatalf("BrowserGetInfo error: %v", err)
	}
	if info == nil {
		t.Fatal("expected info to be non-nil")
	}
	if info.Status == "" {
		t.Error("expected status to be non-empty")
	}
	t.Logf("BrowserInfo: cdp_url=%s, status=%s", info.CDPURL, info.Status)
}

func TestBrowserNavigate(t *testing.T) {
	client := NewClient(testBaseURL, WithSecret(testSecret), WithTimeout(60*time.Second))

	err := client.BrowserNavigate(&model.BrowserNavigateRequest{
		URL: "https://www.example.com",
	})
	if err != nil {
		t.Fatalf("BrowserNavigate error: %v", err)
	}
}

func TestBrowserGetCurrentURL(t *testing.T) {
	client := newTestClient()

	result, err := client.BrowserGetCurrentURL()
	if err != nil {
		t.Fatalf("BrowserGetCurrentURL error: %v", err)
	}
	if result == nil {
		t.Fatal("expected result to be non-nil")
	}
	t.Logf("Current URL: %s", result.URL)
}

func TestBrowserGetTitle(t *testing.T) {
	client := newTestClient()

	result, err := client.BrowserGetTitle()
	if err != nil {
		t.Fatalf("BrowserGetTitle error: %v", err)
	}
	if result == nil {
		t.Fatal("expected result to be non-nil")
	}
	t.Logf("Title: %s", result.Title)
}

func TestBrowserScreenshot(t *testing.T) {
	client := newTestClient()

	result, err := client.BrowserScreenshot(&model.BrowserScreenshotRequest{
		Format: "png",
	})
	if err != nil {
		t.Fatalf("BrowserScreenshot error: %v", err)
	}
	if result == nil {
		t.Fatal("expected result to be non-nil")
	}
	if result.Screenshot == "" {
		t.Error("expected screenshot data to be non-empty")
	}
	t.Logf("Screenshot data length: %d", len(result.Screenshot))
}

func TestBrowserEvaluate(t *testing.T) {
	client := newTestClient()

	result, err := client.BrowserEvaluate(&model.BrowserEvaluateRequest{
		Expression: "1 + 1",
	})
	if err != nil {
		t.Fatalf("BrowserEvaluate error: %v", err)
	}
	if result == nil {
		t.Fatal("expected result to be non-nil")
	}
	t.Logf("Evaluate result: %v", result.Result)
}

func TestBrowserGetHTML(t *testing.T) {
	t.Skip("skipping: browser GetHTML may timeout on some pages")
}
