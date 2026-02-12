package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/deep-agent/sandbox/model"
)

const testBaseURL = "http://localhost:8080"

func TestNewClient(t *testing.T) {
	client := NewClient(testBaseURL)

	if client.baseURL != testBaseURL {
		t.Errorf("expected baseURL %s, got %s", testBaseURL, client.baseURL)
	}

	if client.httpClient == nil {
		t.Error("expected httpClient to be initialized")
	}

	if client.httpClient.Timeout != 30*time.Second {
		t.Errorf("expected default timeout 30s, got %v", client.httpClient.Timeout)
	}
}

func TestNewClientWithOptions(t *testing.T) {
	client := NewClient(testBaseURL,
		WithTimeout(60*time.Second),
		WithSecret("test-secret"),
	)

	if client.httpClient.Timeout != 60*time.Second {
		t.Errorf("expected timeout 60s, got %v", client.httpClient.Timeout)
	}

	if client.tokenProvider == nil {
		t.Error("expected tokenProvider to be set")
	}
}

func TestWithSecretFromEnv(t *testing.T) {
	os.Setenv("TEST_SECRET", "my-secret")
	defer os.Unsetenv("TEST_SECRET")

	client := NewClient(testBaseURL, WithSecretFromEnv("TEST_SECRET"))

	if client.tokenProvider == nil {
		t.Error("expected tokenProvider to be set")
	}

	token, err := client.tokenProvider()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if token == "" {
		t.Error("expected non-empty token")
	}
}

func TestWithSecretFromEnvEmpty(t *testing.T) {
	os.Unsetenv("EMPTY_SECRET")

	client := NewClient(testBaseURL, WithSecretFromEnv("EMPTY_SECRET"))

	token, err := client.tokenProvider()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if token != "" {
		t.Error("expected empty token when env is not set")
	}
}

func TestGetContext(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/sandbox" {
			t.Errorf("expected path /v1/sandbox, got %s", r.URL.Path)
		}
		if r.Method != "GET" {
			t.Errorf("expected method GET, got %s", r.Method)
		}

		resp := map[string]interface{}{
			"code": 0,
			"data": map[string]interface{}{
				"home_dir":  "/home/sandbox",
				"workspace": "/home/sandbox/workspace",
				"os":        "linux",
				"arch":      "amd64",
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	ctx, err := client.GetContext()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if ctx.HomeDir != "/home/sandbox" {
		t.Errorf("expected HomeDir /home/sandbox, got %s", ctx.HomeDir)
	}
	if ctx.Workspace != "/home/sandbox/workspace" {
		t.Errorf("expected Workspace /home/sandbox/workspace, got %s", ctx.Workspace)
	}
}

func TestBashExec(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/bash/exec" {
			t.Errorf("expected path /v1/bash/exec, got %s", r.URL.Path)
		}
		if r.Method != "POST" {
			t.Errorf("expected method POST, got %s", r.Method)
		}

		var req model.BashExecRequest
		json.NewDecoder(r.Body).Decode(&req)

		if req.Command != "echo hello" {
			t.Errorf("expected command 'echo hello', got %s", req.Command)
		}

		resp := map[string]interface{}{
			"code": 0,
			"data": map[string]interface{}{
				"output":    "hello\n",
				"exit_code": 0,
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	result, err := client.BashExec(&model.BashExecRequest{
		Command: "echo hello",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Output != "hello\n" {
		t.Errorf("expected output 'hello\\n', got %s", result.Output)
	}
	if result.ExitCode != 0 {
		t.Errorf("expected exit code 0, got %d", result.ExitCode)
	}
}

func TestFileRead(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/file/read" {
			t.Errorf("expected path /v1/file/read, got %s", r.URL.Path)
		}

		var req model.FileReadRequest
		json.NewDecoder(r.Body).Decode(&req)

		if req.File != "/tmp/test.txt" {
			t.Errorf("expected file /tmp/test.txt, got %s", req.File)
		}

		resp := map[string]interface{}{
			"code": 0,
			"data": map[string]interface{}{
				"content": "file content",
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	result, err := client.FileRead(&model.FileReadRequest{
		File: "/tmp/test.txt",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Content != "file content" {
		t.Errorf("expected content 'file content', got %s", result.Content)
	}
}

func TestFileWrite(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/file/write" {
			t.Errorf("expected path /v1/file/write, got %s", r.URL.Path)
		}

		var req model.FileWriteRequest
		json.NewDecoder(r.Body).Decode(&req)

		if req.File != "/tmp/test.txt" {
			t.Errorf("expected file /tmp/test.txt, got %s", req.File)
		}
		if req.Content != "new content" {
			t.Errorf("expected content 'new content', got %s", req.Content)
		}

		resp := map[string]interface{}{
			"code": 0,
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	err := client.FileWrite(&model.FileWriteRequest{
		File:    "/tmp/test.txt",
		Content: "new content",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestFileList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/file/list" {
			t.Errorf("expected path /v1/file/list, got %s", r.URL.Path)
		}

		resp := map[string]interface{}{
			"code": 0,
			"data": map[string]interface{}{
				"files": []map[string]interface{}{
					{"name": "file1.txt", "path": "/tmp/file1.txt", "is_dir": false},
					{"name": "dir1", "path": "/tmp/dir1", "is_dir": true},
				},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	result, err := client.FileList(&model.FileListRequest{
		Path: "/tmp",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Files) != 2 {
		t.Errorf("expected 2 files, got %d", len(result.Files))
	}
}

func TestFileDelete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/file/delete" {
			t.Errorf("expected path /v1/file/delete, got %s", r.URL.Path)
		}

		resp := map[string]interface{}{"code": 0}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	err := client.FileDelete(&model.FileDeleteRequest{Path: "/tmp/test.txt"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestFileMove(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/file/move" {
			t.Errorf("expected path /v1/file/move, got %s", r.URL.Path)
		}

		var req model.FileMoveRequest
		json.NewDecoder(r.Body).Decode(&req)

		if req.Source != "/tmp/old.txt" || req.Destination != "/tmp/new.txt" {
			t.Errorf("unexpected source/destination")
		}

		resp := map[string]interface{}{"code": 0}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	err := client.FileMove(&model.FileMoveRequest{
		Source:      "/tmp/old.txt",
		Destination: "/tmp/new.txt",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestFileCopy(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/file/copy" {
			t.Errorf("expected path /v1/file/copy, got %s", r.URL.Path)
		}

		resp := map[string]interface{}{"code": 0}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	err := client.FileCopy(&model.FileCopyRequest{
		Source:      "/tmp/src.txt",
		Destination: "/tmp/dst.txt",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMkDir(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/file/mkdir" {
			t.Errorf("expected path /v1/file/mkdir, got %s", r.URL.Path)
		}

		resp := map[string]interface{}{"code": 0}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	err := client.MkDir(&model.MkDirRequest{Path: "/tmp/newdir"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestFileExists(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/file/exists" {
			t.Errorf("expected path /v1/file/exists, got %s", r.URL.Path)
		}

		resp := map[string]interface{}{
			"code": 0,
			"data": map[string]interface{}{
				"exists": true,
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	result, err := client.FileExists("/tmp/test.txt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.Exists {
		t.Error("expected exists to be true")
	}
}

func TestGrepSearch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/grep/search" {
			t.Errorf("expected path /v1/grep/search, got %s", r.URL.Path)
		}

		var req model.GrepRequest
		json.NewDecoder(r.Body).Decode(&req)

		if req.Pattern != "test" {
			t.Errorf("expected pattern 'test', got %s", req.Pattern)
		}

		resp := map[string]interface{}{
			"code": 0,
			"data": map[string]interface{}{
				"output":    "file.txt:1:test line",
				"truncated": false,
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	result, err := client.GrepSearch(&model.GrepRequest{
		Pattern: "test",
		Path:    "/tmp",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Output != "file.txt:1:test line" {
		t.Errorf("unexpected output: %s", result.Output)
	}
}

func TestBrowserGetInfo(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/browser/info" {
			t.Errorf("expected path /v1/browser/info, got %s", r.URL.Path)
		}

		resp := map[string]interface{}{
			"code": 0,
			"data": map[string]interface{}{
				"cdp_url":       "ws://localhost:9222",
				"websocket_url": "ws://localhost:9222",
				"status":        "connected",
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	info, err := client.BrowserGetInfo()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if info.Status != "connected" {
		t.Errorf("expected status 'connected', got %s", info.Status)
	}
}

func TestBrowserNavigate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/browser/navigate" {
			t.Errorf("expected path /v1/browser/navigate, got %s", r.URL.Path)
		}

		var req model.BrowserNavigateRequest
		json.NewDecoder(r.Body).Decode(&req)

		if req.URL != "https://example.com" {
			t.Errorf("expected URL 'https://example.com', got %s", req.URL)
		}

		resp := map[string]interface{}{"code": 0}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	err := client.BrowserNavigate(&model.BrowserNavigateRequest{
		URL: "https://example.com",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBrowserScreenshot(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/browser/screenshot" {
			t.Errorf("expected path /v1/browser/screenshot, got %s", r.URL.Path)
		}

		resp := map[string]interface{}{
			"code": 0,
			"data": map[string]interface{}{
				"screenshot": "base64encodedimage",
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	result, err := client.BrowserScreenshot(&model.BrowserScreenshotRequest{
		Full: true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Screenshot != "base64encodedimage" {
		t.Errorf("unexpected screenshot data")
	}
}

func TestBrowserClick(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{"code": 0}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	err := client.BrowserClick(&model.BrowserClickRequest{Selector: "#btn"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBrowserType(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{"code": 0}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	err := client.BrowserType(&model.BrowserTypeRequest{
		Selector: "#input",
		Text:     "hello",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBrowserEvaluate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"code": 0,
			"data": map[string]interface{}{
				"result": "evaluated result",
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	result, err := client.BrowserEvaluate(&model.BrowserEvaluateRequest{
		Expression: "document.title",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Result != "evaluated result" {
		t.Errorf("unexpected result: %v", result.Result)
	}
}

func TestBrowserScroll(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{"code": 0}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	err := client.BrowserScroll(&model.BrowserScrollRequest{X: 0, Y: 100})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBrowserGetHTML(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"code": 0,
			"data": map[string]interface{}{
				"html": "<div>content</div>",
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	result, err := client.BrowserGetHTML(&model.BrowserGetHTMLRequest{Selector: "body"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.HTML != "<div>content</div>" {
		t.Errorf("unexpected HTML: %s", result.HTML)
	}
}

func TestBrowserWaitVisible(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{"code": 0}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	err := client.BrowserWaitVisible(&model.BrowserWaitVisibleRequest{Selector: "#element"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBrowserGetCurrentURL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"code": 0,
			"data": map[string]interface{}{
				"url": "https://example.com/page",
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	result, err := client.BrowserGetCurrentURL()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.URL != "https://example.com/page" {
		t.Errorf("unexpected URL: %s", result.URL)
	}
}

func TestBrowserGetTitle(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"code": 0,
			"data": map[string]interface{}{
				"title": "Page Title",
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	result, err := client.BrowserGetTitle()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Title != "Page Title" {
		t.Errorf("unexpected title: %s", result.Title)
	}
}

func TestBrowserGetPageInfo(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"code": 0,
			"data": map[string]interface{}{
				"url":    "https://example.com",
				"title":  "Example",
				"width":  1920,
				"height": 1080,
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	info, err := client.BrowserGetPageInfo()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if info.URL != "https://example.com" {
		t.Errorf("unexpected URL: %s", info.URL)
	}
	if info.Title != "Example" {
		t.Errorf("unexpected title: %s", info.Title)
	}
}

func TestBrowserPDF(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"code": 0,
			"data": map[string]interface{}{
				"pdf": "base64pdfdata",
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	result, err := client.BrowserPDF()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.PDF != "base64pdfdata" {
		t.Errorf("unexpected PDF data")
	}
}

func TestAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"code":    1,
			"message": "something went wrong",
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	_, err := client.GetContext()
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err.Error() != "API error: something went wrong" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestAuthorizationHeader(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth == "" {
			t.Error("expected Authorization header")
		}
		if len(auth) < 7 || auth[:7] != "Bearer " {
			t.Errorf("expected Bearer token, got: %s", auth)
		}

		resp := map[string]interface{}{
			"code": 0,
			"data": map[string]interface{}{},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL, WithSecret("test-secret"))
	_, _ = client.GetContext()
}
