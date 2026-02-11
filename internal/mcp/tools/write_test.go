package tools

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteTool_Tool(t *testing.T) {
	tool := WriteToolDef()

	if tool.Name != "Write" {
		t.Errorf("expected tool name 'Write', got '%s'", tool.Name)
	}

	if tool.Description == "" {
		t.Error("expected non-empty description")
	}

	if !strings.Contains(tool.Description, "Writes a file to the local filesystem") {
		t.Error("expected description to contain 'Writes a file to the local filesystem'")
	}

	requiredParams := []string{"file_path", "content"}
	for _, param := range requiredParams {
		if _, ok := tool.InputSchema.Properties[param]; !ok {
			t.Errorf("expected required parameter '%s' in schema", param)
		}
	}
}

func TestWriteTool_Handler_SimpleWrite(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "write_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	handler := WriteHandler(tmpDir)

	testFile := filepath.Join(tmpDir, "test.txt")
	request := mockCallToolRequest(map[string]interface{}{
		"file_path": testFile,
		"content":   "hello world",
	})

	result, err := handler(context.Background(), request)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.IsError {
		t.Errorf("expected success, got error: %v", getTextContent(result))
	}

	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("failed to read written file: %v", err)
	}

	if string(content) != "hello world" {
		t.Errorf("expected 'hello world', got '%s'", string(content))
	}
}

func TestWriteTool_Handler_OverwriteExisting(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "write_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "existing.txt")
	if err := os.WriteFile(testFile, []byte("original content"), 0644); err != nil {
		t.Fatalf("failed to create existing file: %v", err)
	}

	handler := WriteHandler(tmpDir)

	request := mockCallToolRequest(map[string]interface{}{
		"file_path": testFile,
		"content":   "new content",
	})

	result, err := handler(context.Background(), request)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.IsError {
		t.Errorf("expected success, got error: %v", getTextContent(result))
	}

	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("failed to read written file: %v", err)
	}

	if string(content) != "new content" {
		t.Errorf("expected 'new content', got '%s'", string(content))
	}
}

func TestWriteTool_Handler_CreateDirectory(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "write_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	handler := WriteHandler(tmpDir)

	testFile := filepath.Join(tmpDir, "subdir", "nested", "test.txt")
	request := mockCallToolRequest(map[string]interface{}{
		"file_path": testFile,
		"content":   "nested content",
	})

	result, err := handler(context.Background(), request)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.IsError {
		t.Errorf("expected success, got error: %v", getTextContent(result))
	}

	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("failed to read written file: %v", err)
	}

	if string(content) != "nested content" {
		t.Errorf("expected 'nested content', got '%s'", string(content))
	}
}

func TestWriteTool_Handler_EmptyContent(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "write_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	handler := WriteHandler(tmpDir)

	testFile := filepath.Join(tmpDir, "empty.txt")
	request := mockCallToolRequest(map[string]interface{}{
		"file_path": testFile,
		"content":   "",
	})

	result, err := handler(context.Background(), request)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.IsError {
		t.Errorf("expected success, got error: %v", getTextContent(result))
	}

	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("failed to read written file: %v", err)
	}

	if string(content) != "" {
		t.Errorf("expected empty content, got '%s'", string(content))
	}
}

func TestWriteTool_Handler_MissingFilePath(t *testing.T) {
	handler := WriteHandler(os.TempDir())

	request := mockCallToolRequest(map[string]interface{}{
		"content": "some content",
	})

	result, err := handler(context.Background(), request)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.IsError {
		t.Error("expected error for missing file_path")
	}
}

func TestWriteTool_Handler_MissingContent(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "write_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	handler := WriteHandler(tmpDir)

	testFile := filepath.Join(tmpDir, "test.txt")
	request := mockCallToolRequest(map[string]interface{}{
		"file_path": testFile,
	})

	result, err := handler(context.Background(), request)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.IsError {
		t.Error("expected error for missing content")
	}
}

func TestWriteTool_Handler_MultilineContent(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "write_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	handler := WriteHandler(tmpDir)

	testFile := filepath.Join(tmpDir, "multiline.txt")
	multilineContent := "line1\nline2\nline3\n"
	request := mockCallToolRequest(map[string]interface{}{
		"file_path": testFile,
		"content":   multilineContent,
	})

	result, err := handler(context.Background(), request)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.IsError {
		t.Errorf("expected success, got error: %v", getTextContent(result))
	}

	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("failed to read written file: %v", err)
	}

	if string(content) != multilineContent {
		t.Errorf("expected multiline content, got '%s'", string(content))
	}
}

func TestWriteTool_Handler_SuccessMessage(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "write_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	handler := WriteHandler(tmpDir)

	testFile := filepath.Join(tmpDir, "success.txt")
	request := mockCallToolRequest(map[string]interface{}{
		"file_path": testFile,
		"content":   "test",
	})

	result, err := handler(context.Background(), request)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := getTextContent(result)
	if !strings.Contains(output, "File written successfully") {
		t.Errorf("expected success message, got: %s", output)
	}
	if !strings.Contains(output, testFile) {
		t.Errorf("expected file path in message, got: %s", output)
	}
}

func TestWriteTool_Handler_UnicodeContent(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "write_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	handler := WriteHandler(tmpDir)

	testFile := filepath.Join(tmpDir, "unicode.txt")
	unicodeContent := "Hello 世界 🌍 مرحبا"
	request := mockCallToolRequest(map[string]interface{}{
		"file_path": testFile,
		"content":   unicodeContent,
	})

	result, err := handler(context.Background(), request)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.IsError {
		t.Errorf("expected success, got error: %v", getTextContent(result))
	}

	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("failed to read written file: %v", err)
	}

	if string(content) != unicodeContent {
		t.Errorf("expected unicode content, got '%s'", string(content))
	}
}
