package filesystem

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFormatGlobOutput(t *testing.T) {
	tests := []struct {
		name      string
		files     []string
		truncated bool
		want      string
	}{
		{
			name:      "empty files",
			files:     []string{},
			truncated: false,
			want:      "No files found",
		},
		{
			name:      "single file",
			files:     []string{"/path/to/file.txt"},
			truncated: false,
			want:      "/path/to/file.txt",
		},
		{
			name:      "multiple files",
			files:     []string{"/path/to/file1.txt", "/path/to/file2.txt"},
			truncated: false,
			want:      "/path/to/file1.txt\n/path/to/file2.txt",
		},
		{
			name:      "truncated results",
			files:     []string{"/path/to/file.txt"},
			truncated: true,
			want:      "/path/to/file.txt\n\n(Results are truncated. Consider using a more specific path or pattern.)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatGlobOutput(tt.files, tt.truncated)
			if got != tt.want {
				t.Errorf("formatGlobOutput() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestMatchGlobPattern(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		path    string
		want    bool
	}{
		{
			name:    "exact match",
			pattern: "file.txt",
			path:    "file.txt",
			want:    true,
		},
		{
			name:    "wildcard extension",
			pattern: "*.txt",
			path:    "file.txt",
			want:    true,
		},
		{
			name:    "wildcard extension no match",
			pattern: "*.go",
			path:    "file.txt",
			want:    false,
		},
		{
			name:    "double star prefix",
			pattern: "**/*.go",
			path:    "src/main.go",
			want:    true,
		},
		{
			name:    "double star prefix nested",
			pattern: "**/*.go",
			path:    "src/pkg/main.go",
			want:    true,
		},
		{
			name:    "double star with prefix path",
			pattern: "src/**/*.go",
			path:    "src/pkg/main.go",
			want:    true,
		},
		{
			name:    "double star with prefix path no match",
			pattern: "src/**/*.go",
			path:    "other/pkg/main.go",
			want:    false,
		},
		{
			name:    "double star only",
			pattern: "**",
			path:    "any/path/file.txt",
			want:    true,
		},
		{
			name:    "question mark wildcard",
			pattern: "file?.txt",
			path:    "file1.txt",
			want:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := matchGlobPattern(tt.pattern, tt.path)
			if got != tt.want {
				t.Errorf("matchGlobPattern(%q, %q) = %v, want %v", tt.pattern, tt.path, got, tt.want)
			}
		})
	}
}

func TestGlobWithWalk(t *testing.T) {
	tmpDir := t.TempDir()
	m := NewManager(tmpDir)

	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("failed to create subdir: %v", err)
	}

	files := map[string]string{
		"file1.txt":        "content1",
		"file2.txt":        "content2",
		"file3.go":         "package main",
		"subdir/file4.txt": "content4",
		"subdir/file5.go":  "package sub",
	}

	for name, content := range files {
		path := filepath.Join(tmpDir, name)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create file %s: %v", name, err)
		}
	}

	tests := []struct {
		name       string
		pattern    string
		limit      int
		wantCount  int
		wantTrunc  bool
	}{
		{
			name:      "match all txt files recursively",
			pattern:   "*.txt",
			limit:     0,
			wantCount: 3,
			wantTrunc: false,
		},
		{
			name:      "match all go files recursively",
			pattern:   "*.go",
			limit:     0,
			wantCount: 2,
			wantTrunc: false,
		},
		{
			name:      "match with limit",
			pattern:   "*.txt",
			limit:     1,
			wantCount: 1,
			wantTrunc: true,
		},
		{
			name:      "no match",
			pattern:   "*.md",
			limit:     0,
			wantCount: 0,
			wantTrunc: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := m.globWithWalk(tmpDir, tt.pattern, tt.limit)
			if err != nil {
				t.Fatalf("globWithWalk() error = %v", err)
			}

			if result.Count != tt.wantCount {
				t.Errorf("globWithWalk() count = %d, want %d", result.Count, tt.wantCount)
			}

			if result.Truncated != tt.wantTrunc {
				t.Errorf("globWithWalk() truncated = %v, want %v", result.Truncated, tt.wantTrunc)
			}

			if len(result.Files) != result.Count {
				t.Errorf("globWithWalk() files length = %d, count = %d", len(result.Files), result.Count)
			}
		})
	}
}

func TestGlobWithWalkDoubleStarPattern(t *testing.T) {
	tmpDir := t.TempDir()
	m := NewManager(tmpDir)

	subDir := filepath.Join(tmpDir, "a", "b")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("failed to create subdir: %v", err)
	}

	files := []string{
		"root.go",
		"a/mid.go",
		"a/b/deep.go",
	}

	for _, name := range files {
		path := filepath.Join(tmpDir, name)
		if err := os.WriteFile(path, []byte("package main"), 0644); err != nil {
			t.Fatalf("failed to create file %s: %v", name, err)
		}
	}

	result, err := m.globWithWalk(tmpDir, "**/*.go", 0)
	if err != nil {
		t.Fatalf("globWithWalk() error = %v", err)
	}

	if result.Count != 3 {
		t.Errorf("globWithWalk() count = %d, want 3", result.Count)
	}
}
