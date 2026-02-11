package tools

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestGlobTool_Tool(t *testing.T) {
	tool := GlobToolDef()

	if tool.Name != "Glob" {
		t.Errorf("expected tool name 'Glob', got '%s'", tool.Name)
	}

	if tool.Description == "" {
		t.Error("expected non-empty description")
	}

	if !strings.Contains(tool.Description, "Fast file pattern matching") {
		t.Error("expected description to contain 'Fast file pattern matching'")
	}

	if _, ok := tool.InputSchema.Properties["pattern"]; !ok {
		t.Error("expected 'pattern' parameter in schema")
	}

	if _, ok := tool.InputSchema.Properties["path"]; !ok {
		t.Error("expected 'path' parameter in schema")
	}
}

func TestGlobTool_Handler_SimplePattern(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "glob_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	handler := GlobHandler(tmpDir)

	request := mockCallToolRequest(map[string]interface{}{
		"pattern": "*.txt",
	})

	result, err := handler(context.Background(), request)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.IsError {
		t.Errorf("expected success, got error: %v", getTextContent(result))
	}

	output := getTextContent(result)
	if !strings.Contains(output, "test.txt") {
		t.Errorf("expected output to contain 'test.txt', got: %s", output)
	}
}

func TestGlobTool_Handler_DoubleStarPattern(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "glob_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	subDir := filepath.Join(tmpDir, "src", "components")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("failed to create sub dir: %v", err)
	}

	testFile := filepath.Join(subDir, "Button.tsx")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	handler := GlobHandler(tmpDir)

	request := mockCallToolRequest(map[string]interface{}{
		"pattern": "**/*.tsx",
	})

	result, err := handler(context.Background(), request)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.IsError {
		t.Errorf("expected success, got error: %v", getTextContent(result))
	}

	output := getTextContent(result)
	if !strings.Contains(output, "Button.tsx") {
		t.Errorf("expected output to contain 'Button.tsx', got: %s", output)
	}
}

func TestGlobTool_Handler_CustomPath(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "glob_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "custom.js")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	handler := GlobHandler("/some/other/path")

	request := mockCallToolRequest(map[string]interface{}{
		"pattern": "*.js",
		"path":    tmpDir,
	})

	result, err := handler(context.Background(), request)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.IsError {
		t.Errorf("expected success, got error: %v", getTextContent(result))
	}

	output := getTextContent(result)
	if !strings.Contains(output, "custom.js") {
		t.Errorf("expected output to contain 'custom.js', got: %s", output)
	}
}

func TestGlobTool_Handler_NoMatches(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "glob_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	handler := GlobHandler(tmpDir)

	request := mockCallToolRequest(map[string]interface{}{
		"pattern": "*.nonexistent",
	})

	result, err := handler(context.Background(), request)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.IsError {
		t.Error("expected success for no matches")
	}

	output := getTextContent(result)
	if !strings.Contains(output, "No files found") {
		t.Errorf("expected 'No files found' message, got: %s", output)
	}
}

func TestGlobTool_Handler_MissingPattern(t *testing.T) {
	handler := GlobHandler(os.TempDir())

	request := mockCallToolRequest(map[string]interface{}{})

	result, err := handler(context.Background(), request)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.IsError {
		t.Error("expected error for missing pattern")
	}
}

func TestGlobTool_Handler_SortByModTime(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "glob_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	oldFile := filepath.Join(tmpDir, "old.txt")
	if err := os.WriteFile(oldFile, []byte("old"), 0644); err != nil {
		t.Fatalf("failed to create old file: %v", err)
	}
	oldTime := time.Now().Add(-1 * time.Hour)
	os.Chtimes(oldFile, oldTime, oldTime)

	time.Sleep(10 * time.Millisecond)

	newFile := filepath.Join(tmpDir, "new.txt")
	if err := os.WriteFile(newFile, []byte("new"), 0644); err != nil {
		t.Fatalf("failed to create new file: %v", err)
	}

	handler := GlobHandler(tmpDir)

	request := mockCallToolRequest(map[string]interface{}{
		"pattern": "*.txt",
	})

	result, err := handler(context.Background(), request)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := getTextContent(result)
	lines := strings.Split(output, "\n")

	if len(lines) < 2 {
		t.Fatalf("expected at least 2 files, got: %s", output)
	}

	if !strings.Contains(lines[0], "new.txt") {
		t.Errorf("expected newest file first, got: %s", lines[0])
	}
}

func TestGlobTool_Handler_PrefixPattern(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "glob_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	srcDir := filepath.Join(tmpDir, "src")
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		t.Fatalf("failed to create src dir: %v", err)
	}

	testFile := filepath.Join(srcDir, "app.ts")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	otherFile := filepath.Join(tmpDir, "other.ts")
	if err := os.WriteFile(otherFile, []byte("test"), 0644); err != nil {
		t.Fatalf("failed to create other file: %v", err)
	}

	handler := GlobHandler(tmpDir)

	request := mockCallToolRequest(map[string]interface{}{
		"pattern": "src/**/*.ts",
	})

	result, err := handler(context.Background(), request)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := getTextContent(result)
	if !strings.Contains(output, "app.ts") {
		t.Errorf("expected output to contain 'app.ts', got: %s", output)
	}
	if strings.Contains(output, "other.ts") {
		t.Errorf("expected output NOT to contain 'other.ts', got: %s", output)
	}
}
