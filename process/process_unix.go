//go:build !windows

package process

import (
	"errors"
	"os"
	"os/exec"
	"time"

	"golang.org/x/sys/unix"
)

func setSysProcAttr(cmd *exec.Cmd) {
	cmd.SysProcAttr = &unix.SysProcAttr{Setpgid: true}
}

func running(p *os.Process) bool {
	if p == nil {
		return false
	}

	err := p.Signal(unix.Signal(0))
	return err == nil
}

func kill(p *os.Process) error {
	return signal(p, unix.SIGKILL)
}

func signal(p *os.Process, sig os.Signal) error {
	if p == nil {
		return errors.New("process not started")
	}

	return p.Signal(sig)
}

func stop(p *os.Process, done <-chan struct{}, timeout time.Duration, sig ...os.Signal) error {
	if !running(p) {
		return nil
	}

	var signalToSend os.Signal = unix.SIGTERM
	if len(sig) > 0 {
		signalToSend = sig[0]
	}

	if err := signal(p, signalToSend); err != nil {
		if errors.Is(err, os.ErrProcessDone) {
			return nil
		}
		return err
	}

	select {
	case <-done:
		return nil
	case <-time.After(timeout):
		return signal(p, unix.SIGKILL)
	}
}
