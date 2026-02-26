package filesystem

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEditFile(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewManager()

	tests := []struct {
		name           string
		initialContent string
		oldString      string
		newString      string
		opts           EditOptions
		wantContent    string
		wantErr        bool
		errContains    string
	}{
		{
			name:           "simple replacement",
			initialContent: "hello world",
			oldString:      "world",
			newString:      "golang",
			opts:           EditOptions{},
			wantContent:    "hello golang",
			wantErr:        false,
		},
		{
			name:           "multiline replacement",
			initialContent: "line1\nline2\nline3",
			oldString:      "line2",
			newString:      "newline2",
			opts:           EditOptions{},
			wantContent:    "line1\nnewline2\nline3",
			wantErr:        false,
		},
		{
			name:           "replace all occurrences",
			initialContent: "foo bar foo baz foo",
			oldString:      "foo",
			newString:      "qux",
			opts:           EditOptions{ReplaceAll: true},
			wantContent:    "qux bar qux baz qux",
			wantErr:        false,
		},
		{
			name:           "multiple matches without replace all",
			initialContent: "foo bar foo",
			oldString:      "foo",
			newString:      "qux",
			opts:           EditOptions{ReplaceAll: false},
			wantErr:        true,
			errContains:    "old_string not found",
		},
		{
			name:           "old string equals new string",
			initialContent: "hello world",
			oldString:      "world",
			newString:      "world",
			opts:           EditOptions{},
			wantErr:        true,
			errContains:    "old_string and new_string must be different",
		},
		{
			name:           "old string not found",
			initialContent: "hello world",
			oldString:      "notfound",
			newString:      "replacement",
			opts:           EditOptions{},
			wantErr:        true,
			errContains:    "old_string not found",
		},
		{
			name:           "replace with empty string",
			initialContent: "hello world",
			oldString:      " world",
			newString:      "",
			opts:           EditOptions{},
			wantContent:    "hello",
			wantErr:        false,
		},
		{
			name:           "replace entire content",
			initialContent: "hello",
			oldString:      "hello",
			newString:      "goodbye",
			opts:           EditOptions{},
			wantContent:    "goodbye",
			wantErr:        false,
		},
		{
			name:           "multiline block replacement",
			initialContent: "func main() {\n\tfmt.Println(\"hello\")\n}",
			oldString:      "func main() {\n\tfmt.Println(\"hello\")\n}",
			newString:      "func main() {\n\tfmt.Println(\"goodbye\")\n}",
			opts:           EditOptions{},
			wantContent:    "func main() {\n\tfmt.Println(\"goodbye\")\n}",
			wantErr:        false,
		},
		{
			name:           "trimmed whitespace matching",
			initialContent: "  hello world  \n  foo bar  ",
			oldString:      "hello world",
			newString:      "hi there",
			opts:           EditOptions{},
			wantContent:    "  hi there  \n  foo bar  ",
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testFile := filepath.Join(tmpDir, tt.name+".txt")
			if err := os.WriteFile(testFile, []byte(tt.initialContent), 0644); err != nil {
				t.Fatalf("failed to create test file: %v", err)
			}

			err := manager.EditFile(testFile, tt.oldString, tt.newString, tt.opts)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.errContains)
					return
				}
				if tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("expected error containing %q, got %q", tt.errContains, err.Error())
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			content, err := os.ReadFile(testFile)
			if err != nil {
				t.Fatalf("failed to read test file: %v", err)
			}

			if string(content) != tt.wantContent {
				t.Errorf("content mismatch:\ngot:  %q\nwant: %q", string(content), tt.wantContent)
			}
		})
	}
}

func TestEditFile_FileNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewManager()

	err := manager.EditFile(filepath.Join(tmpDir, "nonexistent.txt"), "old", "new", EditOptions{})
	if err == nil {
		t.Error("expected error for non-existent file, got nil")
	}
	if !contains(err.Error(), "failed to read file") {
		t.Errorf("expected error about reading file, got: %v", err)
	}
}

func TestEditFile_LineTrimmedMatching(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewManager()

	initialContent := "    func foo() {\n        return 1\n    }"
	testFile := filepath.Join(tmpDir, "trimmed.txt")
	if err := os.WriteFile(testFile, []byte(initialContent), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	oldString := "func foo() {\n    return 1\n}"
	newString := "func bar() {\n    return 2\n}"

	err := manager.EditFile(testFile, oldString, newString, EditOptions{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("failed to read test file: %v", err)
	}

	if string(content) != newString {
		t.Errorf("content mismatch:\ngot:  %q\nwant: %q", string(content), newString)
	}
}

func TestEditFile_PreservesOtherContent(t *testing.T) {
	tmpDir := t.TempDir()
	manager := NewManager()

	initialContent := "header\n\nfunc old() {}\n\nfooter"
	testFile := filepath.Join(tmpDir, "preserve.txt")
	if err := os.WriteFile(testFile, []byte(initialContent), 0644); err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	err := manager.EditFile(testFile, "func old() {}", "func new() {}", EditOptions{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	content, err := os.ReadFile(testFile)
	if err != nil {
		t.Fatalf("failed to read test file: %v", err)
	}

	expected := "header\n\nfunc new() {}\n\nfooter"
	if string(content) != expected {
		t.Errorf("content mismatch:\ngot:  %q\nwant: %q", string(content), expected)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && findSubstring(s, substr)))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
