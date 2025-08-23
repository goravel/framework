//go:build windows

package process

import "os/exec"

// setSysProcAttr is a no-op on Windows; Setpgid isn't available.
func setSysProcAttr(cmd *exec.Cmd) {}
