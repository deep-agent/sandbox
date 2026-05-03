package sandbox

import (
	"context"
	"io"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/deep-agent/sandbox/types/model"
)

type Sandbox interface {
	ContextProvider
	BashExecutor
	FileManager
	GrepSearcher
	WebClient
	JSONLReader
	MCPToolProvider
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
	FileCreateTemp(req *model.FileCreateTempRequest) (*model.FileCreateTempResult, error)
	FileGlob(req *model.FileGlobRequest) (*model.FileGlobResult, error)
	FileEvalSymlinks(req *model.FileEvalSymlinksRequest) (*model.FileEvalSymlinksResult, error)
	FileAppend(req *model.FileAppendRequest) error
	FileStat(req *model.FileStatRequest) (*model.FileStatResult, error)
	TempDir() (*model.TempDirResult, error)
	UserHomeDir() (*model.UserHomeDirResult, error)
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

type MCPToolProvider interface {
	MCPTools(ctx context.Context, opts ...MCPOption) ([]tool.BaseTool, error)
}

// MCPConfig holds the unified configuration for MCPTools across implementations.
// Fields irrelevant to a given implementation are ignored.
type MCPConfig struct {
	ToolNames    []string
	ExtraHeaders map[string]string
	InitTimeout  time.Duration
}

type MCPOption func(*MCPConfig)

func WithMCPToolNames(names ...string) MCPOption {
	return func(c *MCPConfig) { c.ToolNames = append(c.ToolNames, names...) }
}

// WithMCPHeader adds an HTTP header to MCP requests. Honored by transports
// that speak HTTP; ignored by the in-process local client.
func WithMCPHeader(key, value string) MCPOption {
	return func(c *MCPConfig) {
		if c.ExtraHeaders == nil {
			c.ExtraHeaders = make(map[string]string)
		}
		c.ExtraHeaders[key] = value
	}
}

// WithMCPInitTimeout bounds the MCP session initialization handshake.
// Honored by transports with a handshake; ignored by the local client.
func WithMCPInitTimeout(d time.Duration) MCPOption {
	return func(c *MCPConfig) { c.InitTimeout = d }
}
