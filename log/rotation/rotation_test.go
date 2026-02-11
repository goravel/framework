package rotation

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewStrftime(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		wantErr bool
	}{
		{
			name:    "valid pattern",
			pattern: "/var/log/app-%Y-%m-%d.log",
			wantErr: false,
		},
		{
			name:    "empty pattern",
			pattern: "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sf, err := NewStrftime(tt.pattern)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, sf)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, sf)
			}
		})
	}
}

func TestStrftime_Format(t *testing.T) {
	sf, err := NewStrftime("/var/log/app-%Y-%m-%d.log")
	assert.NoError(t, err)

	testTime := time.Date(2024, 3, 15, 10, 30, 45, 0, time.UTC)
	result := sf.Format(testTime)

	assert.Equal(t, "/var/log/app-2024-03-15.log", result)
}

func TestRotateLogs_New(t *testing.T) {
	rl, err := New("/tmp/test-%Y-%m-%d.log")
	assert.NoError(t, err)
	assert.NotNil(t, rl)
	assert.Equal(t, 24*time.Hour, rl.rotationTime)
}

func TestRotateLogs_WithOptions(t *testing.T) {
	mockClock := clockFunc(func() time.Time {
		return time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)
	})

	rl, err := New(
		"/tmp/test-%Y-%m-%d.log",
		WithRotationTime(1*time.Hour),
		WithRotationCount(5),
		WithClock(mockClock),
	)
	assert.NoError(t, err)
	assert.NotNil(t, rl)
	assert.Equal(t, 1*time.Hour, rl.rotationTime)
	assert.Equal(t, uint(5), rl.rotationCount)
	// Verify clock works by calling Now()
	assert.Equal(t, time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC), rl.clock.Now())
}

func TestRotateLogs_Write(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test-%Y-%m-%d.log")

	mockClock := clockFunc(func() time.Time {
		return time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)
	})

	rl, err := New(logPath, WithClock(mockClock))
	assert.NoError(t, err)
	defer rl.Close()

	// Write some data
	n, err := rl.Write([]byte("test log entry\n"))
	assert.NoError(t, err)
	assert.Equal(t, 15, n)

	// Check that the file was created
	expectedPath := filepath.Join(tmpDir, "test-2024-03-15.log")
	_, err = os.Stat(expectedPath)
	assert.NoError(t, err)

	// Read and verify content
	content, err := os.ReadFile(expectedPath)
	assert.NoError(t, err)
	assert.Equal(t, "test log entry\n", string(content))
}

func TestRotateLogs_Rotation(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test-%Y-%m-%d.log")

	// Start with a fixed time
	currentTime := time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)
	mockClock := &mutableClock{current: currentTime}

	rl, err := New(logPath, WithClock(mockClock), WithRotationTime(24*time.Hour))
	assert.NoError(t, err)
	defer rl.Close()

	// Write to first day
	_, err = rl.Write([]byte("day 1\n"))
	assert.NoError(t, err)

	// Advance time to next day
	mockClock.current = currentTime.Add(24 * time.Hour)

	// Write to second day
	_, err = rl.Write([]byte("day 2\n"))
	assert.NoError(t, err)

	// Check that both files exist
	file1 := filepath.Join(tmpDir, "test-2024-03-15.log")
	file2 := filepath.Join(tmpDir, "test-2024-03-16.log")

	_, err = os.Stat(file1)
	assert.NoError(t, err, "First day's log file should exist")

	_, err = os.Stat(file2)
	assert.NoError(t, err, "Second day's log file should exist")

	// Verify contents
	content1, _ := os.ReadFile(file1)
	assert.Equal(t, "day 1\n", string(content1))

	content2, _ := os.ReadFile(file2)
	assert.Equal(t, "day 2\n", string(content2))
}

func TestRotateLogs_Cleanup(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test-%Y-%m-%d.log")

	currentTime := time.Date(2024, 3, 15, 0, 0, 0, 0, time.UTC)
	mockClock := &mutableClock{current: currentTime}

	rl, err := New(
		logPath,
		WithClock(mockClock),
		WithRotationTime(24*time.Hour),
		WithRotationCount(2), // Keep only 2 files
	)
	assert.NoError(t, err)
	defer rl.Close()

	// Create 4 log files over 4 days
	for i := 0; i < 4; i++ {
		_, err = rl.Write([]byte("log entry\n"))
		assert.NoError(t, err)
		mockClock.current = mockClock.current.Add(24 * time.Hour)
		time.Sleep(10 * time.Millisecond) // Give cleanup goroutine time to run
	}

	// Wait a bit for cleanup goroutine to finish
	time.Sleep(100 * time.Millisecond)

	// Check how many files exist
	entries, err := os.ReadDir(tmpDir)
	assert.NoError(t, err)

	// Should have at most 2 files (rotationCount)
	assert.LessOrEqual(t, len(entries), 2, "Should keep at most 2 files")
}

// mutableClock is a clock implementation for testing that allows changing the time
type mutableClock struct {
	current time.Time
}

func (c *mutableClock) Now() time.Time {
	return c.current
}
