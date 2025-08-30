//go:build windows

package process

import (
	"errors"
	"os"
	"time"
)

func running(p *os.Process) bool {
	if p == nil {
		return false
	}

	_, err := os.FindProcess(p.Pid)
	return err == nil
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
