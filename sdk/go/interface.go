package sandbox

import (
	"io"

	"github.com/deep-agent/sandbox/types/model"
)

type Sandbox interface {
	ContextProvider
	BashExecutor
	FileManager
	GrepSearcher
	WebClient
	JSONLReader
}

type ContextProvider interface {
	GetContext() (*model.SandboxContext, error)
}

type BashExecutor interface {
	BashExec(req *model.BashExecRequest) (*model.BashExecResult, error)
}

type FileManager interface {
	FileRead(req *model.FileReadRequest) (*model.FileReadResult, error)
	FileWrite(req *model.FileWriteRequest) error
	FileList(req *model.FileListRequest) (*model.FileListResult, error)
	FileDelete(req *model.FileDeleteRequest) error
	FileMove(req *model.FileMoveRequest) error
	FileCopy(req *model.FileCopyRequest) error
	MkDir(req *model.MkDirRequest) error
	FileExists(path string) (*model.FileExistsResult, error)
	FileUpload(filename string, reader io.Reader, destPath string) (*model.FileUploadResult, error)
	FileDownload(filePath string) (io.ReadCloser, string, error) // returns body, contentType, error
}

type GrepSearcher interface {
	GrepSearch(req *model.GrepRequest) (*model.GrepResult, error)
}

type WebClient interface {
	WebFetch(req *model.WebFetchRequest) (*model.WebFetchResult, error)
	WebSearch(req *model.WebSearchRequest) (*model.WebSearchResult, error)
}

type JSONLReader interface {
	JSONLCountLines(req *model.JSONLCountRequest) (*model.JSONLCountResult, error)
	JSONLReadLines(req *model.JSONLReadRequest) (*model.JSONLReadResult, error)
	JSONLAppendLine(req *model.JSONLAppendRequest) error
}
