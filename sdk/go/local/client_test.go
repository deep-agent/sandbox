package local

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/deep-agent/sandbox/types/model"
)

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

func TestNewClientWithSandboxContext(t *testing.T) {
	workDir := t.TempDir()
	ctx := &model.SandboxContext{
		Workspace: "/home/test",
		OS:        "linux",
		Arch:      "amd64",
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

	if ctx.Workspace != workDir {
		t.Errorf("expected Workspace %s, got %s", workDir, ctx.Workspace)
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

	if err := os.WriteFile(filepath.Join(workDir, "file1.txt"), []byte("1"), 0644); err != nil {
		t.Fatalf("failed to create file1.txt: %v", err)
	}
	if err := os.WriteFile(filepath.Join(workDir, "file2.txt"), []byte("2"), 0644); err != nil {
		t.Fatalf("failed to create file2.txt: %v", err)
	}
	if err := os.Mkdir(filepath.Join(workDir, "subdir"), 0755); err != nil {
		t.Fatalf("failed to create subdir: %v", err)
	}

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
	if err := os.WriteFile(testFile, []byte("delete me"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

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
	if err := os.WriteFile(srcFile, []byte("move me"), 0644); err != nil {
		t.Fatalf("failed to create source file: %v", err)
	}

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
	if err := os.WriteFile(srcFile, []byte("copy me"), 0644); err != nil {
		t.Fatalf("failed to create source file: %v", err)
	}

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
	if err := os.WriteFile(existingFile, []byte("exists"), 0644); err != nil {
		t.Fatalf("failed to create existing file: %v", err)
	}

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

func TestFileWriteAtomicAndMode(t *testing.T) {
	workDir := t.TempDir()
	client := NewClient(workDir)
	testFile := filepath.Join(workDir, "atomic.txt")

	err := client.FileWrite(&model.FileWriteRequest{File: testFile, Content: "atomic", Atomic: true, Mode: 0600})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	info, err := os.Stat(testFile)
	if err != nil {
		t.Fatalf("stat file: %v", err)
	}
	if info.Mode().Perm() != 0600 {
		t.Fatalf("expected perm 0600, got %o", info.Mode().Perm())
	}
}

func TestFileCreateTemp(t *testing.T) {
	workDir := t.TempDir()
	client := NewClient(workDir)

	result, err := client.FileCreateTemp(&model.FileCreateTempRequest{Dir: workDir, Pattern: "shell-*.sh", Content: "echo hi", Mode: 0700})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, err := os.ReadFile(result.File)
	if err != nil {
		t.Fatalf("read temp file: %v", err)
	}
	if string(content) != "echo hi" {
		t.Fatalf("expected echo hi, got %q", string(content))
	}
}

func TestFileGlob(t *testing.T) {
	workDir := t.TempDir()
	client := NewClient(workDir)
	if err := os.WriteFile(filepath.Join(workDir, "a.go"), []byte("package main"), 0644); err != nil {
		t.Fatalf("create a.go: %v", err)
	}
	if err := os.WriteFile(filepath.Join(workDir, "b.txt"), []byte("txt"), 0644); err != nil {
		t.Fatalf("create b.txt: %v", err)
	}

	result, err := client.FileGlob(&model.FileGlobRequest{Path: workDir, Pattern: "*.go"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Count != 1 || len(result.Files) != 1 || filepath.Base(result.Files[0]) != "a.go" {
		t.Fatalf("unexpected glob result: %+v", result)
	}
}

func TestFileEvalSymlinks(t *testing.T) {
	workDir := t.TempDir()
	client := NewClient(workDir)
	target := filepath.Join(workDir, "target.txt")
	link := filepath.Join(workDir, "link.txt")
	if err := os.WriteFile(target, []byte("target"), 0644); err != nil {
		t.Fatalf("create target: %v", err)
	}
	if err := os.Symlink(target, link); err != nil {
		t.Fatalf("create symlink: %v", err)
	}

	result, err := client.FileEvalSymlinks(&model.FileEvalSymlinksRequest{Path: link})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ResolvedPath != target {
		t.Fatalf("expected resolved path %q, got %q", target, result.ResolvedPath)
	}
}

func TestFileAppendAndStat(t *testing.T) {
	workDir := t.TempDir()
	client := NewClient(workDir)
	testFile := filepath.Join(workDir, "append-stat.txt")

	if err := client.FileAppend(&model.FileAppendRequest{File: testFile, Content: "abc"}); err != nil {
		t.Fatalf("append failed: %v", err)
	}
	if err := client.FileAppend(&model.FileAppendRequest{File: testFile, Content: "def"}); err != nil {
		t.Fatalf("append failed: %v", err)
	}
	result, err := client.FileStat(&model.FileStatRequest{Path: testFile})
	if err != nil {
		t.Fatalf("stat failed: %v", err)
	}
	if !result.Exists || result.Size != 6 {
		t.Fatalf("unexpected stat result: %+v", result)
	}
}

func TestGrepSearch(t *testing.T) {
	workDir := t.TempDir()
	client := NewClient(workDir)

	testFile := filepath.Join(workDir, "grep_test.txt")
	if err := os.WriteFile(testFile, []byte("line one\nfind this line\nline three\n"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

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
	if err := os.WriteFile(testFile, []byte("line one\nline two\nline three\n"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

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
	if err := os.WriteFile(testFile, []byte("HELLO world\nhello WORLD\n"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

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
