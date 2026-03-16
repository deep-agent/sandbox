package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/deep-agent/sandbox/internal/services/browser"
	"github.com/mark3labs/mcp-go/mcp"
)

// --- browser_navigate ---

func BrowserNavigateToolDef() mcp.Tool {
	return mcp.NewTool("browser_navigate",
		mcp.WithDescription("Navigate the browser to a specified URL."),
		mcp.WithString("url",
			mcp.Required(),
			mcp.Description("The URL to navigate to"),
		),
		mcp.WithBoolean("new_tab",
			mcp.Description("Whether to open in a new tab. Default: false"),
		),
	)
}

func BrowserNavigateHandler(cdpURL string) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		url, err := request.RequireString("url")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		newTab := request.GetBool("new_tab", false)

		controller := browser.NewController(cdpURL)
		if newTab {
			if err := controller.NavigateNewTab(url); err != nil {
				return mcp.NewToolResultError("Error: " + err.Error()), nil
			}
			return mcp.NewToolResultText(fmt.Sprintf("Opened new tab with URL: %s", url)), nil
		}

		if err := controller.Navigate(url); err != nil {
			return mcp.NewToolResultError("Error: " + err.Error()), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Navigated to: %s", url)), nil
	}
}

// --- browser_click ---

func BrowserClickToolDef() mcp.Tool {
	return mcp.NewTool("browser_click",
		mcp.WithDescription("Click an element by CSS selector or at specific viewport coordinates. Use selector for CSS-based clicking, or coordinate_x/coordinate_y for pixel-precise clicking."),
		mcp.WithString("selector",
			mcp.Description("CSS selector for the element to click. Use this OR coordinates."),
		),
		mcp.WithNumber("coordinate_x",
			mcp.Description("X coordinate (pixels from left edge of viewport). Use with coordinate_y."),
		),
		mcp.WithNumber("coordinate_y",
			mcp.Description("Y coordinate (pixels from top edge of viewport). Use with coordinate_x."),
		),
	)
}

func BrowserClickHandler(cdpURL string) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		controller := browser.NewController(cdpURL)

		// Check for coordinate-based clicking
		coordX := request.GetFloat("coordinate_x", -1)
		coordY := request.GetFloat("coordinate_y", -1)
		if coordX >= 0 && coordY >= 0 {
			if err := controller.ClickCoordinate(int(coordX), int(coordY)); err != nil {
				return mcp.NewToolResultError("Error: " + err.Error()), nil
			}
			return mcp.NewToolResultText(fmt.Sprintf("Clicked at coordinates (%d, %d)", int(coordX), int(coordY))), nil
		}

		// Fall back to selector-based clicking
		selector, err := request.RequireString("selector")
		if err != nil {
			return mcp.NewToolResultError("Provide either selector or both coordinate_x and coordinate_y"), nil
		}

		if err := controller.Click(selector); err != nil {
			return mcp.NewToolResultError("Error: " + err.Error()), nil
		}
		return mcp.NewToolResultText(fmt.Sprintf("Clicked element: %s", selector)), nil
	}
}

// --- browser_type ---

func BrowserTypeToolDef() mcp.Tool {
	return mcp.NewTool("browser_type",
		mcp.WithDescription("Type text into an input element in the browser."),
		mcp.WithString("selector",
			mcp.Required(),
			mcp.Description("CSS selector for the input element"),
		),
		mcp.WithString("text",
			mcp.Required(),
			mcp.Description("The text to type"),
		),
	)
}

func BrowserTypeHandler(cdpURL string) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		selector, err := request.RequireString("selector")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		text, err := request.RequireString("text")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		controller := browser.NewController(cdpURL)
		if err := controller.Type(selector, text); err != nil {
			return mcp.NewToolResultError("Error: " + err.Error()), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Typed '%s' into element: %s", text, selector)), nil
	}
}

// --- browser_get_state ---

func BrowserGetStateToolDef() mcp.Tool {
	return mcp.NewTool("browser_get_state",
		mcp.WithDescription("Get the current state of the page including URL, title, tabs, and all interactive elements."),
		mcp.WithBoolean("include_screenshot",
			mcp.Description("Whether to include a screenshot of the current page. Default: false"),
		),
	)
}

func BrowserGetStateHandler(cdpURL string) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		includeScreenshot := request.GetBool("include_screenshot", false)

		controller := browser.NewController(cdpURL)
		state, err := controller.GetState(includeScreenshot)
		if err != nil {
			return mcp.NewToolResultError("Error: " + err.Error()), nil
		}

		output, _ := json.Marshal(state)
		return mcp.NewToolResultText(string(output)), nil
	}
}

// --- browser_get_html ---

func BrowserGetHTMLToolDef() mcp.Tool {
	return mcp.NewTool("browser_get_html",
		mcp.WithDescription("Get the raw HTML of the current page or a specific element by CSS selector."),
		mcp.WithString("selector",
			mcp.Description("Optional CSS selector to get HTML of a specific element. If omitted, returns full page HTML."),
		),
	)
}

func BrowserGetHTMLHandler(cdpURL string) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		selector := request.GetString("selector", "")
		if selector == "" {
			selector = "html"
		}

		controller := browser.NewController(cdpURL)
		html, err := controller.GetHTML(selector)
		if err != nil {
			return mcp.NewToolResultError("Error: " + err.Error()), nil
		}

		if len(html) > 50000 {
			html = html[:50000] + "\n... (HTML truncated)"
		}

		return mcp.NewToolResultText(html), nil
	}
}

// --- browser_screenshot ---

func BrowserScreenshotToolDef() mcp.Tool {
	return mcp.NewTool("browser_screenshot",
		mcp.WithDescription("Take a screenshot of the current page. Returns the screenshot as a base64-encoded PNG image."),
		mcp.WithBoolean("full_page",
			mcp.Description("Whether to capture the full scrollable page or just the visible viewport. Default: false"),
		),
	)
}

func BrowserScreenshotHandler(cdpURL string) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		fullPage := request.GetBool("full_page", false)

		controller := browser.NewController(cdpURL)
		screenshot, err := controller.Screenshot(&browser.ScreenshotOptions{
			Full: fullPage,
		})
		if err != nil {
			return mcp.NewToolResultError("Error: " + err.Error()), nil
		}

		return mcp.NewToolResultText(screenshot), nil
	}
}

// --- browser_scroll ---

func BrowserScrollToolDef() mcp.Tool {
	return mcp.NewTool("browser_scroll",
		mcp.WithDescription("Scroll the page up or down."),
		mcp.WithString("direction",
			mcp.Description("Direction to scroll: 'up' or 'down'. Default: 'down'"),
			mcp.Enum("up", "down"),
		),
	)
}

func BrowserScrollHandler(cdpURL string) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		direction := request.GetString("direction", "down")

		controller := browser.NewController(cdpURL)
		if err := controller.ScrollByDirection(direction); err != nil {
			return mcp.NewToolResultError("Error: " + err.Error()), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Scrolled %s", direction)), nil
	}
}

// --- browser_go_back ---

func BrowserGoBackToolDef() mcp.Tool {
	return mcp.NewTool("browser_go_back",
		mcp.WithDescription("Go back to the previous page in browser history."),
	)
}

func BrowserGoBackHandler(cdpURL string) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		controller := browser.NewController(cdpURL)
		if err := controller.GoBack(); err != nil {
			return mcp.NewToolResultError("Error: " + err.Error()), nil
		}

		return mcp.NewToolResultText("Navigated back"), nil
	}
}

// --- browser_list_tabs ---

func BrowserListTabsToolDef() mcp.Tool {
	return mcp.NewTool("browser_list_tabs",
		mcp.WithDescription("List all open browser tabs."),
	)
}

func BrowserListTabsHandler(cdpURL string) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		controller := browser.NewController(cdpURL)
		tabs, err := controller.ListTabs()
		if err != nil {
			return mcp.NewToolResultError("Error: " + err.Error()), nil
		}

		output, _ := json.Marshal(tabs)
		return mcp.NewToolResultText(string(output)), nil
	}
}

// --- browser_switch_tab ---

func BrowserSwitchTabToolDef() mcp.Tool {
	return mcp.NewTool("browser_switch_tab",
		mcp.WithDescription("Switch to a different browser tab."),
		mcp.WithString("tab_id",
			mcp.Required(),
			mcp.Description("The tab ID to switch to (last 4 characters of target ID, from browser_list_tabs)"),
		),
	)
}

func BrowserSwitchTabHandler(cdpURL string) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		tabID, err := request.RequireString("tab_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		controller := browser.NewController(cdpURL)
		if err := controller.SwitchTab(tabID); err != nil {
			return mcp.NewToolResultError("Error: " + err.Error()), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Switched to tab: %s", tabID)), nil
	}
}

// --- browser_close_tab ---

func BrowserCloseTabToolDef() mcp.Tool {
	return mcp.NewTool("browser_close_tab",
		mcp.WithDescription("Close a browser tab."),
		mcp.WithString("tab_id",
			mcp.Required(),
			mcp.Description("The tab ID to close (last 4 characters of target ID, from browser_list_tabs)"),
		),
	)
}

func BrowserCloseTabHandler(cdpURL string) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		tabID, err := request.RequireString("tab_id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		controller := browser.NewController(cdpURL)
		if err := controller.CloseTab(tabID); err != nil {
			return mcp.NewToolResultError("Error: " + err.Error()), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Closed tab: %s", tabID)), nil
	}
}

// --- browser_evaluate (Go-specific, kept) ---

func BrowserEvaluateToolDef() mcp.Tool {
	return mcp.NewTool("browser_evaluate",
		mcp.WithDescription("Execute JavaScript code in the browser and return the result."),
		mcp.WithString("expression",
			mcp.Required(),
			mcp.Description("JavaScript expression to evaluate"),
		),
	)
}

func BrowserEvaluateHandler(cdpURL string) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		expression, err := request.RequireString("expression")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		controller := browser.NewController(cdpURL)
		result, err := controller.Evaluate(expression)
		if err != nil {
			return mcp.NewToolResultError("Error: " + err.Error()), nil
		}

		output, _ := json.Marshal(result)
		return mcp.NewToolResultText(string(output)), nil
	}
}

// --- browser_wait_visible (Go-specific, kept) ---

func BrowserWaitVisibleToolDef() mcp.Tool {
	return mcp.NewTool("browser_wait_visible",
		mcp.WithDescription("Wait for an element to become visible on the page."),
		mcp.WithString("selector",
			mcp.Required(),
			mcp.Description("CSS selector for the element to wait for"),
		),
	)
}

func BrowserWaitVisibleHandler(cdpURL string) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		selector, err := request.RequireString("selector")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		controller := browser.NewController(cdpURL)
		if err := controller.WaitVisible(selector); err != nil {
			return mcp.NewToolResultError("Error: " + err.Error()), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Element is now visible: %s", selector)), nil
	}
}

// --- browser_pdf (Go-specific, kept) ---

func BrowserPDFToolDef() mcp.Tool {
	return mcp.NewTool("browser_pdf",
		mcp.WithDescription("Generate a PDF of the current page. Returns the PDF as a base64-encoded string."),
	)
}

func BrowserPDFHandler(cdpURL string) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		controller := browser.NewController(cdpURL)
		pdf, err := controller.PDF()
		if err != nil {
			return mcp.NewToolResultError("Error: " + err.Error()), nil
		}

		return mcp.NewToolResultText(pdf), nil
	}
}
