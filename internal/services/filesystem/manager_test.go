package filesystem

import (
	"testing"
)

func TestValidatePath(t *testing.T) {
	m := NewManager("/tmp/workspace")

	tests := []struct {
		name      string
		path      string
		wantErr   bool
		checkAbs  bool
	}{
		{
			name:     "relative path .task",
			path:     ".task",
			wantErr:  false,
			checkAbs: true,
		},
		{
			name:     "absolute path",
			path:     "/tmp/workspace/file.txt",
			wantErr:  false,
			checkAbs: true,
		},
		{
			name:     "path with dot dot",
			path:     "../outside",
			wantErr:  false,
			checkAbs: true,
		},
		{
			name:     "current directory",
			path:     ".",
			wantErr:  false,
			checkAbs: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			absPath, err := m.validatePath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("validatePath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.checkAbs && absPath == "" {
				t.Errorf("validatePath() returned empty path")
			}
			t.Logf("input: %q -> output: %q", tt.path, absPath)
		})
	}
}
