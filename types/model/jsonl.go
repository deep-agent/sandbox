package model

type JSONLCountRequest struct {
	File string `json:"file" vd:"len($)>0"`
}

type JSONLCountResult struct {
	Lines int `json:"lines"`
}

type JSONLReadRequest struct {
	File      string `json:"file" vd:"len($)>0"`
	StartLine int    `json:"start_line" vd:"$>=0"`
	Count     *int   `json:"count"`
}

type JSONLReadResult struct {
	Lines []string `json:"lines"`
}

type JSONLAppendRequest struct {
	File       string   `json:"file" vd:"len($)>0"`
	JSONString []string `json:"json_string" vd:"len($)>0"`
}
