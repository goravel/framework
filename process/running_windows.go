//go:build windows

package process

import (
	"errors"
	"os"
	"time"
)

// Running actively queries the OS to see if the process still exists.
//
// NOTE: This method returns `true` for a terminated process that has not yet
// been cleaned up by a call to `Wait()`. Due to this, a monitoring loop
// requires a concurrent call to `result.Wait()` to ensure termination.
//
// Correct Usage Example:
//
//		done := make(chan struct{})
//		result, _ := process.New().Start("timeout", "2")
//
//		go func() {
//	   	defer close(done)
//	   	// This loop will continue as long as the process object exists.
//	   	for result.Running() {
//	   		fmt.Println("Process is still running...")
//	   		time.Sleep(500 * time.Millisecond)
//	   	}
//		}()
//
//		// This call is mandatory. It blocks, cleans up the process, and allows the goroutine to exit.
//		result.Wait()
//		<-done // Wait for monitoring goroutine to finish.
func (r *Running) Running() bool {
	if r.cmd == nil || r.cmd.Process == nil {
		return false
	}

	// Actively probe the OS by trying to find the process by its PID.
	// A nil error means the OS found the process.
	_, err := os.FindProcess(r.cmd.Process.Pid)
	return err == nil
}

func (r *Running) Kill() error {
	if r.cmd == nil || r.cmd.Process == nil {
		return errors.New("process not started")
	}
	return r.cmd.Process.Kill()
}

// Signal sends a signal; on Windows, only os.Interrupt and os.Kill are supported.
func (r *Running) Signal(sig os.Signal) error {
	if r.cmd == nil || r.cmd.Process == nil {
		return errors.New("process not started")
	}
	return r.cmd.Process.Signal(sig)
}

// Stop terminates the process; on Windows, this is an alias for Kill().
func (r *Running) Stop(timeout time.Duration, sig ...os.Signal) error {
	if r.cmd == nil || r.cmd.Process == nil || !r.Running() {
		return nil
	}
	return r.Kill()
}
