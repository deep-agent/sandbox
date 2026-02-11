package model

type WebFetchRequest struct {
	URL    string `json:"url" vd:"len($)>0"`
	Prompt string `json:"prompt" vd:"len($)>0"`
}

type WebFetchResult struct {
	Content string `json:"content"`
}

type WebSearchRequest struct {
	Query          string   `json:"query" vd:"len($)>1"`
	AllowedDomains []string `json:"allowed_domains,omitempty"`
	BlockedDomains []string `json:"blocked_domains,omitempty"`
	NumResults     int      `json:"num,omitempty"`
	Language       string   `json:"lr,omitempty"`
}

type WebSearchResultItem struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	Snippet string `json:"snippet"`
}

type WebSearchResult struct {
	Results []WebSearchResultItem `json:"results"`
}
