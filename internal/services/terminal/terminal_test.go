package terminal

import (
	"io"
	"strings"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		shell   string
		wantErr bool
	}{
		{
			name:    "default shell",
			shell:   "",
			wantErr: false,
		},
		{
			name:    "explicit bash",
			shell:   "/bin/bash",
			wantErr: false,
		},
		{
			name:    "explicit zsh",
			shell:   "/bin/zsh",
			wantErr: false,
		},
		{
			name:    "invalid shell",
			shell:   "/nonexistent/shell",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			workDir := t.TempDir()
			term, err := New(tt.shell, workDir, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil {
				defer term.Close()
				if term.cmd == nil {
					t.Error("New() returned terminal with nil cmd")
				}
				if term.ptmx == nil {
					t.Error("New() returned terminal with nil ptmx")
				}
			}
		})
	}
}

func TestTerminal_ReadWrite(t *testing.T) {
	workDir := t.TempDir()
	term, err := New("/bin/bash", workDir, []string{"PS1="})
	if err != nil {
		t.Fatalf("Failed to create terminal: %v", err)
	}
	defer term.Close()

	testCmd := "echo hello\n"
	n, err := term.Write([]byte(testCmd))
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}
	if n != len(testCmd) {
		t.Errorf("Write() wrote %d bytes, want %d", n, len(testCmd))
	}

	buf := make([]byte, 1024)
	var output strings.Builder
	deadline := time.Now().Add(3 * time.Second)

	for time.Now().Before(deadline) {
		n, err = term.Read(buf)
		if err != nil && err != io.EOF {
			break
		}
		if n > 0 {
			output.Write(buf[:n])
			if strings.Contains(output.String(), "hello") {
				break
			}
		}
	}

	if !strings.Contains(output.String(), "hello") {
		t.Errorf("Read() output = %q, want to contain 'hello'", output.String())
	}
}

func TestTerminal_Resize(t *testing.T) {
	workDir := t.TempDir()
	term, err := New("/bin/bash", workDir, nil)
	if err != nil {
		t.Fatalf("Failed to create terminal: %v", err)
	}
	defer term.Close()

	tests := []struct {
		name string
		size Size
	}{
		{
			name: "standard size",
			size: Size{Rows: 24, Cols: 80},
		},
		{
			name: "large size",
			size: Size{Rows: 100, Cols: 200},
		},
		{
			name: "small size",
			size: Size{Rows: 10, Cols: 40},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := term.Resize(tt.size)
			if err != nil {
				t.Errorf("Resize() error = %v", err)
			}
		})
	}
}

func TestTerminal_Resize_AfterClose(t *testing.T) {
	workDir := t.TempDir()
	term, err := New("/bin/bash", workDir, nil)
	if err != nil {
		t.Fatalf("Failed to create terminal: %v", err)
	}

	term.Close()

	err = term.Resize(Size{Rows: 24, Cols: 80})
	if err != io.EOF {
		t.Errorf("Resize() after close error = %v, want io.EOF", err)
	}
}

func TestTerminal_Close(t *testing.T) {
	workDir := t.TempDir()
	term, err := New("/bin/bash", workDir, nil)
	if err != nil {
		t.Fatalf("Failed to create terminal: %v", err)
	}

	err = term.Close()
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}

	if !term.closed {
		t.Error("Close() did not set closed to true")
	}

	err = term.Close()
	if err != nil {
		t.Errorf("Close() called twice error = %v, want nil", err)
	}
}

func TestTerminal_Wait(t *testing.T) {
	workDir := t.TempDir()
	term, err := New("/bin/bash", workDir, nil)
	if err != nil {
		t.Fatalf("Failed to create terminal: %v", err)
	}

	done := make(chan error, 1)
	go func() {
		done <- term.Wait()
	}()

	err = term.Close()
	if err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Error("Wait() timed out")
	}
}
