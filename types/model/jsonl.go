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
	Count     int    `json:"count" vd:"$>0"`
}

type JSONLReadResult struct {
	Lines []string `json:"lines"`
}
