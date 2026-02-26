package bash

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestNewExecutor(t *testing.T) {
	executor := NewExecutor()

	if executor == nil {
		t.Fatal("NewExecutor returned nil")
	}

	if executor.timeout != 30*time.Second {
		t.Errorf("expected default timeout 30s, got %v", executor.timeout)
	}
}

func TestSetTimeout(t *testing.T) {
	executor := NewExecutor()

	newTimeout := 60 * time.Second
	executor.SetTimeout(newTimeout)

	if executor.timeout != newTimeout {
		t.Errorf("expected timeout %v, got %v", newTimeout, executor.timeout)
	}
}

func TestExecute_EchoHello(t *testing.T) {
	workDir := t.TempDir()
	executor := NewExecutor()

	ctx := context.Background()
	result, err := executor.Execute(ctx, "echo hello", workDir, nil)

	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if result.ExitCode != 0 {
		t.Errorf("expected exit code 0, got %d", result.ExitCode)
	}

	if strings.TrimSpace(result.Stdout) != "hello" {
		t.Errorf("expected stdout 'hello', got %q", result.Stdout)
	}

	if result.TimedOut {
		t.Error("unexpected timeout")
	}
}

func TestExecute_ExitCode(t *testing.T) {
	workDir := t.TempDir()
	executor := NewExecutor()

	ctx := context.Background()
	result, err := executor.Execute(ctx, "exit 42", workDir, nil)

	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if result.ExitCode != 42 {
		t.Errorf("expected exit code 42, got %d", result.ExitCode)
	}
}

func TestExecute_Timeout(t *testing.T) {
	workDir := t.TempDir()
	executor := NewExecutor()
	executor.SetTimeout(500 * time.Millisecond)

	ctx := context.Background()
	startTime := time.Now()
	result, err := executor.Execute(ctx, "sleep 10", workDir, nil)
	elapsed := time.Since(startTime)

	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if elapsed > 2*time.Second {
		t.Errorf("command should have timed out around 500ms, but took %v", elapsed)
	}

	if result.TimedOut {
		if !strings.Contains(result.Metadata, "timed out") {
			t.Errorf("expected metadata to contain 'timed out', got %q", result.Metadata)
		}
	} else {
		if result.ExitCode == 0 {
			t.Error("expected non-zero exit code or timeout")
		}
	}
}

func TestExecute_Stderr(t *testing.T) {
	workDir := t.TempDir()
	executor := NewExecutor()

	ctx := context.Background()
	result, err := executor.Execute(ctx, "echo error >&2", workDir, nil)

	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if strings.TrimSpace(result.Stderr) != "error" {
		t.Errorf("expected stderr 'error', got %q", result.Stderr)
	}
}

func TestExecute_WithTruncate(t *testing.T) {
	workDir := t.TempDir()
	executor := NewExecutor()

	ctx := context.Background()
	truncateOpts := &TruncateOptions{
		MaxLines: 2,
		MaxBytes: 100,
	}

	result, err := executor.Execute(ctx, "echo -e 'line1\nline2\nline3\nline4'", workDir, truncateOpts)

	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if !result.Truncated {
		t.Error("expected output to be truncated")
	}

	if !strings.Contains(result.Output, "truncated") {
		t.Errorf("expected output to contain truncation message, got %q", result.Output)
	}
}

func TestTruncateOutput_NoTruncation(t *testing.T) {
	output := "line1\nline2"
	result, truncated := truncateOutput(output, 10, 1000)

	if truncated {
		t.Error("expected no truncation")
	}

	if result != output {
		t.Errorf("expected %q, got %q", output, result)
	}
}

func TestTruncateOutput_TruncateByLines(t *testing.T) {
	output := "line1\nline2\nline3\nline4\nline5"
	result, truncated := truncateOutput(output, 2, 10000)

	if !truncated {
		t.Error("expected truncation")
	}

	if !strings.HasPrefix(result, "line1\nline2") {
		t.Errorf("expected result to start with first 2 lines, got %q", result)
	}

	if !strings.Contains(result, "truncated") {
		t.Errorf("expected truncation message in result")
	}
}

func TestTruncateOutput_TruncateByBytes(t *testing.T) {
	output := "line1\nline2\nline3\nline4\nline5"
	result, truncated := truncateOutput(output, 100, 12)

	if !truncated {
		t.Error("expected truncation")
	}

	if !strings.Contains(result, "truncated") {
		t.Errorf("expected truncation message in result")
	}
}

func TestCountLines(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"", 0},
		{"hello", 1},
		{"hello\n", 2},
		{"hello\nworld", 2},
		{"hello\nworld\n", 3},
		{"a\nb\nc\nd", 4},
	}

	for _, tt := range tests {
		result := countLines(tt.input)
		if result != tt.expected {
			t.Errorf("countLines(%q) = %d, expected %d", tt.input, result, tt.expected)
		}
	}
}

func TestSplitLines(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{"", nil},
		{"hello", []string{"hello"}},
		{"hello\n", []string{"hello"}},
		{"hello\nworld", []string{"hello", "world"}},
		{"hello\nworld\n", []string{"hello", "world"}},
		{"a\nb\nc", []string{"a", "b", "c"}},
	}

	for _, tt := range tests {
		result := splitLines(tt.input)
		if len(result) != len(tt.expected) {
			t.Errorf("splitLines(%q) = %v, expected %v", tt.input, result, tt.expected)
			continue
		}
		for i := range result {
			if result[i] != tt.expected[i] {
				t.Errorf("splitLines(%q)[%d] = %q, expected %q", tt.input, i, result[i], tt.expected[i])
			}
		}
	}
}

func TestFormatMetadata(t *testing.T) {
	lines := []string{"error1", "error2"}
	result := formatMetadata(lines)

	if !strings.HasPrefix(result, "<bash_metadata>\n") {
		t.Errorf("expected result to start with <bash_metadata>\\n, got %q", result)
	}

	if !strings.HasSuffix(result, "</bash_metadata>") {
		t.Errorf("expected result to end with </bash_metadata>, got %q", result)
	}

	if !strings.Contains(result, "error1\n") {
		t.Errorf("expected result to contain 'error1\\n', got %q", result)
	}

	if !strings.Contains(result, "error2\n") {
		t.Errorf("expected result to contain 'error2\\n', got %q", result)
	}
}

func TestFormatMetadata_Empty(t *testing.T) {
	result := formatMetadata(nil)

	expected := "<bash_metadata>\n</bash_metadata>"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
}

func TestExecute_DurationMs(t *testing.T) {
	workDir := t.TempDir()
	executor := NewExecutor()

	ctx := context.Background()
	result, err := executor.Execute(ctx, "sleep 0.1", workDir, nil)

	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if result.DurationMs < 100 {
		t.Errorf("expected duration >= 100ms, got %d ms", result.DurationMs)
	}
}

func TestExecute_CombinedOutput(t *testing.T) {
	workDir := t.TempDir()
	executor := NewExecutor()

	ctx := context.Background()
	result, err := executor.Execute(ctx, "echo stdout; echo stderr >&2", workDir, nil)

	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if !strings.Contains(result.Output, "stdout") {
		t.Errorf("expected output to contain 'stdout', got %q", result.Output)
	}

	if !strings.Contains(result.Output, "stderr") {
		t.Errorf("expected output to contain 'stderr', got %q", result.Output)
	}
}
