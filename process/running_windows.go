//go:build windows

package process

import (
	"errors"
	"os"
	"time"

	"golang.org/x/sys/windows"
)

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
	return code == windows.STILL_ACTIVE
}

func kill(p *os.Process) error {
	if p == nil {
		return errors.New("process not started")
	}
	return p.Kill()
}

func signal(p *os.Process, sig os.Signal) error {
	if p == nil {
		return errors.New("process not started")
	}
	return p.Signal(sig)
}

func stop(p *os.Process, _ <-chan struct{}, _ time.Duration, _ ...os.Signal) error {
	if !running(p) {
		return nil
	}
	return kill(p)
}
