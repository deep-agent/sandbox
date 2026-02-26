package tools

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestBashTool_Tool(t *testing.T) {
	tool := BashToolDef()

	if tool.Name != "Bash" {
		t.Errorf("expected tool name 'Bash', got '%s'", tool.Name)
	}

	if tool.Description == "" {
		t.Error("expected non-empty description")
	}

	if tool.InputSchema.Properties == nil {
		t.Error("expected input schema properties")
	}

	requiredParams := []string{"command"}
	for _, param := range requiredParams {
		if _, ok := tool.InputSchema.Properties[param]; !ok {
			t.Errorf("expected required parameter '%s' in schema", param)
		}
	}

	optionalParams := []string{"timeout_ms", "description"}
	for _, param := range optionalParams {
		if _, ok := tool.InputSchema.Properties[param]; !ok {
			t.Errorf("expected optional parameter '%s' in schema", param)
		}
	}
}

func TestBashTool_Handler_SimpleCommand(t *testing.T) {
	handler := BashHandler(os.TempDir())

	request := mockCallToolRequest(map[string]interface{}{
		"command": "echo hello",
	})

	result, err := handler(context.Background(), request)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.IsError {
		t.Errorf("expected success, got error: %v", getTextContent(result))
	}

	output := getTextContent(result)
	if output != "hello\n" {
		t.Errorf("expected 'hello\\n', got '%s'", output)
	}
}

func TestBashTool_Handler_CommandWithExitCode(t *testing.T) {
	handler := BashHandler(os.TempDir())

	request := mockCallToolRequest(map[string]interface{}{
		"command": "exit 1",
	})

	result, err := handler(context.Background(), request)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.IsError {
		t.Error("expected error result for non-zero exit code")
	}
}

func TestBashTool_Handler_MissingCommand(t *testing.T) {
	handler := BashHandler(os.TempDir())

	request := mockCallToolRequest(map[string]interface{}{})

	result, err := handler(context.Background(), request)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !result.IsError {
		t.Error("expected error result for missing command")
	}
}

func TestBashTool_Handler_Timeout(t *testing.T) {
	handler := BashHandler(os.TempDir())

	request := mockCallToolRequest(map[string]interface{}{
		"command":    "sleep 5",
		"timeout_ms": float64(100),
	})

	start := time.Now()
	result, err := handler(context.Background(), request)
	duration := time.Since(start)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if duration > 2*time.Second {
		t.Errorf("expected timeout within 2s, took %v", duration)
	}

	if !result.IsError {
		t.Error("expected error result for timeout")
	}
}

func TestBashTool_Handler_TimeoutMax(t *testing.T) {
	handler := BashHandler(os.TempDir())

	request := mockCallToolRequest(map[string]interface{}{
		"command":    "echo test",
		"timeout_ms": float64(999999999),
	})

	_, err := handler(context.Background(), request)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBashTool_Handler_OutputTruncation(t *testing.T) {
	handler := BashHandler(os.TempDir())

	request := mockCallToolRequest(map[string]interface{}{
		"command": "yes | head -n 50000",
	})

	result, err := handler(context.Background(), request)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := getTextContent(result)
	if len(output) > 35000 {
		t.Errorf("expected output to be truncated, got length %d", len(output))
	}
}

func TestBashTool_Handler_WorkingDirectory(t *testing.T) {
	tmpDir := os.TempDir()
	handler := BashHandler(tmpDir)

	request := mockCallToolRequest(map[string]interface{}{
		"command": "pwd",
	})

	result, err := handler(context.Background(), request)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := getTextContent(result)
	if output == "" {
		t.Error("expected non-empty output for pwd command")
	}
}

func TestBashTool_Handler_EnvironmentVariables(t *testing.T) {
	handler := BashHandler(os.TempDir())

	request := mockCallToolRequest(map[string]interface{}{
		"command": "echo $HOME",
	})

	result, err := handler(context.Background(), request)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.IsError {
		t.Errorf("expected success, got error: %v", getTextContent(result))
	}
}

func TestBashTool_Handler_PipedCommand(t *testing.T) {
	handler := BashHandler(os.TempDir())

	request := mockCallToolRequest(map[string]interface{}{
		"command": "echo 'hello world' | wc -w",
	})

	result, err := handler(context.Background(), request)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.IsError {
		t.Errorf("expected success, got error: %v", getTextContent(result))
	}
}
