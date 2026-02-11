package sandbox

import (
	"testing"

	"github.com/deep-agent/sandbox/model"
)

func TestGrepSearch(t *testing.T) {
	client := newTestClient()

	testFile := "/home/sandbox/workspace/grep_test_file.txt"
	testContent := "line1 hello\nline2 world\nline3 hello world"

	err := client.FileWrite(&model.FileWriteRequest{
		File:    testFile,
		Content: testContent,
	})
	if err != nil {
		t.Fatalf("FileWrite error: %v", err)
	}
	defer func() {
		_ = client.FileDelete(&model.FileDeleteRequest{Path: testFile})
	}()

	result, err := client.GrepSearch(&model.GrepRequest{
		Pattern: "hello",
		Path:    "/home/sandbox/workspace",
	})
	if err != nil {
		t.Fatalf("GrepSearch error: %v", err)
	}
	if result == nil {
		t.Fatal("expected result to be non-nil")
	}
	if result.Output == "" {
		t.Error("expected non-empty output")
	}
	t.Logf("GrepSearch output: %s", result.Output)
}
