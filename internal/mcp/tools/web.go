package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/deep-agent/sandbox/internal/services/web"
	"github.com/mark3labs/mcp-go/mcp"
)

func WebFetchToolDef() mcp.Tool {
	return mcp.NewTool("WebFetch",
		mcp.WithDescription(`IMPORTANT: WebFetch WILL FAIL for authenticated or private URLs. Before using this tool, check if the URL points to an authenticated service (e.g. Google Docs, Confluence, Jira, GitHub). If so, you MUST use ToolSearch first to find a specialized tool that provides authenticated access.

- Fetches content from a specified URL and processes it using an AI model
- Takes a URL and a prompt as input
- Fetches the URL content, converts HTML to markdown
- Processes the content with the prompt using a small, fast model
- Returns the model's response about the content
- Use this tool when you need to retrieve and analyze web content

Usage notes:
  - IMPORTANT: If an MCP-provided web fetch tool is available, prefer using that tool instead of this one, as it may have fewer restrictions.
  - The URL must be a fully-formed valid URL
  - HTTP URLs will be automatically upgraded to HTTPS
  - The prompt should describe what information you want to extract from the page
  - This tool is read-only and does not modify any files
  - Results may be summarized if the content is very large
  - Includes a self-cleaning 15-minute cache for faster responses when repeatedly accessing the same URL
  - When a URL redirects to a different host, the tool will inform you and provide the redirect URL in a special format. You should then make a new WebFetch request with the redirect URL to fetch the content.
  - For GitHub URLs, prefer using the gh CLI via Bash instead (e.g., gh pr view, gh issue view, gh api).`),
		mcp.WithString("url",
			mcp.Required(),
			mcp.Description("The URL to fetch content from"),
		),
		mcp.WithString("prompt",
			mcp.Required(),
			mcp.Description("The prompt to run on the fetched content"),
		),
	)
}

func WebFetchHandler() func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	fetcher := web.NewFetcher()

	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		url, err := request.RequireString("url")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		prompt, err := request.RequireString("prompt")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		result, err := fetcher.Fetch(ctx, url)
		if err != nil {
			return mcp.NewToolResultError("Error fetching URL: " + err.Error()), nil
		}

		if result.RedirectURL != "" {
			return mcp.NewToolResultText(fmt.Sprintf("The URL redirected to a different host. Please make a new request with the redirect URL:\n\nRedirect URL: %s", result.RedirectURL)), nil
		}

		content := result.Content
		if len(content) > 100000 {
			content = content[:100000] + "\n\n[Content truncated due to length...]"
		}

		output := fmt.Sprintf("URL: %s\nPrompt: %s\n\n---\n\nContent:\n%s", url, prompt, content)
		return mcp.NewToolResultText(output), nil
	}
}

func WebSearchToolDef() mcp.Tool {
	return mcp.NewTool("WebSearch",
		mcp.WithDescription(`- Allows Claude to search the web and use the results to inform responses
- Provides up-to-date information for current events and recent data
- Returns search result information formatted as search result blocks, including links as markdown hyperlinks
- Use this tool for accessing information beyond Claude's knowledge cutoff
- Searches are performed automatically within a single API call

CRITICAL REQUIREMENT - You MUST follow this:
  - After answering the user's question, you MUST include a "Sources:" section at the end of your response
  - In the Sources section, list all relevant URLs from the search results as markdown hyperlinks: [Title](URL)
  - This is MANDATORY - never skip including sources in your response
  - Example format:

    [Your answer here]

    Sources:
    - [Source Title 1](https://example.com/1)
    - [Source Title 2](https://example.com/2)

Usage notes:
  - Domain filtering is supported to include or block specific websites
  - Web search is only available in the US`),
		mcp.WithString("query",
			mcp.Required(),
			mcp.Description("The search query to use"),
		),
		mcp.WithString("allowed_domains",
			mcp.Description("Comma-separated list of domains to include in search results"),
		),
		mcp.WithString("blocked_domains",
			mcp.Description("Comma-separated list of domains to exclude from search results"),
		),
		mcp.WithNumber("num",
			mcp.Description("Maximum number of search results to return (default: 5, max: 10)"),
		),
		mcp.WithString("lr",
			mcp.Description("Language restriction for search results (e.g., 'lang_en' for English)"),
		),
	)
}

func WebSearchHandler() func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	searcher := web.NewSearcher()

	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		query, err := request.RequireString("query")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		var allowedDomains []string
		if domains := request.GetString("allowed_domains", ""); domains != "" {
			allowedDomains = strings.Split(domains, ",")
			for i := range allowedDomains {
				allowedDomains[i] = strings.TrimSpace(allowedDomains[i])
			}
		}

		var blockedDomains []string
		if domains := request.GetString("blocked_domains", ""); domains != "" {
			blockedDomains = strings.Split(domains, ",")
			for i := range blockedDomains {
				blockedDomains[i] = strings.TrimSpace(blockedDomains[i])
			}
		}

		opts := web.SearchOptions{
			Query:          query,
			AllowedDomains: allowedDomains,
			BlockedDomains: blockedDomains,
			NumResults:     int(request.GetFloat("num", 5)),
			Language:       request.GetString("lr", ""),
		}

		result, err := searcher.Search(ctx, opts)
		if err != nil {
			return mcp.NewToolResultError("Error searching: " + err.Error()), nil
		}

		if len(result.Results) == 0 {
			return mcp.NewToolResultText("No search results found for query: " + query), nil
		}

		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("Search results for: %s\n\n", query))
		for i, r := range result.Results {
			sb.WriteString(fmt.Sprintf("## %d. [%s](%s)\n", i+1, r.Title, r.URL))
			sb.WriteString(fmt.Sprintf("%s\n\n", r.Snippet))
		}

		return mcp.NewToolResultText(sb.String()), nil
	}
}
