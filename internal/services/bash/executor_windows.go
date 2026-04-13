//go:build windows

package bash

import (
	"os/exec"
	"syscall"

	"github.com/deep-agent/sandbox/pkg/logger"
)

// bashSysProcAttr returns Windows process attributes that place the child in
// its own process group so the tree can be killed with taskkill /T.
func bashSysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
	}
}

// killProcessGroup kills the process and its children on Windows.
func killProcessGroup(cmd *exec.Cmd) {
	if err := cmd.Process.Kill(); err != nil {
		logger.Printf("background command kill error: %v", err)
	}
}
