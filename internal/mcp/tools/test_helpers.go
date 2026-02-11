package tools

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func mockCallToolRequest(args map[string]interface{}) mcp.CallToolRequest {
	return mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: args,
		},
	}
}

func getTextContent(result *mcp.CallToolResult) string {
	if result == nil || len(result.Content) == 0 {
		return ""
	}
	for _, c := range result.Content {
		if textContent, ok := c.(mcp.TextContent); ok {
			return textContent.Text
		}
	}
	return ""
}
