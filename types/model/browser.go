package model

type BrowserInfo struct {
	CDPURL    string `json:"cdp_url"`
	WebSocket string `json:"websocket_url"`
	Status    string `json:"status"`
}

type BrowserNavigateRequest struct {
	URL string `json:"url" vd:"len($)>0"`
}

type BrowserScreenshotRequest struct {
	Format  string `json:"format,omitempty"`
	Quality int    `json:"quality,omitempty"`
	Full    bool   `json:"full,omitempty"`
}

type BrowserScreenshotResult struct {
	Screenshot string `json:"screenshot"`
}

type BrowserClickRequest struct {
	Selector string `json:"selector" vd:"len($)>0"`
}

type BrowserTypeRequest struct {
	Selector string `json:"selector" vd:"len($)>0"`
	Text     string `json:"text" vd:"len($)>0"`
}

type BrowserEvaluateRequest struct {
	Expression string `json:"expression" vd:"len($)>0"`
}

type BrowserEvaluateResult struct {
	Result interface{} `json:"result"`
}

type BrowserScrollRequest struct {
	X int64 `json:"x"`
	Y int64 `json:"y"`
}

type BrowserGetHTMLRequest struct {
	Selector string `json:"selector" vd:"len($)>0"`
}

type BrowserGetHTMLResult struct {
	HTML string `json:"html"`
}

type BrowserWaitVisibleRequest struct {
	Selector string `json:"selector" vd:"len($)>0"`
}

type BrowserURLResult struct {
	URL string `json:"url"`
}

type BrowserTitleResult struct {
	Title string `json:"title"`
}

type BrowserPDFResult struct {
	PDF string `json:"pdf"`
}

type BrowserPageInfo struct {
	URL    string `json:"url"`
	Title  string `json:"title"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}
