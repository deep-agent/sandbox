//go:build !windows

package bash

import (
	"os/exec"
	"syscall"

	"github.com/deep-agent/sandbox/pkg/logger"
)

// bashSysProcAttr returns Unix process attributes that place the child in its
// own process group so the entire tree can be killed with kill(-pgid).
func bashSysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{
		Setpgid: true,
	}
}

// killProcessGroup kills the entire process group of cmd.
func killProcessGroup(cmd *exec.Cmd) {
	if err := syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL); err != nil {
		logger.Printf("background command kill error: %v", err)
	}
}
