//go:build !windows

package process

import (
	"os/exec"
	"syscall"
)

// setSysProcAttr sets POSIX-specific process attributes.
func setSysProcAttr(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}
