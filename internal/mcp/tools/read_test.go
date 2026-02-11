package tools

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadTool_Tool(t *testing.T) {
	tool := ReadToolDef()

	if tool.Name != "Read" {
		t.Errorf("expected tool name 'Read', got '%s'", tool.Name)
	}

	if tool.Description == "" {
		t.Error("expected non-empty description")
	}

	if !strings.Contains(tool.Description, "Reads a file from the local filesystem") {
		t.Error("expected description to contain 'Reads a file from the local filesystem'")
	}

	requiredParams := []string{"file_path"}
	for _, param := range requiredParams {
		if _, ok := tool.InputSchema.Properties[param]; !ok {
			t.Errorf("expected required parameter '%s' in schema", param)
		}
	}

	optionalParams := []string{"offset", "limit"}
	for _, param := range optionalParams {
		if _, ok := tool.InputSchema.Properties[param]; !ok {
			t.Errorf("expected optional parameter '%s' in schema", param)
		}
	}
}

func TestReadTool_Handler_SimpleRead(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "read_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "test.txt")
	content := "line1\nline2\nline3"
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	handler := ReadHandler(tmpDir)

	request := mockCallToolRequest(map[string]interface{}{
		"file_path": testFile,
	})

	result, err := handler(context.Background(), request)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.IsError {
		t.Errorf("expected success, got error: %v", getTextContent(result))
	}

	output := getTextContent(result)
	if !strings.Contains(output, "line1") {
		t.Errorf("expected output to contain 'line1', got: %s", output)
	}
	if !strings.Contains(output, "line2") {
		t.Errorf("expected output to contain 'line2', got: %s", output)
	}
	if !strings.Contains(output, "line3") {
		t.Errorf("expected output to contain 'line3', got: %s", output)
	}
}

func TestReadTool_Handler_LineNumbers(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "read_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "test.txt")
	content := "first\nsecond\nthird"
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	handler := ReadHandler(tmpDir)

	request := mockCallToolRequest(map[string]interface{}{
		"file_path": testFile,
	})

	result, err := handler(context.Background(), request)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := getTextContent(result)
	if !strings.Contains(output, "1\t") {
		t.Errorf("expected output to contain line number '1', got: %s", output)
	}
	if !strings.Contains(output, "2\t") {
		t.Errorf("expected output to contain line number '2', got: %s", output)
	}
	if !strings.Contains(output, "3\t") {
		t.Errorf("expected output to contain line number '3', got: %s", output)
	}
}

func TestReadTool_Handler_Offset(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "read_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "test.txt")
	content := "line1\nline2\nline3\nline4\nline5"
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	handler := ReadHandler(tmpDir)

	request := mockCallToolRequest(map[string]interface{}{
		"file_path": testFile,
		"offset":    float64(3),
	})

	result, err := handler(context.Background(), request)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := getTextContent(result)
	if strings.Contains(output, "line1") {
		t.Errorf("expected output NOT to contain 'line1', got: %s", output)
	}
	if strings.Contains(output, "line2") {
		t.Errorf("expected output NOT to contain 'line2', got: %s", output)
	}
	if !strings.Contains(output, "line3") {
		t.Errorf("expected output to contain 'line3', got: %s", output)
	}
}

func TestReadTool_Handler_Limit(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "read_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "test.txt")
	content := "line1\nline2\nline3\nline4\nline5"
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	handler := ReadHandler(tmpDir)

	request := mockCallToolRequest(map[string]interface{}{
		"file_path": testFile,
		"limit":     float64(2),
	})

	result, err := handler(context.Background(), request)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := getTextContent(result)
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != 2 {
		t.Errorf("expected 2 lines, got %d: %s", len(lines), output)
	}
}

func TestReadTool_Handler_OffsetAndLimit(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "read_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "test.txt")
	content := "line1\nline2\nline3\nline4\nline5"
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	handler := ReadHandler(tmpDir)

	request := mockCallToolRequest(map[string]interface{}{
		"file_path": testFile,
		"offset":    float64(2),
		"limit":     float64(2),
	})

	result, err := handler(context.Background(), request)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := getTextContent(result)
	if strings.Contains(output, "line1") {
		t.Errorf("expected output NOT to contain 'line1', got: %s", output)
	}
	if !strings.Contains(output, "line2") {
		t.Errorf("expected output to contain 'line2', got: %s", output)
	}
	if !strings.Contains(output, "line3") {
		t.Errorf("expected output to contain 'line3', got: %s", output)
	}
	if strings.Contains(output, "line4") {
		t.Errorf("expected output NOT to contain 'line4', got: %s", output)
	}
}

func TestReadTool_Handler_LineTruncation(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "read_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "test.txt")
	longLine := strings.Repeat("a", 2500)
	if err := os.WriteFile(testFile, []byte(longLine), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	handler := ReadHandler(tmpDir)

	request := mockCallToolRequest(map[string]interface{}{
		"file_path": testFile,
	})

	result, err := handler(context.Background(), request)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := getTextContent(result)
	if !strings.Contains(output, "...") {
		t.Errorf("expected truncation indicator '...', got: %s", output[:100])
	}
}

func TestReadTool_Handler_EmptyFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "read_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "empty.txt")
	if err := os.WriteFile(testFile, []byte(""), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	handler := ReadHandler(tmpDir)

	request := mockCallToolRequest(map[string]interface{}{
		"file_path": testFile,
	})

	result, err := handler(context.Background(), request)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := getTextContent(result)
	if !strings.Contains(output, "empty contents") {
		t.Errorf("expected empty file warning, got: %s", output)
	}
}

func TestReadTool_Handler_FileNotFound(t *testing.T) {
	handler := ReadHandler(os.TempDir())

	request := mockCallToolRequest(map[string]interface{}{
		"file_path": "/nonexistent/file.txt",
	})

	result, err := handler(context.Background(), request)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.IsError {
		t.Error("expected error for non-existent file")
	}
}

func TestReadTool_Handler_MissingFilePath(t *testing.T) {
	handler := ReadHandler(os.TempDir())

	request := mockCallToolRequest(map[string]interface{}{})

	result, err := handler(context.Background(), request)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.IsError {
		t.Error("expected error for missing file_path")
	}
}

func TestReadTool_Handler_DefaultLimit(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "read_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "test.txt")
	var lines []string
	for i := 1; i <= 2500; i++ {
		lines = append(lines, "line")
	}
	content := strings.Join(lines, "\n")
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	handler := ReadHandler(tmpDir)

	request := mockCallToolRequest(map[string]interface{}{
		"file_path": testFile,
	})

	result, err := handler(context.Background(), request)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := getTextContent(result)
	outputLines := strings.Split(strings.TrimSpace(output), "\n")
	if len(outputLines) != 2000 {
		t.Errorf("expected 2000 lines (default limit), got %d", len(outputLines))
	}
}
