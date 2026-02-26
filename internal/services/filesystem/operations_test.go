package filesystem

import (
	"os"
	"path/filepath"
	"testing"
)

func TestListDir(t *testing.T) {
	tmpDir := t.TempDir()
	m := NewManager()

	subDir := filepath.Join(tmpDir, "subdir")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("failed to create subdir: %v", err)
	}

	if err := os.WriteFile(filepath.Join(tmpDir, "file1.txt"), []byte("content1"), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "file2.go"), []byte("package main"), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	files, err := m.ListDir(tmpDir)
	if err != nil {
		t.Fatalf("ListDir() error = %v", err)
	}

	if len(files) != 3 {
		t.Errorf("ListDir() returned %d files, want 3", len(files))
	}

	foundDir := false
	foundFile := false
	for _, f := range files {
		if f.Name == "subdir" && f.IsDir {
			foundDir = true
		}
		if f.Name == "file1.txt" && !f.IsDir {
			foundFile = true
		}
	}

	if !foundDir {
		t.Error("ListDir() did not return subdir")
	}
	if !foundFile {
		t.Error("ListDir() did not return file1.txt")
	}
}

func TestListDirEmpty(t *testing.T) {
	tmpDir := t.TempDir()
	m := NewManager()

	files, err := m.ListDir(tmpDir)
	if err != nil {
		t.Fatalf("ListDir() error = %v", err)
	}

	if len(files) != 0 {
		t.Errorf("ListDir() returned %d files, want 0", len(files))
	}
}

func TestListDirNotExist(t *testing.T) {
	tmpDir := t.TempDir()
	m := NewManager()

	_, err := m.ListDir(filepath.Join(tmpDir, "nonexistent"))
	if err == nil {
		t.Error("ListDir() expected error for nonexistent directory")
	}
}

func TestDeleteFile(t *testing.T) {
	tmpDir := t.TempDir()
	m := NewManager()

	filePath := filepath.Join(tmpDir, "to_delete.txt")
	if err := os.WriteFile(filePath, []byte("delete me"), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	if err := m.DeleteFile(filePath); err != nil {
		t.Fatalf("DeleteFile() error = %v", err)
	}

	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		t.Error("DeleteFile() file still exists")
	}
}

func TestDeleteFileDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	m := NewManager()

	dirPath := filepath.Join(tmpDir, "dir_to_delete")
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		t.Fatalf("failed to create directory: %v", err)
	}

	if err := os.WriteFile(filepath.Join(dirPath, "file.txt"), []byte("content"), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	if err := m.DeleteFile(dirPath); err != nil {
		t.Fatalf("DeleteFile() error = %v", err)
	}

	if _, err := os.Stat(dirPath); !os.IsNotExist(err) {
		t.Error("DeleteFile() directory still exists")
	}
}

func TestMoveFile(t *testing.T) {
	tmpDir := t.TempDir()
	m := NewManager()

	srcPath := filepath.Join(tmpDir, "source.txt")
	dstPath := filepath.Join(tmpDir, "destination.txt")

	content := "test content"
	if err := os.WriteFile(srcPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create source file: %v", err)
	}

	if err := m.MoveFile(srcPath, dstPath); err != nil {
		t.Fatalf("MoveFile() error = %v", err)
	}

	if _, err := os.Stat(srcPath); !os.IsNotExist(err) {
		t.Error("MoveFile() source file still exists")
	}

	data, err := os.ReadFile(dstPath)
	if err != nil {
		t.Fatalf("failed to read destination file: %v", err)
	}

	if string(data) != content {
		t.Errorf("MoveFile() content = %q, want %q", string(data), content)
	}
}

func TestMoveFileSourceNotExist(t *testing.T) {
	tmpDir := t.TempDir()
	m := NewManager()

	err := m.MoveFile(filepath.Join(tmpDir, "nonexistent.txt"), filepath.Join(tmpDir, "dest.txt"))
	if err == nil {
		t.Error("MoveFile() expected error for nonexistent source")
	}
}

func TestCopyFile(t *testing.T) {
	tmpDir := t.TempDir()
	m := NewManager()

	srcPath := filepath.Join(tmpDir, "source.txt")
	dstPath := filepath.Join(tmpDir, "copy.txt")

	content := "test content to copy"
	if err := os.WriteFile(srcPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create source file: %v", err)
	}

	if err := m.CopyFile(srcPath, dstPath); err != nil {
		t.Fatalf("CopyFile() error = %v", err)
	}

	srcData, err := os.ReadFile(srcPath)
	if err != nil {
		t.Fatalf("failed to read source file: %v", err)
	}
	if string(srcData) != content {
		t.Error("CopyFile() source file was modified")
	}

	dstData, err := os.ReadFile(dstPath)
	if err != nil {
		t.Fatalf("failed to read destination file: %v", err)
	}
	if string(dstData) != content {
		t.Errorf("CopyFile() destination content = %q, want %q", string(dstData), content)
	}
}

func TestCopyFileCreatesDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	m := NewManager()

	srcPath := filepath.Join(tmpDir, "source.txt")
	dstPath := filepath.Join(tmpDir, "newdir", "subdir", "copy.txt")

	if err := os.WriteFile(srcPath, []byte("content"), 0644); err != nil {
		t.Fatalf("failed to create source file: %v", err)
	}

	if err := m.CopyFile(srcPath, dstPath); err != nil {
		t.Fatalf("CopyFile() error = %v", err)
	}

	if _, err := os.Stat(dstPath); os.IsNotExist(err) {
		t.Error("CopyFile() destination file was not created")
	}
}

func TestCopyFileSourceNotExist(t *testing.T) {
	tmpDir := t.TempDir()
	m := NewManager()

	err := m.CopyFile(filepath.Join(tmpDir, "nonexistent.txt"), filepath.Join(tmpDir, "dest.txt"))
	if err == nil {
		t.Error("CopyFile() expected error for nonexistent source")
	}
}

func TestExists(t *testing.T) {
	tmpDir := t.TempDir()
	m := NewManager()

	filePath := filepath.Join(tmpDir, "exists.txt")
	if err := os.WriteFile(filePath, []byte("content"), 0644); err != nil {
		t.Fatalf("failed to create file: %v", err)
	}

	if !m.Exists(filePath) {
		t.Error("Exists() returned false for existing file")
	}

	if m.Exists(filepath.Join(tmpDir, "nonexistent.txt")) {
		t.Error("Exists() returned true for nonexistent file")
	}
}

func TestExistsDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	m := NewManager()

	dirPath := filepath.Join(tmpDir, "existsdir")
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		t.Fatalf("failed to create directory: %v", err)
	}

	if !m.Exists(dirPath) {
		t.Error("Exists() returned false for existing directory")
	}
}

func TestMkDir(t *testing.T) {
	tmpDir := t.TempDir()
	m := NewManager()

	dirPath := filepath.Join(tmpDir, "newdir")
	if err := m.MkDir(dirPath); err != nil {
		t.Fatalf("MkDir() error = %v", err)
	}

	info, err := os.Stat(dirPath)
	if err != nil {
		t.Fatalf("failed to stat directory: %v", err)
	}

	if !info.IsDir() {
		t.Error("MkDir() did not create a directory")
	}
}

func TestMkDirNested(t *testing.T) {
	tmpDir := t.TempDir()
	m := NewManager()

	dirPath := filepath.Join(tmpDir, "a", "b", "c")
	if err := m.MkDir(dirPath); err != nil {
		t.Fatalf("MkDir() error = %v", err)
	}

	info, err := os.Stat(dirPath)
	if err != nil {
		t.Fatalf("failed to stat directory: %v", err)
	}

	if !info.IsDir() {
		t.Error("MkDir() did not create nested directories")
	}
}

func TestMkDirAlreadyExists(t *testing.T) {
	tmpDir := t.TempDir()
	m := NewManager()

	dirPath := filepath.Join(tmpDir, "existsdir")
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		t.Fatalf("failed to create directory: %v", err)
	}

	if err := m.MkDir(dirPath); err != nil {
		t.Fatalf("MkDir() should not error for existing directory, got: %v", err)
	}
}
