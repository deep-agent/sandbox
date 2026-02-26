package local

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/deep-agent/sandbox/model"
)

func getTestCDPURL() string {
	if u := os.Getenv("TEST_CDP_URL"); u != "" {
		return u
	}
	return "ws://localhost:8080/cdp"
}

func getActualWSURL(cdpURL string) (string, error) {
	parsedURL, err := url.Parse(cdpURL)
	if err != nil {
		return "", fmt.Errorf("failed to parse cdpURL: %w", err)
	}

	scheme := "http"
	if parsedURL.Scheme == "wss" {
		scheme = "https"
	}

	httpURL := fmt.Sprintf("%s://%s%s/json/version", scheme, parsedURL.Host, parsedURL.Path)
	resp, err := http.Get(httpURL)
	if err != nil {
		return cdpURL, nil
	}
	defer resp.Body.Close()

	var result struct {
		WebSocketDebuggerUrl string `json:"webSocketDebuggerUrl"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return cdpURL, nil
	}

	wsURL := result.WebSocketDebuggerUrl
	if strings.HasPrefix(wsURL, "ws://localhost/") {
		wsURL = strings.Replace(wsURL, "ws://localhost/", fmt.Sprintf("ws://%s%s/", parsedURL.Host, parsedURL.Path), 1)
	} else if strings.HasPrefix(wsURL, "ws://localhost:9222/") {
		wsURL = strings.Replace(wsURL, "ws://localhost:9222/", fmt.Sprintf("ws://%s%s/", parsedURL.Host, parsedURL.Path), 1)
	}

	return wsURL, nil
}

func TestNewClient(t *testing.T) {
	workDir := t.TempDir()
	client := NewClient(workDir)

	if client.bashExecutor == nil {
		t.Error("expected bashExecutor to be initialized")
	}
	if client.fileManager == nil {
		t.Error("expected fileManager to be initialized")
	}
	if client.sandboxCtx == nil {
		t.Error("expected sandboxCtx to be initialized")
	}
	if client.sandboxCtx.Workspace != workDir {
		t.Errorf("expected Workspace %s, got %s", workDir, client.sandboxCtx.Workspace)
	}
}

func TestNewClientWithBrowserCDP(t *testing.T) {
	workDir := t.TempDir()
	client := NewClient(workDir, WithBrowserCDP(getTestCDPURL()))

	if client.browserCtrl == nil {
		t.Error("expected browserCtrl to be initialized")
	}
}

func TestNewClientWithSandboxContext(t *testing.T) {
	workDir := t.TempDir()
	ctx := &model.SandboxContext{
		HomeDir: "/home/test",
		OS:      "linux",
		Arch:    "amd64",
	}

	client := NewClient(workDir, WithSandboxContext(ctx))

	if client.sandboxCtx != ctx {
		t.Error("expected sandboxCtx to be set")
	}
}

func TestGetContext(t *testing.T) {
	workDir := t.TempDir()
	client := NewClient(workDir)

	ctx, err := client.GetContext()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ctx.HomeDir != workDir {
		t.Errorf("expected HomeDir %s, got %s", workDir, ctx.HomeDir)
	}
}

func TestBashExec(t *testing.T) {
	workDir := t.TempDir()
	client := NewClient(workDir)

	result, err := client.BashExec(&model.BashExecRequest{
		Command: "echo hello",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result.Output, "hello") {
		t.Errorf("expected output to contain 'hello', got %s", result.Output)
	}
	if result.ExitCode != 0 {
		t.Errorf("expected exit code 0, got %d", result.ExitCode)
	}
}

func TestBashExecWithTimeout(t *testing.T) {
	workDir := t.TempDir()
	client := NewClient(workDir)

	result, err := client.BashExec(&model.BashExecRequest{
		Command:   "echo test",
		TimeoutMS: 5000,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result.Output, "test") {
		t.Errorf("expected output to contain 'test', got %s", result.Output)
	}
}

func TestBashExecWithExitCode(t *testing.T) {
	workDir := t.TempDir()
	client := NewClient(workDir)

	result, err := client.BashExec(&model.BashExecRequest{
		Command: "exit 1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.ExitCode != 1 {
		t.Errorf("expected exit code 1, got %d", result.ExitCode)
	}
}

func TestFileWrite(t *testing.T) {
	workDir := t.TempDir()
	client := NewClient(workDir)

	testFile := filepath.Join(workDir, "test.txt")
	err := client.FileWrite(&model.FileWriteRequest{
		File:    testFile,
		Content: "hello world",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	if string(content) != "hello world" {
		t.Errorf("expected content 'hello world', got %s", string(content))
	}
}

func TestFileWriteBase64(t *testing.T) {
	workDir := t.TempDir()
	client := NewClient(workDir)

	testFile := filepath.Join(workDir, "test_base64.txt")
	originalContent := "hello base64"
	base64Content := base64.StdEncoding.EncodeToString([]byte(originalContent))

	err := client.FileWrite(&model.FileWriteRequest{
		File:    testFile,
		Content: base64Content,
		Base64:  true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	if string(content) != originalContent {
		t.Errorf("expected content '%s', got %s", originalContent, string(content))
	}
}

func TestFileRead(t *testing.T) {
	workDir := t.TempDir()
	client := NewClient(workDir)

	testFile := filepath.Join(workDir, "read_test.txt")
	err := os.WriteFile(testFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	result, err := client.FileRead(&model.FileReadRequest{
		File: testFile,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Content != "test content" {
		t.Errorf("expected content 'test content', got %s", result.Content)
	}
}

func TestFileReadBase64(t *testing.T) {
	workDir := t.TempDir()
	client := NewClient(workDir)

	testFile := filepath.Join(workDir, "read_base64_test.txt")
	originalContent := "test base64 content"
	err := os.WriteFile(testFile, []byte(originalContent), 0644)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	result, err := client.FileRead(&model.FileReadRequest{
		File:   testFile,
		Base64: true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	decoded, err := base64.StdEncoding.DecodeString(result.Content)
	if err != nil {
		t.Fatalf("failed to decode base64: %v", err)
	}

	if string(decoded) != originalContent {
		t.Errorf("expected content '%s', got %s", originalContent, string(decoded))
	}
}

func TestFileList(t *testing.T) {
	workDir := t.TempDir()
	client := NewClient(workDir)

	os.WriteFile(filepath.Join(workDir, "file1.txt"), []byte("1"), 0644)
	os.WriteFile(filepath.Join(workDir, "file2.txt"), []byte("2"), 0644)
	os.Mkdir(filepath.Join(workDir, "subdir"), 0755)

	result, err := client.FileList(&model.FileListRequest{
		Path: workDir,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Files) != 3 {
		t.Errorf("expected 3 files, got %d", len(result.Files))
	}
}

func TestFileDelete(t *testing.T) {
	workDir := t.TempDir()
	client := NewClient(workDir)

	testFile := filepath.Join(workDir, "to_delete.txt")
	os.WriteFile(testFile, []byte("delete me"), 0644)

	err := client.FileDelete(&model.FileDeleteRequest{
		Path: testFile,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(testFile); !os.IsNotExist(err) {
		t.Error("expected file to be deleted")
	}
}

func TestFileMove(t *testing.T) {
	workDir := t.TempDir()
	client := NewClient(workDir)

	srcFile := filepath.Join(workDir, "src.txt")
	dstFile := filepath.Join(workDir, "dst.txt")
	os.WriteFile(srcFile, []byte("move me"), 0644)

	err := client.FileMove(&model.FileMoveRequest{
		Source:      srcFile,
		Destination: dstFile,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := os.Stat(srcFile); !os.IsNotExist(err) {
		t.Error("expected source file to be removed")
	}

	content, err := os.ReadFile(dstFile)
	if err != nil {
		t.Fatalf("failed to read destination file: %v", err)
	}
	if string(content) != "move me" {
		t.Errorf("expected content 'move me', got %s", string(content))
	}
}

func TestFileCopy(t *testing.T) {
	workDir := t.TempDir()
	client := NewClient(workDir)

	srcFile := filepath.Join(workDir, "src_copy.txt")
	dstFile := filepath.Join(workDir, "dst_copy.txt")
	os.WriteFile(srcFile, []byte("copy me"), 0644)

	err := client.FileCopy(&model.FileCopyRequest{
		Source:      srcFile,
		Destination: dstFile,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	srcContent, _ := os.ReadFile(srcFile)
	dstContent, _ := os.ReadFile(dstFile)

	if string(srcContent) != string(dstContent) {
		t.Error("expected source and destination to have same content")
	}
}

func TestMkDir(t *testing.T) {
	workDir := t.TempDir()
	client := NewClient(workDir)

	newDir := filepath.Join(workDir, "new", "nested", "dir")
	err := client.MkDir(&model.MkDirRequest{
		Path: newDir,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	info, err := os.Stat(newDir)
	if err != nil {
		t.Fatalf("failed to stat directory: %v", err)
	}
	if !info.IsDir() {
		t.Error("expected path to be a directory")
	}
}

func TestFileExists(t *testing.T) {
	workDir := t.TempDir()
	client := NewClient(workDir)

	existingFile := filepath.Join(workDir, "existing.txt")
	os.WriteFile(existingFile, []byte("exists"), 0644)

	result, err := client.FileExists(existingFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Exists {
		t.Error("expected file to exist")
	}

	nonExistingFile := filepath.Join(workDir, "non_existing.txt")
	result, err = client.FileExists(nonExistingFile)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Exists {
		t.Error("expected file to not exist")
	}
}

func TestGrepSearch(t *testing.T) {
	workDir := t.TempDir()
	client := NewClient(workDir)

	testFile := filepath.Join(workDir, "grep_test.txt")
	os.WriteFile(testFile, []byte("line one\nfind this line\nline three\n"), 0644)

	result, err := client.GrepSearch(&model.GrepRequest{
		Pattern: "find",
		Path:    workDir,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(result.Output, "find") {
		t.Errorf("expected output to contain 'find', got %s", result.Output)
	}
}

func TestGrepSearchNoMatch(t *testing.T) {
	workDir := t.TempDir()
	client := NewClient(workDir)

	testFile := filepath.Join(workDir, "grep_nomatch.txt")
	os.WriteFile(testFile, []byte("line one\nline two\nline three\n"), 0644)

	result, err := client.GrepSearch(&model.GrepRequest{
		Pattern: "nonexistent",
		Path:    workDir,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Output != "No matches found" {
		t.Logf("output: %s", result.Output)
	}
}

func TestGrepSearchCaseInsensitive(t *testing.T) {
	workDir := t.TempDir()
	client := NewClient(workDir)

	testFile := filepath.Join(workDir, "grep_case.txt")
	os.WriteFile(testFile, []byte("HELLO world\nhello WORLD\n"), 0644)

	result, err := client.GrepSearch(&model.GrepRequest{
		Pattern:         "hello",
		Path:            workDir,
		CaseInsensitive: true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(strings.ToLower(result.Output), "hello") {
		t.Errorf("expected output to contain matches, got %s", result.Output)
	}
}

func newBrowserClient(t *testing.T) *Client {
	workDir := t.TempDir()
	cdpURL := getTestCDPURL()

	actualURL, err := getActualWSURL(cdpURL)
	if err != nil {
		t.Skipf("Failed to get actual WebSocket URL: %v", err)
	}

	client := NewClient(workDir, WithBrowserCDP(actualURL))

	info, err := client.BrowserGetInfo()
	if err != nil {
		t.Skipf("CDP not available at %s: %v", actualURL, err)
	}
	if info.Status != "connected" {
		t.Skipf("CDP not connected: %s", info.Status)
	}

	return client
}

func TestBrowserGetInfo(t *testing.T) {
	client := newBrowserClient(t)

	info, err := client.BrowserGetInfo()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if info.Status != "connected" {
		t.Errorf("expected status 'connected', got %s", info.Status)
	}
	if info.CDPURL == "" {
		t.Error("expected CDPURL to be set")
	}
}

func TestBrowserNavigate(t *testing.T) {
	client := newBrowserClient(t)

	err := client.BrowserNavigate(&model.BrowserNavigateRequest{
		URL: "https://example.com",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBrowserScreenshot(t *testing.T) {
	client := newBrowserClient(t)

	client.BrowserNavigate(&model.BrowserNavigateRequest{URL: "https://example.com"})

	result, err := client.BrowserScreenshot(&model.BrowserScreenshotRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Screenshot == "" {
		t.Error("expected screenshot data")
	}
}

func TestBrowserScreenshotFull(t *testing.T) {
	client := newBrowserClient(t)

	client.BrowserNavigate(&model.BrowserNavigateRequest{URL: "https://example.com"})

	result, err := client.BrowserScreenshot(&model.BrowserScreenshotRequest{
		Full: true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Screenshot == "" {
		t.Error("expected screenshot data")
	}
}

func TestBrowserEvaluate(t *testing.T) {
	client := newBrowserClient(t)

	client.BrowserNavigate(&model.BrowserNavigateRequest{URL: "https://example.com"})

	result, err := client.BrowserEvaluate(&model.BrowserEvaluateRequest{
		Expression: "1 + 1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Result != float64(2) {
		t.Errorf("expected result 2, got %v", result.Result)
	}
}

func TestBrowserEvaluateDocumentTitle(t *testing.T) {
	client := newBrowserClient(t)

	if err := client.BrowserNavigate(&model.BrowserNavigateRequest{URL: "https://example.com"}); err != nil {
		t.Fatalf("navigate failed: %v", err)
	}

	result, err := client.BrowserEvaluate(&model.BrowserEvaluateRequest{
		Expression: "document.title",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	t.Logf("document.title = %v", result.Result)
}

func TestBrowserScroll(t *testing.T) {
	client := newBrowserClient(t)

	client.BrowserNavigate(&model.BrowserNavigateRequest{URL: "https://example.com"})

	err := client.BrowserScroll(&model.BrowserScrollRequest{X: 0, Y: 100})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBrowserGetHTML(t *testing.T) {
	client := newBrowserClient(t)

	if err := client.BrowserNavigate(&model.BrowserNavigateRequest{URL: "https://example.com"}); err != nil {
		t.Fatalf("navigate failed: %v", err)
	}

	result, err := client.BrowserEvaluate(&model.BrowserEvaluateRequest{
		Expression: "document.documentElement.outerHTML.length",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	t.Logf("HTML length via JS: %v", result.Result)
}

func TestBrowserGetCurrentURL(t *testing.T) {
	client := newBrowserClient(t)

	if err := client.BrowserNavigate(&model.BrowserNavigateRequest{URL: "https://example.com"}); err != nil {
		t.Fatalf("navigate failed: %v", err)
	}

	result, err := client.BrowserGetCurrentURL()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	t.Logf("current URL: %s", result.URL)
}

func TestBrowserGetTitle(t *testing.T) {
	client := newBrowserClient(t)

	if err := client.BrowserNavigate(&model.BrowserNavigateRequest{URL: "https://example.com"}); err != nil {
		t.Fatalf("navigate failed: %v", err)
	}

	result, err := client.BrowserGetTitle()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	t.Logf("title: %s", result.Title)
}

func TestBrowserGetPageInfo(t *testing.T) {
	client := newBrowserClient(t)

	if err := client.BrowserNavigate(&model.BrowserNavigateRequest{URL: "https://example.com"}); err != nil {
		t.Fatalf("navigate failed: %v", err)
	}

	info, err := client.BrowserGetPageInfo()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	t.Logf("page info: URL=%s, Title=%s", info.URL, info.Title)
}

func TestBrowserPDF(t *testing.T) {
	client := newBrowserClient(t)

	client.BrowserNavigate(&model.BrowserNavigateRequest{URL: "https://example.com"})

	result, err := client.BrowserPDF()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.PDF == "" {
		t.Error("expected PDF data")
	}
}

func TestBrowserNotInitialized(t *testing.T) {
	workDir := t.TempDir()
	client := NewClient(workDir)

	_, err := client.BrowserGetInfo()
	if err == nil {
		t.Error("expected error when browser not initialized")
	}

	if !strings.Contains(err.Error(), "browser controller not initialized") {
		t.Errorf("unexpected error message: %v", err)
	}
}
