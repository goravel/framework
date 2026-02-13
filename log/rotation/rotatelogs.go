package rotation

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// Clock is the interface used to determine the current time
type Clock interface {
	Now() time.Time
}

// clockFunc is a function that implements Clock
type clockFunc func() time.Time

func (f clockFunc) Now() time.Time {
	return f()
}

// RotateLogs represents a log file that gets automatically rotated
type RotateLogs struct {
	pattern       *Strftime
	clock         Clock
	rotationTime  time.Duration
	rotationCount uint

	mutex         sync.Mutex
	currentFile   *os.File
	currentPath   string
	lastRotation  time.Time
	cleanupWg     sync.WaitGroup // For testing - wait for cleanup to complete
}

// Option is a functional option for RotateLogs
type Option func(*RotateLogs)

// WithRotationTime sets the rotation interval
func WithRotationTime(d time.Duration) Option {
	return func(rl *RotateLogs) {
		rl.rotationTime = d
	}
}

// WithRotationCount sets the number of files to keep
func WithRotationCount(count uint) Option {
	return func(rl *RotateLogs) {
		rl.rotationCount = count
	}
}

// WithClock sets a custom clock for testing
func WithClock(clock Clock) Option {
	return func(rl *RotateLogs) {
		rl.clock = clock
	}
}

// New creates a new RotateLogs instance
func New(pattern string, options ...Option) (*RotateLogs, error) {
	strftime, err := NewStrftime(pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid pattern: %w", err)
	}

	rl := &RotateLogs{
		pattern:      strftime,
		clock:        clockFunc(time.Now),
		rotationTime: 24 * time.Hour,
	}

	for _, opt := range options {
		opt(rl)
	}

	return rl, nil
}

// Write implements io.Writer interface
func (rl *RotateLogs) Write(p []byte) (n int, err error) {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	// Check if we need to rotate
	if err := rl.rotate(); err != nil {
		return 0, err
	}

	return rl.currentFile.Write(p)
}

// rotate performs the rotation if needed
func (rl *RotateLogs) rotate() error {
	now := rl.clock.Now()
	
	// Calculate the base time for the current rotation period
	baseTime := now.Truncate(rl.rotationTime)
	
	// Check if we need to rotate
	if rl.currentFile != nil && rl.lastRotation.Equal(baseTime) {
		return nil // No rotation needed
	}

	// Generate the new filename
	newPath := rl.pattern.Format(baseTime)
	
	// Close the current file if it exists
	if rl.currentFile != nil {
		if err := rl.currentFile.Close(); err != nil {
			return fmt.Errorf("failed to close current file: %w", err)
		}
		rl.currentFile = nil
	}

	// Create the directory if it doesn't exist
	dir := filepath.Dir(newPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Open the new file
	file, err := os.OpenFile(newPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	rl.currentFile = file
	rl.currentPath = newPath
	rl.lastRotation = baseTime

	// Clean up old files if rotation count is set
	if rl.rotationCount > 0 {
		rl.cleanupWg.Add(1)
		go func() {
			defer rl.cleanupWg.Done()
			rl.cleanup()
		}()
	}

	return nil
}

// cleanup removes old log files based on rotation count
func (rl *RotateLogs) cleanup() {
	// Generate glob pattern from the strftime pattern
	// For example, "storage/logs/daily-%Y-%m-%d.log" becomes "storage/logs/daily-*.log"
	globPattern := rl.generateGlobPattern()
	
	// Find all files matching the pattern
	matches, err := filepath.Glob(globPattern)
	if err != nil {
		return
	}

	// Filter to only include regular files (not symlinks or directories)
	var logFiles []string
	for _, path := range matches {
		fi, err := os.Stat(path)
		if err != nil {
			continue
		}
		
		// Skip if not a regular file
		if !fi.Mode().IsRegular() {
			continue
		}
		
		logFiles = append(logFiles, path)
	}

	// Sort files by modification time (oldest first)
	sort.Slice(logFiles, func(i, j int) bool {
		infoI, errI := os.Stat(logFiles[i])
		infoJ, errJ := os.Stat(logFiles[j])
		if errI != nil || errJ != nil {
			return false
		}
		return infoI.ModTime().Before(infoJ.ModTime())
	})

	// Keep only the most recent rotationCount files
	if uint(len(logFiles)) > rl.rotationCount {
		filesToDelete := logFiles[:len(logFiles)-int(rl.rotationCount)]
		for _, path := range filesToDelete {
			_ = os.Remove(path) // Ignore errors during cleanup - files may already be deleted
		}
	}
}

// generateGlobPattern converts a strftime pattern to a glob pattern
func (rl *RotateLogs) generateGlobPattern() string {
	// Get the pattern and replace all strftime specifiers with *
	pattern := rl.pattern.pattern
	
	// Replace common strftime patterns with wildcards
	replacements := []string{
		"%Y", "*",
		"%m", "*",
		"%d", "*",
		"%H", "*",
		"%M", "*",
		"%S", "*",
	}
	
	result := pattern
	for i := 0; i < len(replacements); i += 2 {
		result = strings.ReplaceAll(result, replacements[i], replacements[i+1])
	}
	
	return result
}

// Close closes the current log file
func (rl *RotateLogs) Close() error {
	// Wait for any pending cleanups to complete
	rl.cleanupWg.Wait()

	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	if rl.currentFile == nil {
		return nil
	}

	err := rl.currentFile.Close()
	rl.currentFile = nil
	return err
}

// Ensure RotateLogs implements io.WriteCloser
var _ io.WriteCloser = (*RotateLogs)(nil)
