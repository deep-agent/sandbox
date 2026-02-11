package tools

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGrepTool_Tool(t *testing.T) {
	tool := GrepToolDef()

	if tool.Name != "Grep" {
		t.Errorf("expected tool name 'Grep', got '%s'", tool.Name)
	}

	if tool.Description == "" {
		t.Error("expected non-empty description")
	}

	if !strings.Contains(tool.Description, "ripgrep") {
		t.Error("expected description to contain 'ripgrep'")
	}

	requiredParams := []string{"pattern"}
	for _, param := range requiredParams {
		if _, ok := tool.InputSchema.Properties[param]; !ok {
			t.Errorf("expected required parameter '%s' in schema", param)
		}
	}
}

func TestGrepTool_Handler_SimpleSearch(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "grep_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("hello world\nfoo bar\nhello again"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	handler := GrepHandler(tmpDir)

	request := mockCallToolRequest(map[string]interface{}{
		"pattern": "hello",
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

func TestGrepTool_Handler_ContentMode(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "grep_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("line1\nhello world\nline3"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	handler := GrepHandler(tmpDir)

	request := mockCallToolRequest(map[string]interface{}{
		"pattern":     "hello",
		"output_mode": "content",
	})

	result, err := handler(context.Background(), request)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := getTextContent(result)
	if !strings.Contains(output, "hello world") {
		t.Errorf("expected output to contain 'hello world', got: %s", output)
	}
	if !strings.Contains(output, "2:") {
		t.Errorf("expected output to contain line number '2:', got: %s", output)
	}
}

func TestGrepTool_Handler_CountMode(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "grep_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("hello\nhello\nhello"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	handler := GrepHandler(tmpDir)

	request := mockCallToolRequest(map[string]interface{}{
		"pattern":     "hello",
		"output_mode": "count",
	})

	result, err := handler(context.Background(), request)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := getTextContent(result)
	if !strings.Contains(output, "3") {
		t.Errorf("expected output to contain '3', got: %s", output)
	}
}

func TestGrepTool_Handler_CaseInsensitive(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "grep_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("HELLO world"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	handler := GrepHandler(tmpDir)

	request := mockCallToolRequest(map[string]interface{}{
		"pattern":          "hello",
		"case_insensitive": true,
	})

	result, err := handler(context.Background(), request)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := getTextContent(result)
	if !strings.Contains(output, "test.txt") {
		t.Errorf("expected output to contain 'test.txt', got: %s", output)
	}
}

func TestGrepTool_Handler_GlobFilter(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "grep_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	jsFile := filepath.Join(tmpDir, "app.js")
	if err := os.WriteFile(jsFile, []byte("function hello() {}"), 0644); err != nil {
		t.Fatalf("failed to create js file: %v", err)
	}

	txtFile := filepath.Join(tmpDir, "readme.txt")
	if err := os.WriteFile(txtFile, []byte("hello readme"), 0644); err != nil {
		t.Fatalf("failed to create txt file: %v", err)
	}

	handler := GrepHandler(tmpDir)

	request := mockCallToolRequest(map[string]interface{}{
		"pattern": "hello",
		"glob":    "*.js",
	})

	result, err := handler(context.Background(), request)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := getTextContent(result)
	if !strings.Contains(output, "app.js") {
		t.Errorf("expected output to contain 'app.js', got: %s", output)
	}
	if strings.Contains(output, "readme.txt") {
		t.Errorf("expected output NOT to contain 'readme.txt', got: %s", output)
	}
}

func TestGrepTool_Handler_ContextLines(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "grep_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "test.txt")
	content := "line1\nline2\ntarget\nline4\nline5"
	if err := os.WriteFile(testFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	handler := GrepHandler(tmpDir)

	request := mockCallToolRequest(map[string]interface{}{
		"pattern":       "target",
		"output_mode":   "content",
		"context_lines": float64(1),
	})

	result, err := handler(context.Background(), request)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := getTextContent(result)
	if !strings.Contains(output, "line2") {
		t.Errorf("expected output to contain context line 'line2', got: %s", output)
	}
	if !strings.Contains(output, "line4") {
		t.Errorf("expected output to contain context line 'line4', got: %s", output)
	}
}

func TestGrepTool_Handler_NoMatches(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "grep_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("hello world"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	handler := GrepHandler(tmpDir)

	request := mockCallToolRequest(map[string]interface{}{
		"pattern": "nonexistent",
	})

	result, err := handler(context.Background(), request)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := getTextContent(result)
	if !strings.Contains(output, "No matches found") {
		t.Errorf("expected 'No matches found' message, got: %s", output)
	}
}

func TestGrepTool_Handler_MissingPattern(t *testing.T) {
	handler := GrepHandler(os.TempDir())

	request := mockCallToolRequest(map[string]interface{}{})

	result, err := handler(context.Background(), request)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.IsError {
		t.Error("expected error for missing pattern")
	}
}

func TestGrepTool_Handler_CustomPath(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "grep_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "custom.txt")
	if err := os.WriteFile(testFile, []byte("custom content"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	handler := GrepHandler("/some/other/path")

	request := mockCallToolRequest(map[string]interface{}{
		"pattern": "custom",
		"path":    tmpDir,
	})

	result, err := handler(context.Background(), request)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := getTextContent(result)
	if !strings.Contains(output, "custom.txt") {
		t.Errorf("expected output to contain 'custom.txt', got: %s", output)
	}
}

func TestGrepTool_Handler_RegexPattern(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "grep_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("function test() {}\nfunction hello() {}"), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	handler := GrepHandler(tmpDir)

	request := mockCallToolRequest(map[string]interface{}{
		"pattern":     "function .+",
		"output_mode": "content",
	})

	result, err := handler(context.Background(), request)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := getTextContent(result)
	if !strings.Contains(output, "function test") {
		t.Errorf("expected output to contain 'function test', got: %s", output)
	}
	if !strings.Contains(output, "function hello") {
		t.Errorf("expected output to contain 'function hello', got: %s", output)
	}
}
