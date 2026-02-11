package model

type FileReadRequest struct {
	File   string `json:"file" vd:"len($)>0"`
	Base64 bool   `json:"base64,omitempty"`
}

type FileReadResult struct {
	Content string `json:"content"`
}

type FileWriteRequest struct {
	File    string `json:"file" vd:"len($)>0"`
	Content string `json:"content" vd:"len($)>0"`
	Base64  bool   `json:"base64,omitempty"`
}

type FileListRequest struct {
	Path string `json:"path" vd:"len($)>0"`
}

type FileInfo struct {
	Name    string `json:"name"`
	Path    string `json:"path"`
	Size    int64  `json:"size"`
	IsDir   bool   `json:"is_dir"`
	Mode    string `json:"mode"`
	ModTimeUnix int64  `json:"mod_time_unix"`
}

type FileListResult struct {
	Files []FileInfo `json:"files"`
}

type FileDeleteRequest struct {
	Path string `json:"path" vd:"len($)>0"`
}

type FileMoveRequest struct {
	Source      string `json:"source" vd:"len($)>0"`
	Destination string `json:"destination" vd:"len($)>0"`
}

type FileCopyRequest struct {
	Source      string `json:"source" vd:"len($)>0"`
	Destination string `json:"destination" vd:"len($)>0"`
}

type MkDirRequest struct {
	Path string `json:"path" vd:"len($)>0"`
}

type FileExistsResult struct {
	Exists bool `json:"exists"`
}
