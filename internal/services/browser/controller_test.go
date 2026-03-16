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
// Run `make cdp` to start Chrome, then run the tests.
// Skipped automatically if Chrome CDP is not available.
func TestController_CDP(t *testing.T) {
	cdpURL, err := getCDPURL(9222)
	if err != nil {
		t.Skipf("Skipping CDP tests: %v", err)
	}

	c := NewController(cdpURL)
	defer c.Close()

	// ---- Basic connectivity ----

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

	// ---- Navigation ----

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

	// ---- Screenshots ----

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

	// ---- HTML ----

	t.Run("GetHTML_WithSelector", func(t *testing.T) {
		html, err := c.GetHTML("body")
		if err != nil {
			t.Fatalf("GetHTML failed: %v", err)
		}
		if !strings.Contains(html, "Example Domain") {
			t.Fatalf("expected 'Example Domain' in HTML, got: %s", html[:min(200, len(html))])
		}
		t.Logf("%d chars", len(html))
	})

	t.Run("GetHTML_FullPage", func(t *testing.T) {
		html, err := c.GetHTML("html")
		if err != nil {
			t.Fatalf("GetHTML(html) failed: %v", err)
		}
		if !strings.Contains(html, "<head>") {
			t.Fatalf("expected <head> in full page HTML")
		}
		t.Logf("full page HTML: %d chars", len(html))
	})

	// ---- Visibility ----

	t.Run("WaitVisible", func(t *testing.T) {
		if err := c.WaitVisible("h1"); err != nil {
			t.Fatalf("WaitVisible failed: %v", err)
		}
	})

	// ---- Evaluate ----

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

	// ---- Scroll (legacy absolute) ----

	t.Run("Scroll", func(t *testing.T) {
		if err := c.Scroll(0, 100); err != nil {
			t.Fatalf("Scroll failed: %v", err)
		}
	})

	// ---- ScrollByDirection ----

	t.Run("ScrollByDirection_Down", func(t *testing.T) {
		if err := c.ScrollByDirection("down"); err != nil {
			t.Fatalf("ScrollByDirection(down) failed: %v", err)
		}
	})

	t.Run("ScrollByDirection_Up", func(t *testing.T) {
		if err := c.ScrollByDirection("up"); err != nil {
			t.Fatalf("ScrollByDirection(up) failed: %v", err)
		}
	})

	// ---- Click (selector) ----

	t.Run("Click", func(t *testing.T) {
		// Navigate back to example.com which has a link
		_ = c.Navigate("https://www.example.com")
		if err := c.Click("a"); err != nil {
			t.Fatalf("Click failed: %v", err)
		}
	})

	// ---- ClickCoordinate ----

	t.Run("ClickCoordinate", func(t *testing.T) {
		_ = c.Navigate("https://www.example.com")
		// Click somewhere in the viewport (center-ish)
		if err := c.ClickCoordinate(200, 200); err != nil {
			t.Fatalf("ClickCoordinate failed: %v", err)
		}
	})

	// ---- Type ----

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

	// ---- GoBack ----

	t.Run("GoBack", func(t *testing.T) {
		// Navigate to example.com, then to another page, then go back
		if err := c.Navigate("https://www.example.com"); err != nil {
			t.Fatalf("Navigate to example.com failed: %v", err)
		}
		if err := c.Navigate("https://www.example.org"); err != nil {
			t.Fatalf("Navigate to example.org failed: %v", err)
		}
		if err := c.GoBack(); err != nil {
			t.Fatalf("GoBack failed: %v", err)
		}
		// Give the browser a moment to navigate back
		time.Sleep(500 * time.Millisecond)
		url, err := c.GetCurrentURL()
		if err != nil {
			t.Fatalf("GetCurrentURL after GoBack failed: %v", err)
		}
		t.Logf("URL after GoBack: %s", url)
	})

	// ---- GetState ----

	t.Run("GetState_NoScreenshot", func(t *testing.T) {
		_ = c.Navigate("https://www.example.com")
		state, err := c.GetState(false)
		if err != nil {
			t.Fatalf("GetState failed: %v", err)
		}
		if state.URL == "" {
			t.Fatal("GetState returned empty URL")
		}
		if state.Title == "" {
			t.Fatal("GetState returned empty title")
		}
		if state.Screenshot != "" {
			t.Fatal("GetState should not include screenshot when not requested")
		}
		t.Logf("url=%s title=%s elements=%d tabs=%d",
			state.URL, state.Title, len(state.InteractiveElements), len(state.Tabs))
		// example.com has at least one link
		if len(state.InteractiveElements) == 0 {
			t.Fatal("expected at least one interactive element on example.com")
		}
		for i, elem := range state.InteractiveElements {
			t.Logf("  element[%d]: tag=%s text=%q selector=%q href=%q",
				i, elem.Tag, elem.Text, elem.Selector, elem.Href)
		}
	})

	t.Run("GetState_WithScreenshot", func(t *testing.T) {
		state, err := c.GetState(true)
		if err != nil {
			t.Fatalf("GetState(screenshot) failed: %v", err)
		}
		if state.Screenshot == "" {
			t.Fatal("GetState should include screenshot when requested")
		}
		data, _ := base64.StdEncoding.DecodeString(state.Screenshot)
		if len(data) < 100 {
			t.Fatal("screenshot too small")
		}
		t.Logf("screenshot=%d bytes, elements=%d", len(data), len(state.InteractiveElements))
	})

	// ---- Tabs ----

	t.Run("ListTabs", func(t *testing.T) {
		tabs, err := c.ListTabs()
		if err != nil {
			t.Fatalf("ListTabs failed: %v", err)
		}
		if len(tabs) == 0 {
			t.Fatal("expected at least one tab")
		}
		for i, tab := range tabs {
			t.Logf("  tab[%d]: id=%s url=%s title=%s", i, tab.TabID, tab.URL, tab.Title)
		}
	})

	t.Run("NavigateNewTab", func(t *testing.T) {
		// Count tabs before
		tabsBefore, err := c.ListTabs()
		if err != nil {
			t.Fatalf("ListTabs before failed: %v", err)
		}
		countBefore := len(tabsBefore)

		if err := c.NavigateNewTab("https://www.example.org"); err != nil {
			t.Fatalf("NavigateNewTab failed: %v", err)
		}
		// Wait for new tab to appear
		time.Sleep(1 * time.Second)

		tabsAfter, err := c.ListTabs()
		if err != nil {
			t.Fatalf("ListTabs after failed: %v", err)
		}
		if len(tabsAfter) <= countBefore {
			t.Fatalf("expected more tabs after NavigateNewTab, before=%d after=%d", countBefore, len(tabsAfter))
		}
		t.Logf("tabs before=%d, after=%d", countBefore, len(tabsAfter))
		for i, tab := range tabsAfter {
			t.Logf("  tab[%d]: id=%s url=%s", i, tab.TabID, tab.URL)
		}
	})

	t.Run("SwitchTab", func(t *testing.T) {
		tabs, err := c.ListTabs()
		if err != nil {
			t.Fatalf("ListTabs failed: %v", err)
		}
		if len(tabs) < 2 {
			t.Skip("need at least 2 tabs to test SwitchTab")
		}
		// Switch to the first tab
		targetTab := tabs[0]
		if err := c.SwitchTab(targetTab.TabID); err != nil {
			t.Fatalf("SwitchTab failed: %v", err)
		}
		t.Logf("switched to tab %s (%s)", targetTab.TabID, targetTab.URL)
	})

	t.Run("CloseTab", func(t *testing.T) {
		tabs, err := c.ListTabs()
		if err != nil {
			t.Fatalf("ListTabs failed: %v", err)
		}
		if len(tabs) < 2 {
			t.Skip("need at least 2 tabs to test CloseTab")
		}
		// Find the example.org tab we created in NavigateNewTab
		// (avoid closing our controller's own tab)
		var targetTab *TabInfo
		for i := range tabs {
			if strings.Contains(tabs[i].URL, "example.org") {
				targetTab = &tabs[i]
				break
			}
		}
		if targetTab == nil {
			// Fall back: close any tab that isn't the controller's current URL
			currentURL, _ := c.GetCurrentURL()
			for i := range tabs {
				if tabs[i].URL != currentURL {
					targetTab = &tabs[i]
					break
				}
			}
		}
		if targetTab == nil {
			t.Skip("no safe tab to close")
		}
		if err := c.CloseTab(targetTab.TabID); err != nil {
			t.Fatalf("CloseTab failed: %v", err)
		}
		// Give browser time to process the close
		time.Sleep(500 * time.Millisecond)
		// Verify tab count decreased
		tabsAfter, err := c.ListTabs()
		if err != nil {
			t.Fatalf("ListTabs after close failed: %v", err)
		}
		if len(tabsAfter) >= len(tabs) {
			t.Fatalf("tab count should decrease, before=%d after=%d", len(tabs), len(tabsAfter))
		}
		t.Logf("closed tab %s (%s), tabs: %d -> %d", targetTab.TabID, targetTab.URL, len(tabs), len(tabsAfter))
	})

	// ---- Cookies ----

	t.Run("GetCookies", func(t *testing.T) {
		_ = c.Navigate("https://www.example.com")
		cookies, err := c.GetCookies()
		if err != nil {
			t.Fatalf("GetCookies failed: %v", err)
		}
		t.Logf("cookies=%s", cookies)
	})

	// ---- PDF ----

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

// TestController_BaiduSearch is an end-to-end integration test that:
//  1. Opens https://www.baidu.com/
//  2. Types "北京今天天气" into the search box
//  3. Clicks the search button
//  4. Waits for and extracts search results
//  5. Closes the page tab
//
// Requires Chrome running with --remote-debugging-port=9222 (make cdp).
func TestController_BaiduSearch(t *testing.T) {
	cdpURL, err := getCDPURL(9222)
	if err != nil {
		t.Skipf("Skipping: %v", err)
	}

	c := NewController(cdpURL)
	defer c.Close()

	// Step 1: Navigate to Baidu
	t.Log("Step 1: Navigating to baidu.com ...")
	if err := c.Navigate("https://www.baidu.com/"); err != nil {
		t.Fatalf("Navigate failed: %v", err)
	}
	// Wait for the search input to appear
	if err := c.WaitVisible("#kw"); err != nil {
		t.Fatalf("WaitVisible(#kw) failed: %v", err)
	}
	t.Log("  -> baidu.com loaded")

	// Step 2: Type the search query
	t.Log("Step 2: Typing '北京今天天气' ...")
	// Baidu's input has complex JS handlers; use Evaluate to set value and
	// dispatch an input event so Baidu's JS recognises the change.
	_, err = c.Evaluate(`(function(){
		var el = document.querySelector("#kw");
		el.focus();
		el.value = "北京今天天气";
		el.dispatchEvent(new Event("input", {bubbles: true}));
	})()`)
	if err != nil {
		t.Fatalf("Set input value failed: %v", err)
	}
	// Verify the input value
	val, err := c.Evaluate(`document.querySelector("#kw").value`)
	if err != nil {
		t.Fatalf("Evaluate input value failed: %v", err)
	}
	t.Logf("  -> input value: %v", val)

	// Step 3: Click the search button
	t.Log("Step 3: Submitting search ...")
	// Baidu's autocomplete overlay can block the #su button;
	// submit the form via JS for reliability.
	_, err = c.Evaluate(`document.querySelector("#su").click()`)
	if err != nil {
		// Fallback: submit the form directly
		_, err = c.Evaluate(`document.querySelector("#form").submit()`)
		if err != nil {
			t.Fatalf("Submit search failed: %v", err)
		}
	}
	// Wait for results page to load
	time.Sleep(3 * time.Second)

	// Verify we navigated to search results
	currentURL, err := c.GetCurrentURL()
	if err != nil {
		t.Fatalf("GetCurrentURL failed: %v", err)
	}
	t.Logf("  -> results URL: %s", currentURL)
	if !strings.Contains(currentURL, "baidu.com") {
		t.Fatalf("expected baidu.com in URL, got %s", currentURL)
	}

	// Step 4: Extract search results
	t.Log("Step 4: Extracting search results ...")

	// Get page state to see the interactive elements
	state, err := c.GetState(false)
	if err != nil {
		t.Fatalf("GetState failed: %v", err)
	}
	t.Logf("  -> page title: %s", state.Title)
	t.Logf("  -> interactive elements: %d", len(state.InteractiveElements))

	// Extract the text content of the search results via JavaScript
	resultsJS := `(function() {
		var items = document.querySelectorAll('.result, .c-container, .result-op');
		var results = [];
		items.forEach(function(item, i) {
			if (i >= 5) return; // top 5 results
			var title = '';
			var desc = '';
			var titleEl = item.querySelector('h3, .t');
			if (titleEl) title = titleEl.textContent.trim();
			var descEl = item.querySelector('.c-abstract, .content-right_8Zs40, span[class*="content"]');
			if (descEl) desc = descEl.textContent.trim();
			if (title) results.push({index: i+1, title: title, desc: desc.substring(0, 200)});
		});
		return JSON.stringify(results);
	})()`

	resultJSON, err := c.Evaluate(resultsJS)
	if err != nil {
		t.Fatalf("Evaluate results failed: %v", err)
	}

	resultStr := fmt.Sprintf("%v", resultJSON)
	t.Logf("  -> search results JSON: %s", resultStr)

	// Also get a simplified text extract as fallback
	textJS := `(function() {
		var el = document.querySelector('#content_left');
		if (!el) el = document.querySelector('#wrapper');
		if (!el) el = document.body;
		return el.innerText.substring(0, 2000);
	})()`

	textContent, err := c.Evaluate(textJS)
	if err != nil {
		t.Fatalf("Evaluate text content failed: %v", err)
	}
	contentStr := fmt.Sprintf("%v", textContent)
	if len(contentStr) < 50 {
		t.Fatalf("search results too short, got %d chars: %s", len(contentStr), contentStr)
	}
	t.Logf("  -> text content (%d chars):\n%s", len(contentStr), contentStr)

	// Take a screenshot for visual verification
	screenshot, err := c.Screenshot(nil)
	if err != nil {
		t.Fatalf("Screenshot failed: %v", err)
	}
	data, _ := base64.StdEncoding.DecodeString(screenshot)
	t.Logf("  -> screenshot: %d bytes", len(data))

	// Step 5: Close the tab (navigate away to clean up)
	t.Log("Step 5: Closing page ...")
	if err := c.Navigate("about:blank"); err != nil {
		t.Fatalf("Navigate to blank failed: %v", err)
	}
	t.Log("  -> done, page closed")
}
