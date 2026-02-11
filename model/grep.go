package model

type GrepRequest struct {
	Pattern         string `json:"pattern" vd:"len($)>0"`
	Path            string `json:"path,omitempty"`
	Glob            string `json:"glob,omitempty"`
	CaseInsensitive bool   `json:"case_insensitive,omitempty"`
	ContextLines    int    `json:"context_lines,omitempty"`
	OutputMode      string `json:"output_mode,omitempty"`
	Limit           int    `json:"limit,omitempty"`
	MaxLineLength   int    `json:"max_line_length,omitempty"`
}

type GrepResult struct {
	Output    string `json:"output"`
	Truncated bool   `json:"truncated"`
}
