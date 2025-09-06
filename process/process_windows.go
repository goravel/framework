//go:build windows

// TODO: Revisit process handling for windows, need a better way to handle process using native APIs
package process

import (
	"os"
	"os/exec"
	"time"

	"golang.org/x/sys/windows"

	"github.com/goravel/framework/errors"
)

// stillActive is a Win32 constant that indicates a process is still running.
// It is not exported by the Go standard library, so we define it here.
const stillActive = 259

// setSysProcAttr is a no-op on Windows; Setpgid isn't available.
func setSysProcAttr(cmd *exec.Cmd) {}

func running(p *os.Process) bool {
	if p == nil {
		return false
	}

	h, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, uint32(p.Pid))
	if err != nil {
		// If we cannot open the process (access denied or not found), assume not running.
		return false
	}
	defer windows.CloseHandle(h)

	var code uint32
	if err := windows.GetExitCodeProcess(h, &code); err != nil {
		return false
	}
	return code == stillActive
}

func kill(p *os.Process) error {
	if p == nil {
		return errors.ProcessNotStarted
	}
	return p.Kill()
}

func signal(p *os.Process, sig os.Signal) error {
	if p == nil {
		return errors.ProcessNotStarted
	}
	return p.Signal(sig)
}

func stop(p *os.Process, _ <-chan struct{}, _ time.Duration, _ ...os.Signal) error {
	if !running(p) {
		return nil
	}
	return kill(p)
}
