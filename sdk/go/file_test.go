package sandbox

import (
	"testing"
	"time"

	"github.com/deep-agent/sandbox/model"
)

func TestFileWriteAndRead(t *testing.T) {
	client := newTestClient()
	testFile := "/home/sandbox/workspace/sdk_test_file.txt"
	testContent := "hello from sdk test"

	err := client.FileWrite(&model.FileWriteRequest{
		File:    testFile,
		Content: testContent,
	})
	if err != nil {
		t.Fatalf("FileWrite error: %v", err)
	}

	result, err := client.FileRead(&model.FileReadRequest{
		File: testFile,
	})
	if err != nil {
		t.Fatalf("FileRead error: %v", err)
	}
	if result.Content != testContent {
		t.Errorf("expected content %q, got %q", testContent, result.Content)
	}
	t.Logf("FileRead content: %s", result.Content)

	err = client.FileDelete(&model.FileDeleteRequest{
		Path: testFile,
	})
	if err != nil {
		t.Fatalf("FileDelete error: %v", err)
	}
}

func TestFileList(t *testing.T) {
	client := newTestClient()

	result, err := client.FileList(&model.FileListRequest{
		Path: "/home/sandbox/workspace",
	})
	if err != nil {
		t.Fatalf("FileList error: %v", err)
	}
	if result == nil {
		t.Fatal("expected result to be non-nil")
	}
	t.Logf("FileList found %d files", len(result.Files))
}

func TestMkDirAndDelete(t *testing.T) {
	client := newTestClient()
	testDir := "/home/sandbox/workspace/sdk_test_dir_" + time.Now().Format("20060102150405")

	err := client.MkDir(&model.MkDirRequest{
		Path: testDir,
	})
	if err != nil {
		t.Fatalf("MkDir error: %v", err)
	}

	result, err := client.FileExists(testDir)
	if err != nil {
		t.Fatalf("FileExists error: %v", err)
	}
	if !result.Exists {
		t.Error("expected directory to exist")
	}

	err = client.FileDelete(&model.FileDeleteRequest{
		Path: testDir,
	})
	if err != nil {
		t.Fatalf("FileDelete error: %v", err)
	}
}

func TestFileCopy(t *testing.T) {
	client := newTestClient()
	srcFile := "/home/sandbox/workspace/sdk_copy_src.txt"
	dstFile := "/home/sandbox/workspace/sdk_copy_dst.txt"
	testContent := "copy test content"

	err := client.FileWrite(&model.FileWriteRequest{
		File:    srcFile,
		Content: testContent,
	})
	if err != nil {
		t.Fatalf("FileWrite error: %v", err)
	}

	err = client.FileCopy(&model.FileCopyRequest{
		Source:      srcFile,
		Destination: dstFile,
	})
	if err != nil {
		t.Fatalf("FileCopy error: %v", err)
	}

	result, err := client.FileRead(&model.FileReadRequest{
		File: dstFile,
	})
	if err != nil {
		t.Fatalf("FileRead error: %v", err)
	}
	if result.Content != testContent {
		t.Errorf("expected content %q, got %q", testContent, result.Content)
	}

	_ = client.FileDelete(&model.FileDeleteRequest{Path: srcFile})
	_ = client.FileDelete(&model.FileDeleteRequest{Path: dstFile})
}

func TestFileMove(t *testing.T) {
	client := newTestClient()
	srcFile := "/home/sandbox/workspace/sdk_move_src.txt"
	dstFile := "/home/sandbox/workspace/sdk_move_dst.txt"
	testContent := "move test content"

	err := client.FileWrite(&model.FileWriteRequest{
		File:    srcFile,
		Content: testContent,
	})
	if err != nil {
		t.Fatalf("FileWrite error: %v", err)
	}

	err = client.FileMove(&model.FileMoveRequest{
		Source:      srcFile,
		Destination: dstFile,
	})
	if err != nil {
		t.Fatalf("FileMove error: %v", err)
	}

	srcExists, _ := client.FileExists(srcFile)
	if srcExists != nil && srcExists.Exists {
		t.Error("expected source file to not exist after move")
	}

	result, err := client.FileRead(&model.FileReadRequest{
		File: dstFile,
	})
	if err != nil {
		t.Fatalf("FileRead error: %v", err)
	}
	if result.Content != testContent {
		t.Errorf("expected content %q, got %q", testContent, result.Content)
	}

	_ = client.FileDelete(&model.FileDeleteRequest{Path: dstFile})
}
