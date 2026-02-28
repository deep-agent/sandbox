package sandbox

import "github.com/deep-agent/sandbox/types/model"

type Sandbox interface {
	ContextProvider
	BashExecutor
	FileManager
	GrepSearcher
	BrowserController
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
}

type GrepSearcher interface {
	GrepSearch(req *model.GrepRequest) (*model.GrepResult, error)
}

type BrowserController interface {
	BrowserGetInfo() (*model.BrowserInfo, error)
	BrowserNavigate(req *model.BrowserNavigateRequest) error
	BrowserScreenshot(req *model.BrowserScreenshotRequest) (*model.BrowserScreenshotResult, error)
	BrowserClick(req *model.BrowserClickRequest) error
	BrowserType(req *model.BrowserTypeRequest) error
	BrowserEvaluate(req *model.BrowserEvaluateRequest) (*model.BrowserEvaluateResult, error)
	BrowserScroll(req *model.BrowserScrollRequest) error
	BrowserGetHTML(req *model.BrowserGetHTMLRequest) (*model.BrowserGetHTMLResult, error)
	BrowserWaitVisible(req *model.BrowserWaitVisibleRequest) error
	BrowserGetCurrentURL() (*model.BrowserURLResult, error)
	BrowserGetTitle() (*model.BrowserTitleResult, error)
	BrowserGetPageInfo() (*model.BrowserPageInfo, error)
	BrowserPDF() (*model.BrowserPDFResult, error)
}
