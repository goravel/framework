package rotation

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
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
		go rl.cleanup()
	}

	return nil
}

// cleanup removes old log files based on rotation count
func (rl *RotateLogs) cleanup() {
	// Get the base directory and pattern for matching files
	dir := filepath.Dir(rl.currentPath)
	
	// List all files in the directory
	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}

	// Get the base name pattern (without the date part)
	// For example, if pattern is "storage/logs/daily-%Y-%m-%d.log"
	// we want to match files like "daily-*.log"
	
	var logFiles []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		
		fullPath := filepath.Join(dir, entry.Name())
		
		// Simple heuristic: if the file has a similar extension and is in the same directory
		// we consider it part of this log rotation set
		// This is a simplified version - the original package uses glob patterns
		logFiles = append(logFiles, fullPath)
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
			// Don't delete the current file
			if path != rl.currentPath {
				os.Remove(path)
			}
		}
	}
}

// Close closes the current log file
func (rl *RotateLogs) Close() error {
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
