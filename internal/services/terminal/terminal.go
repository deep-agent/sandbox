package terminal

import (
	"io"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"unsafe"

	"github.com/creack/pty"
)

type Terminal struct {
	cmd    *exec.Cmd
	ptmx   *os.File
	mu     sync.Mutex
	closed bool
}

type Size struct {
	Rows uint16 `json:"rows"`
	Cols uint16 `json:"cols"`
}

func New(shell string, workDir string, env []string) (*Terminal, error) {
	if shell == "" {
		shell = "/bin/bash"
	}

	cmd := exec.Command(shell)
	cmd.Dir = workDir
	cmd.Env = append(os.Environ(), env...)

	ptmx, err := pty.Start(cmd)
	if err != nil {
		return nil, err
	}

	return &Terminal{
		cmd:  cmd,
		ptmx: ptmx,
	}, nil
}

func (t *Terminal) Read(p []byte) (n int, err error) {
	return t.ptmx.Read(p)
}

func (t *Terminal) Write(p []byte) (n int, err error) {
	return t.ptmx.Write(p)
}

func (t *Terminal) Resize(size Size) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed {
		return io.EOF
	}

	ws := struct {
		Row    uint16
		Col    uint16
		Xpixel uint16
		Ypixel uint16
	}{
		Row: size.Rows,
		Col: size.Cols,
	}

	_, _, errno := syscall.Syscall(
		syscall.SYS_IOCTL,
		t.ptmx.Fd(),
		syscall.TIOCSWINSZ,
		uintptr(unsafe.Pointer(&ws)),
	)

	if errno != 0 {
		return errno
	}

	return nil
}

func (t *Terminal) Close() error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.closed {
		return nil
	}
	t.closed = true

	if t.cmd.Process != nil {
		t.cmd.Process.Signal(syscall.SIGTERM)
	}

	return t.ptmx.Close()
}

func (t *Terminal) Wait() error {
	return t.cmd.Wait()
}
