package jsonl

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func createTestFile(t *testing.T, lines []string) string {
	t.Helper()
	filePath := filepath.Join(t.TempDir(), "test.jsonl")
	content := strings.Join(lines, "\n")
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}
	return filePath
}

func TestCountLines(t *testing.T) {
	s := NewService()
	lines := []string{
		`{"id":1,"name":"alice"}`,
		`{"id":2,"name":"bob"}`,
		`{"id":3,"name":"charlie"}`,
	}
	filePath := createTestFile(t, lines)

	count, err := s.CountLines(filePath)
	if err != nil {
		t.Fatalf("CountLines() error = %v", err)
	}
	if count != 3 {
		t.Errorf("CountLines() = %d, want 3", count)
	}
}

func TestCountLinesEmptyFile(t *testing.T) {
	s := NewService()
	filePath := filepath.Join(t.TempDir(), "empty.jsonl")
	if err := os.WriteFile(filePath, []byte(""), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	count, err := s.CountLines(filePath)
	if err != nil {
		t.Fatalf("CountLines() error = %v", err)
	}
	if count != 0 {
		t.Errorf("CountLines() = %d, want 0", count)
	}
}

func TestCountLinesFileNotExist(t *testing.T) {
	s := NewService()
	_, err := s.CountLines("/nonexistent/file.jsonl")
	if err == nil {
		t.Error("CountLines() expected error for nonexistent file")
	}
}

func TestReadLines(t *testing.T) {
	s := NewService()
	lines := []string{
		`{"id":1}`,
		`{"id":2}`,
		`{"id":3}`,
		`{"id":4}`,
		`{"id":5}`,
	}
	filePath := createTestFile(t, lines)

	result, err := s.ReadLines(filePath, 0, 3)
	if err != nil {
		t.Fatalf("ReadLines() error = %v", err)
	}
	if len(result) != 3 {
		t.Fatalf("ReadLines() returned %d lines, want 3", len(result))
	}
	if result[0] != `{"id":1}` {
		t.Errorf("ReadLines()[0] = %q, want %q", result[0], `{"id":1}`)
	}
	if result[2] != `{"id":3}` {
		t.Errorf("ReadLines()[2] = %q, want %q", result[2], `{"id":3}`)
	}
}

func TestReadLinesWithOffset(t *testing.T) {
	s := NewService()
	lines := []string{
		`{"id":1}`,
		`{"id":2}`,
		`{"id":3}`,
		`{"id":4}`,
		`{"id":5}`,
	}
	filePath := createTestFile(t, lines)

	result, err := s.ReadLines(filePath, 2, 2)
	if err != nil {
		t.Fatalf("ReadLines() error = %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("ReadLines() returned %d lines, want 2", len(result))
	}
	if result[0] != `{"id":3}` {
		t.Errorf("ReadLines()[0] = %q, want %q", result[0], `{"id":3}`)
	}
	if result[1] != `{"id":4}` {
		t.Errorf("ReadLines()[1] = %q, want %q", result[1], `{"id":4}`)
	}
}

func TestReadLinesBeyondEnd(t *testing.T) {
	s := NewService()
	lines := []string{
		`{"id":1}`,
		`{"id":2}`,
	}
	filePath := createTestFile(t, lines)

	result, err := s.ReadLines(filePath, 1, 10)
	if err != nil {
		t.Fatalf("ReadLines() error = %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("ReadLines() returned %d lines, want 1", len(result))
	}
	if result[0] != `{"id":2}` {
		t.Errorf("ReadLines()[0] = %q, want %q", result[0], `{"id":2}`)
	}
}

func TestReadLinesStartBeyondEnd(t *testing.T) {
	s := NewService()
	lines := []string{
		`{"id":1}`,
		`{"id":2}`,
	}
	filePath := createTestFile(t, lines)

	result, err := s.ReadLines(filePath, 10, 5)
	if err != nil {
		t.Fatalf("ReadLines() error = %v", err)
	}
	if len(result) != 0 {
		t.Errorf("ReadLines() returned %d lines, want 0", len(result))
	}
}

func TestReadLinesFileNotExist(t *testing.T) {
	s := NewService()
	_, err := s.ReadLines("/nonexistent/file.jsonl", 0, 10)
	if err == nil {
		t.Error("ReadLines() expected error for nonexistent file")
	}
}

func TestAppendLineNewFile(t *testing.T) {
	s := NewService()
	filePath := filepath.Join(t.TempDir(), "new.jsonl")

	if err := s.AppendLine(filePath, `{"id":1}`); err != nil {
		t.Fatalf("AppendLine() error = %v", err)
	}

	content, _ := os.ReadFile(filePath)
	if string(content) != "{\"id\":1}\n" {
		t.Errorf("file content = %q, want %q", string(content), "{\"id\":1}\n")
	}
}

func TestAppendLineExistingFile(t *testing.T) {
	s := NewService()
	lines := []string{`{"id":1}`}
	filePath := createTestFile(t, lines)

	if err := s.AppendLine(filePath, `{"id":2}`); err != nil {
		t.Fatalf("AppendLine() error = %v", err)
	}

	count, err := s.CountLines(filePath)
	if err != nil {
		t.Fatalf("CountLines() error = %v", err)
	}
	if count != 2 {
		t.Errorf("CountLines() = %d, want 2", count)
	}

	result, err := s.ReadLines(filePath, 1, 1)
	if err != nil {
		t.Fatalf("ReadLines() error = %v", err)
	}
	if len(result) != 1 || result[0] != `{"id":2}` {
		t.Errorf("ReadLines() = %v, want [%q]", result, `{"id":2}`)
	}
}

func TestAppendLineFileWithTrailingNewline(t *testing.T) {
	s := NewService()
	filePath := filepath.Join(t.TempDir(), "trailing.jsonl")
	if err := os.WriteFile(filePath, []byte("{\"id\":1}\n"), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	if err := s.AppendLine(filePath, `{"id":2}`); err != nil {
		t.Fatalf("AppendLine() error = %v", err)
	}

	content, _ := os.ReadFile(filePath)
	if string(content) != "{\"id\":1}\n{\"id\":2}\n" {
		t.Errorf("file content = %q, want %q", string(content), "{\"id\":1}\n{\"id\":2}\n")
	}
}

func TestAppendLineMultiple(t *testing.T) {
	s := NewService()
	filePath := filepath.Join(t.TempDir(), "multi.jsonl")

	for i := 1; i <= 3; i++ {
		line := fmt.Sprintf(`{"id":%d}`, i)
		if err := s.AppendLine(filePath, line); err != nil {
			t.Fatalf("AppendLine() error = %v", err)
		}
	}

	count, err := s.CountLines(filePath)
	if err != nil {
		t.Fatalf("CountLines() error = %v", err)
	}
	if count != 3 {
		t.Errorf("CountLines() = %d, want 3", count)
	}
}
