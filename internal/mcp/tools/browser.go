package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/deep-agent/sandbox/internal/services/browser"
	"github.com/mark3labs/mcp-go/mcp"
)

func BrowserNavigateToolDef() mcp.Tool {
	return mcp.NewTool("browser_navigate",
		mcp.WithDescription("Navigate the browser to a specified URL. Use this to open web pages for browsing, testing, or scraping."),
		mcp.WithString("url",
			mcp.Required(),
			mcp.Description("The URL to navigate to (e.g., 'https://example.com')"),
		),
	)
}

func BrowserNavigateHandler(cdpURL string) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		url, err := request.RequireString("url")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		controller := browser.NewController(cdpURL)
		if err := controller.Navigate(url); err != nil {
			return mcp.NewToolResultError("Error: " + err.Error()), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Successfully navigated to: %s", url)), nil
	}
}

func BrowserScreenshotToolDef() mcp.Tool {
	return mcp.NewTool("browser_screenshot",
		mcp.WithDescription("Capture a screenshot of the current browser page. Returns the screenshot as a base64-encoded PNG image."),
		mcp.WithBoolean("full_page",
			mcp.Description("If true, capture the full scrollable page. If false, capture only the visible viewport. Default: false"),
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

func BrowserClickToolDef() mcp.Tool {
	return mcp.NewTool("browser_click",
		mcp.WithDescription("Click on an element in the browser page using a CSS selector. The element must be visible."),
		mcp.WithString("selector",
			mcp.Required(),
			mcp.Description("CSS selector for the element to click (e.g., '#submit-button', '.nav-link', 'button[type=submit]')"),
		),
	)
}

func BrowserClickHandler(cdpURL string) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		selector, err := request.RequireString("selector")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		controller := browser.NewController(cdpURL)
		if err := controller.Click(selector); err != nil {
			return mcp.NewToolResultError("Error: " + err.Error()), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Successfully clicked element: %s", selector)), nil
	}
}

func BrowserTypeToolDef() mcp.Tool {
	return mcp.NewTool("browser_type",
		mcp.WithDescription("Type text into an input element in the browser. First clicks the element, then types the text."),
		mcp.WithString("selector",
			mcp.Required(),
			mcp.Description("CSS selector for the input element (e.g., '#search-input', 'input[name=email]')"),
		),
		mcp.WithString("text",
			mcp.Required(),
			mcp.Description("The text to type into the element"),
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

		return mcp.NewToolResultText(fmt.Sprintf("Successfully typed text into: %s", selector)), nil
	}
}

func BrowserGetURLToolDef() mcp.Tool {
	return mcp.NewTool("browser_get_url",
		mcp.WithDescription("Get the current URL of the browser page."),
	)
}

func BrowserGetURLHandler(cdpURL string) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		controller := browser.NewController(cdpURL)
		url, err := controller.GetCurrentURL()
		if err != nil {
			return mcp.NewToolResultError("Error: " + err.Error()), nil
		}

		return mcp.NewToolResultText(url), nil
	}
}

func BrowserGetTitleToolDef() mcp.Tool {
	return mcp.NewTool("browser_get_title",
		mcp.WithDescription("Get the title of the current browser page."),
	)
}

func BrowserGetTitleHandler(cdpURL string) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		controller := browser.NewController(cdpURL)
		title, err := controller.GetTitle()
		if err != nil {
			return mcp.NewToolResultError("Error: " + err.Error()), nil
		}

		return mcp.NewToolResultText(title), nil
	}
}

func BrowserGetHTMLToolDef() mcp.Tool {
	return mcp.NewTool("browser_get_html",
		mcp.WithDescription("Get the outer HTML of an element matching the CSS selector."),
		mcp.WithString("selector",
			mcp.Required(),
			mcp.Description("CSS selector for the element (e.g., '#content', 'body', '.main-content')"),
		),
	)
}

func BrowserGetHTMLHandler(cdpURL string) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		selector, err := request.RequireString("selector")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
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

func BrowserEvaluateToolDef() mcp.Tool {
	return mcp.NewTool("browser_evaluate",
		mcp.WithDescription("Execute JavaScript code in the browser and return the result. Use this for custom DOM manipulation or data extraction."),
		mcp.WithString("expression",
			mcp.Required(),
			mcp.Description("JavaScript expression to evaluate (e.g., 'document.querySelectorAll(\"a\").length')"),
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

func BrowserScrollToolDef() mcp.Tool {
	return mcp.NewTool("browser_scroll",
		mcp.WithDescription("Scroll the browser page to a specific position."),
		mcp.WithNumber("x",
			mcp.Description("Horizontal scroll position in pixels. Default: 0"),
		),
		mcp.WithNumber("y",
			mcp.Required(),
			mcp.Description("Vertical scroll position in pixels"),
		),
	)
}

func BrowserScrollHandler(cdpURL string) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		x := int64(request.GetFloat("x", 0))
		y := int64(request.GetFloat("y", 0))

		controller := browser.NewController(cdpURL)
		if err := controller.Scroll(x, y); err != nil {
			return mcp.NewToolResultError("Error: " + err.Error()), nil
		}

		return mcp.NewToolResultText(fmt.Sprintf("Scrolled to position: (%d, %d)", x, y)), nil
	}
}

func BrowserWaitVisibleToolDef() mcp.Tool {
	return mcp.NewTool("browser_wait_visible",
		mcp.WithDescription("Wait for an element to become visible on the page. Useful after navigation or dynamic content loading."),
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

func BrowserGetPageInfoToolDef() mcp.Tool {
	return mcp.NewTool("browser_get_page_info",
		mcp.WithDescription("Get information about the current page including URL and title."),
	)
}

func BrowserGetPageInfoHandler(cdpURL string) func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		controller := browser.NewController(cdpURL)
		info, err := controller.GetPageInfo()
		if err != nil {
			return mcp.NewToolResultError("Error: " + err.Error()), nil
		}

		output, _ := json.Marshal(info)
		return mcp.NewToolResultText(string(output)), nil
	}
}

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
