//go:build !windows

package process

import (
	"errors"
	"os"
	"time"

	"golang.org/x/sys/unix"
)

// Running actively queries the OS to see if the process still exists.
//
// NOTE: This method returns `true` for a "zombie" process (one that has terminated
// but has not been reaped by a call to `Wait()`). Due to this, a monitoring loop
// requires a concurrent call to `result.Wait()` to ensure termination.
//
// Correct Usage Example:
//
//	done := make(chan struct{})
//	result, _ := process.New().Start("sleep", "2")
//
//	go func() {
//	    defer close(done)
//	    // This loop will continue as long as the process is running OR a zombie.
//	    for result.Running() {
//	        fmt.Println("Process is still running...")
//	        time.Sleep(500 * time.Millisecond)
//	    }
//	}()
//
//	// This call is mandatory. It blocks, reaps the zombie, and allows the goroutine to exit.
//	result.Wait()
//	<-done // Wait for monitoring goroutine to finish.
func (r *Running) Running() bool {
	if r.cmd == nil || r.cmd.Process == nil {
		return false
	}

	// Actively probe the OS by sending a null signal.
	// A nil error means the OS found the process (running or zombie).
	err := r.cmd.Process.Signal(unix.Signal(0))
	return err == nil
}

func (r *Running) Kill() error {
	return r.Signal(unix.SIGKILL)
}

func (r *Running) Signal(sig os.Signal) error {
	if r.cmd == nil || r.cmd.Process == nil {
		return errors.New("process not started")
	}
	return r.cmd.Process.Signal(sig)
}

// Stop gracefully stops the process by sending SIGTERM and waiting for a timeout.
func (r *Running) Stop(timeout time.Duration, sig ...os.Signal) error {
	if !r.Running() {
		return nil
	}

	var signalToSend os.Signal = unix.SIGTERM
	if len(sig) > 0 {
		signalToSend = sig[0]
	}

	if err := r.Signal(signalToSend); err != nil {
		if errors.Is(err, os.ErrProcessDone) {
			return nil
		}
		return err
	}

	done := make(chan struct{})
	go func() {
		r.Wait()
		close(done)
	}()

	select {
	case <-time.After(timeout):
		return r.Kill()
	case <-done:
		return nil
	}
}
