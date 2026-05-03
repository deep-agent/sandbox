package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/deep-agent/sandbox/types/model"
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
				"home_dir": "/home/sandbox",
				"os":       "linux",
				"arch":     "amd64",
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
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
	if ctx.OS != "linux" {
		t.Errorf("expected OS linux, got %s", ctx.OS)
	}
	if ctx.Arch != "amd64" {
		t.Errorf("expected Arch amd64, got %s", ctx.Arch)
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
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("decode request: %v", err)
		}

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
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	result, err := client.BashExec(&model.BashExecRequest{
		Command: "echo hello",
		Cwd:     "/tmp",
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
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("decode request: %v", err)
		}

		if req.File != "/tmp/test.txt" {
			t.Errorf("expected file /tmp/test.txt, got %s", req.File)
		}

		resp := map[string]interface{}{
			"code": 0,
			"data": map[string]interface{}{
				"content": "file content",
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
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
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("decode request: %v", err)
		}

		if req.File != "/tmp/test.txt" {
			t.Errorf("expected file /tmp/test.txt, got %s", req.File)
		}
		if req.Content != "new content" {
			t.Errorf("expected content 'new content', got %s", req.Content)
		}

		resp := map[string]interface{}{
			"code": 0,
		}
		_ = json.NewEncoder(w).Encode(resp)
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
		_ = json.NewEncoder(w).Encode(resp)
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
		_ = json.NewEncoder(w).Encode(resp)
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
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("decode request: %v", err)
		}

		if req.Source != "/tmp/old.txt" || req.Destination != "/tmp/new.txt" {
			t.Errorf("unexpected source/destination")
		}

		resp := map[string]interface{}{"code": 0}
		_ = json.NewEncoder(w).Encode(resp)
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
		_ = json.NewEncoder(w).Encode(resp)
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
		_ = json.NewEncoder(w).Encode(resp)
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
		_ = json.NewEncoder(w).Encode(resp)
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

func TestFileWriteAtomicAndMode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/file/write" {
			t.Errorf("expected path /v1/file/write, got %s", r.URL.Path)
		}
		var req model.FileWriteRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("decode request: %v", err)
		}
		if !req.Atomic || req.Mode != 0600 {
			t.Errorf("expected atomic=true mode=0600, got atomic=%v mode=%o", req.Atomic, req.Mode)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"code": 0})
	}))
	defer server.Close()

	client := NewClient(server.URL)
	if err := client.FileWrite(&model.FileWriteRequest{File: "/tmp/test.txt", Content: "new content", Atomic: true, Mode: 0600}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestFileCreateTemp(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/file/create-temp" {
			t.Errorf("expected path /v1/file/create-temp, got %s", r.URL.Path)
		}
		var req model.FileCreateTempRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("decode request: %v", err)
		}
		if req.Pattern != "shell-*.sh" || req.Mode != 0700 {
			t.Errorf("unexpected request: %+v", req)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"code": 0, "data": map[string]interface{}{"file": "/tmp/shell-123.sh"}})
	}))
	defer server.Close()

	client := NewClient(server.URL)
	result, err := client.FileCreateTemp(&model.FileCreateTempRequest{Dir: "/tmp", Pattern: "shell-*.sh", Content: "echo hi", Mode: 0700})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.File != "/tmp/shell-123.sh" {
		t.Fatalf("unexpected temp file result: %+v", result)
	}
}

func TestFileGlob(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/file/glob" {
			t.Errorf("expected path /v1/file/glob, got %s", r.URL.Path)
		}
		var req model.FileGlobRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("decode request: %v", err)
		}
		if req.Pattern != "*.go" {
			t.Errorf("expected pattern *.go, got %s", req.Pattern)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"code": 0, "data": map[string]interface{}{"files": []string{"/tmp/a.go"}, "count": 1, "truncated": false, "output": "/tmp/a.go"}})
	}))
	defer server.Close()

	client := NewClient(server.URL)
	result, err := client.FileGlob(&model.FileGlobRequest{Path: "/tmp", Pattern: "*.go"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Count != 1 || len(result.Files) != 1 {
		t.Fatalf("unexpected glob result: %+v", result)
	}
}

func TestFileEvalSymlinks(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/file/eval-symlinks" {
			t.Errorf("expected path /v1/file/eval-symlinks, got %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"code": 0, "data": map[string]interface{}{"resolved_path": "/tmp/real.txt"}})
	}))
	defer server.Close()

	client := NewClient(server.URL)
	result, err := client.FileEvalSymlinks(&model.FileEvalSymlinksRequest{Path: "/tmp/link.txt"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ResolvedPath != "/tmp/real.txt" {
		t.Fatalf("unexpected symlink result: %+v", result)
	}
}

func TestFileAppend(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/file/append" {
			t.Errorf("expected path /v1/file/append, got %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"code": 0})
	}))
	defer server.Close()

	client := NewClient(server.URL)
	if err := client.FileAppend(&model.FileAppendRequest{File: "/tmp/test.txt", Content: "abc"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestFileStat(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/file/stat" {
			t.Errorf("expected path /v1/file/stat, got %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"code": 0, "data": map[string]interface{}{"exists": true, "is_dir": false, "size": 3}})
	}))
	defer server.Close()

	client := NewClient(server.URL)
	result, err := client.FileStat(&model.FileStatRequest{Path: "/tmp/test.txt"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Exists || result.Size != 3 {
		t.Fatalf("unexpected stat result: %+v", result)
	}
}

func TestGrepSearch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/grep/search" {
			t.Errorf("expected path /v1/grep/search, got %s", r.URL.Path)
		}

		var req model.GrepRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("decode request: %v", err)
		}

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
		_ = json.NewEncoder(w).Encode(resp)
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

func TestAPIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"code":    1,
			"message": "something went wrong",
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	_, err := client.GetContext()
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err.Error() != "API error (code 1): something went wrong" {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestAPIErrorNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"code":    404,
			"message": "file not found: /tmp/noexist.txt",
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	_, err := client.FileRead(&model.FileReadRequest{File: "/tmp/noexist.txt"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !errors.Is(err, os.ErrNotExist) {
		t.Errorf("expected os.ErrNotExist, got: %v", err)
	}

	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected APIError, got: %T", err)
	}
	if apiErr.Code != 404 {
		t.Errorf("expected code 404, got %d", apiErr.Code)
	}
}

func TestAPIErrorNonNotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"code":    500,
			"message": "internal error",
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL)
	_, err := client.FileRead(&model.FileReadRequest{File: "/tmp/test.txt"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if errors.Is(err, os.ErrNotExist) {
		t.Error("should not be os.ErrNotExist for code 500")
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
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewClient(server.URL, WithSecret("test-secret"))
	if _, err := client.GetContext(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
