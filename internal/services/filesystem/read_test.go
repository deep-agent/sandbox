package filesystem

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestReadFile(t *testing.T) {
	tmpDir := t.TempDir()
	m := NewManager(tmpDir)

	content := "Hello, World!\nLine 2\nLine 3"
	filePath := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	result, err := m.ReadFile(filePath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	if result != content {
		t.Errorf("ReadFile() = %q, want %q", result, content)
	}
}

func TestReadFileEmpty(t *testing.T) {
	tmpDir := t.TempDir()
	m := NewManager(tmpDir)

	filePath := filepath.Join(tmpDir, "empty.txt")
	if err := os.WriteFile(filePath, []byte(""), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	result, err := m.ReadFile(filePath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}

	if result != "" {
		t.Errorf("ReadFile() = %q, want empty string", result)
	}
}

func TestReadFileNotExist(t *testing.T) {
	tmpDir := t.TempDir()
	m := NewManager(tmpDir)

	_, err := m.ReadFile(filepath.Join(tmpDir, "nonexistent.txt"))
	if err == nil {
		t.Error("ReadFile() expected error for nonexistent file")
	}
}

func TestReadFileBase64(t *testing.T) {
	tmpDir := t.TempDir()
	m := NewManager(tmpDir)

	content := []byte{0x00, 0x01, 0x02, 0xFF, 0xFE, 0xFD}
	filePath := filepath.Join(tmpDir, "binary.bin")
	if err := os.WriteFile(filePath, content, 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	result, err := m.ReadFileBase64(filePath)
	if err != nil {
		t.Fatalf("ReadFileBase64() error = %v", err)
	}

	expected := base64.StdEncoding.EncodeToString(content)
	if result != expected {
		t.Errorf("ReadFileBase64() = %q, want %q", result, expected)
	}
}

func TestReadFileBase64Text(t *testing.T) {
	tmpDir := t.TempDir()
	m := NewManager(tmpDir)

	content := "Hello, World!"
	filePath := filepath.Join(tmpDir, "text.txt")
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	result, err := m.ReadFileBase64(filePath)
	if err != nil {
		t.Fatalf("ReadFileBase64() error = %v", err)
	}

	expected := base64.StdEncoding.EncodeToString([]byte(content))
	if result != expected {
		t.Errorf("ReadFileBase64() = %q, want %q", result, expected)
	}
}

func TestReadFileBase64NotExist(t *testing.T) {
	tmpDir := t.TempDir()
	m := NewManager(tmpDir)

	_, err := m.ReadFileBase64(filepath.Join(tmpDir, "nonexistent.txt"))
	if err == nil {
		t.Error("ReadFileBase64() expected error for nonexistent file")
	}
}

func TestReadFileWithOptionsBasic(t *testing.T) {
	tmpDir := t.TempDir()
	m := NewManager(tmpDir)

	lines := []string{"Line 1", "Line 2", "Line 3", "Line 4", "Line 5"}
	content := strings.Join(lines, "\n")
	filePath := filepath.Join(tmpDir, "multiline.txt")
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	result, err := m.ReadFileWithOptions(filePath, ReadOptions{})
	if err != nil {
		t.Fatalf("ReadFileWithOptions() error = %v", err)
	}

	if result.LinesRead != 5 {
		t.Errorf("ReadFileWithOptions() LinesRead = %d, want 5", result.LinesRead)
	}
}

func TestReadFileWithOptionsOffset(t *testing.T) {
	tmpDir := t.TempDir()
	m := NewManager(tmpDir)

	lines := []string{"Line 1", "Line 2", "Line 3", "Line 4", "Line 5"}
	content := strings.Join(lines, "\n")
	filePath := filepath.Join(tmpDir, "multiline.txt")
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	result, err := m.ReadFileWithOptions(filePath, ReadOptions{Offset: 3})
	if err != nil {
		t.Fatalf("ReadFileWithOptions() error = %v", err)
	}

	if result.LinesRead != 3 {
		t.Errorf("ReadFileWithOptions() LinesRead = %d, want 3", result.LinesRead)
	}

	if !strings.Contains(result.Content, "Line 3") {
		t.Error("ReadFileWithOptions() content should start from Line 3")
	}

	if strings.Contains(result.Content, "Line 1") || strings.Contains(result.Content, "Line 2") {
		t.Error("ReadFileWithOptions() content should not contain Line 1 or Line 2")
	}
}

func TestReadFileWithOptionsLimit(t *testing.T) {
	tmpDir := t.TempDir()
	m := NewManager(tmpDir)

	lines := []string{"Line 1", "Line 2", "Line 3", "Line 4", "Line 5"}
	content := strings.Join(lines, "\n")
	filePath := filepath.Join(tmpDir, "multiline.txt")
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	result, err := m.ReadFileWithOptions(filePath, ReadOptions{Limit: 2})
	if err != nil {
		t.Fatalf("ReadFileWithOptions() error = %v", err)
	}

	if result.LinesRead != 2 {
		t.Errorf("ReadFileWithOptions() LinesRead = %d, want 2", result.LinesRead)
	}

	if !strings.Contains(result.Content, "Line 1") || !strings.Contains(result.Content, "Line 2") {
		t.Error("ReadFileWithOptions() should contain Line 1 and Line 2")
	}

	if strings.Contains(result.Content, "Line 3") {
		t.Error("ReadFileWithOptions() should not contain Line 3")
	}
}

func TestReadFileWithOptionsOffsetAndLimit(t *testing.T) {
	tmpDir := t.TempDir()
	m := NewManager(tmpDir)

	lines := []string{"Line 1", "Line 2", "Line 3", "Line 4", "Line 5"}
	content := strings.Join(lines, "\n")
	filePath := filepath.Join(tmpDir, "multiline.txt")
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	result, err := m.ReadFileWithOptions(filePath, ReadOptions{Offset: 2, Limit: 2})
	if err != nil {
		t.Fatalf("ReadFileWithOptions() error = %v", err)
	}

	if result.LinesRead != 2 {
		t.Errorf("ReadFileWithOptions() LinesRead = %d, want 2", result.LinesRead)
	}

	if !strings.Contains(result.Content, "Line 2") || !strings.Contains(result.Content, "Line 3") {
		t.Error("ReadFileWithOptions() should contain Line 2 and Line 3")
	}
}

func TestReadFileWithOptionsLineNumbers(t *testing.T) {
	tmpDir := t.TempDir()
	m := NewManager(tmpDir)

	lines := []string{"Line 1", "Line 2", "Line 3"}
	content := strings.Join(lines, "\n")
	filePath := filepath.Join(tmpDir, "multiline.txt")
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	result, err := m.ReadFileWithOptions(filePath, ReadOptions{WithLineNumber: true})
	if err != nil {
		t.Fatalf("ReadFileWithOptions() error = %v", err)
	}

	if !strings.Contains(result.Content, "1\t") {
		t.Error("ReadFileWithOptions() should contain line number 1")
	}
	if !strings.Contains(result.Content, "2\t") {
		t.Error("ReadFileWithOptions() should contain line number 2")
	}
}

func TestReadFileWithOptionsMaxLineLength(t *testing.T) {
	tmpDir := t.TempDir()
	m := NewManager(tmpDir)

	longLine := strings.Repeat("a", 100)
	filePath := filepath.Join(tmpDir, "longline.txt")
	if err := os.WriteFile(filePath, []byte(longLine), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	result, err := m.ReadFileWithOptions(filePath, ReadOptions{MaxLineLength: 10})
	if err != nil {
		t.Fatalf("ReadFileWithOptions() error = %v", err)
	}

	if !strings.Contains(result.Content, "...") {
		t.Error("ReadFileWithOptions() should truncate long lines with ...")
	}

	contentWithoutNewline := strings.TrimSuffix(result.Content, "\n")
	if len(contentWithoutNewline) > 15 {
		t.Errorf("ReadFileWithOptions() truncated line too long: %d", len(contentWithoutNewline))
	}
}

func TestReadFileWithOptionsTotalLine(t *testing.T) {
	tmpDir := t.TempDir()
	m := NewManager(tmpDir)

	lines := []string{"Line 1", "Line 2", "Line 3", "Line 4", "Line 5"}
	content := strings.Join(lines, "\n")
	filePath := filepath.Join(tmpDir, "multiline.txt")
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	result, err := m.ReadFileWithOptions(filePath, ReadOptions{Limit: 2})
	if err != nil {
		t.Fatalf("ReadFileWithOptions() error = %v", err)
	}

	if result.TotalLine < 2 {
		t.Errorf("ReadFileWithOptions() TotalLine = %d, should be at least 2", result.TotalLine)
	}
}

func TestReadFileWithOptionsNotExist(t *testing.T) {
	tmpDir := t.TempDir()
	m := NewManager(tmpDir)

	_, err := m.ReadFileWithOptions(filepath.Join(tmpDir, "nonexistent.txt"), ReadOptions{})
	if err == nil {
		t.Error("ReadFileWithOptions() expected error for nonexistent file")
	}
}
