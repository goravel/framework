package driver

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	sessioncontract "github.com/goravel/framework/contracts/session"
	"github.com/goravel/framework/facades"
	"github.com/goravel/framework/support/carbon"
	"github.com/goravel/framework/support/file"
)

// File implements the session.Driver interface using the filesystem.
type File struct {
	path    string       // Directory where session files are stored
	minutes int          // Session lifetime in minutes
	mu      sync.RWMutex // Protects access during file operations
}

func newFile(path string, minutes int) (*File, error) {
	if path == "" {
		return nil, fmt.Errorf("session file path cannot be empty")
	}

	return &File{
		path:    path,
		minutes: minutes,
	}, nil
}

// FileDriverFactory creates an instance of the file session driver using framework config.
// This function matches the 'func() (session.Driver, error)' signature for config 'via'.
func FileDriverFactory() (sessioncontract.Driver, error) {
	config := facades.Config()

	lifetime := config.GetInt("session.lifetime")
	filesPath := config.GetString("session.files")

	if filesPath == "" {
		return nil, fmt.Errorf("session.files path is not configured; required for file session driver") // Return error: driver cannot function without a path
	}

	instance, err := newFile(filesPath, lifetime)
	if err != nil {
		return nil, err
	}
	return instance, nil
}

func (f *File) Close() error {
	return nil
}

// Destroy removes a session file by its ID.
func (f *File) Destroy(id string) error {
	f.mu.Lock() // Exclusive lock for delete operation
	defer f.mu.Unlock()

	if f.path == "" {

		return fmt.Errorf("session path not configured") // Return error
	}

	filePath := f.getFilePath(id)

	err := file.Remove(filePath)
	// Log errors unless the file simply didn't exist
	if err != nil && !os.IsNotExist(err) {

		return fmt.Errorf("failed to destroy session file '%s': %w", id, err)
	}
	return nil
}

// Gc performs garbage collection, removing expired session files.
func (f *File) Gc(maxLifetime int) error {
	f.mu.Lock() // Exclusive lock for GC potentially deleting many files
	defer f.mu.Unlock()

	if f.path == "" {

		return fmt.Errorf("session path not configured") // Return error
	}
	if maxLifetime <= 0 {

		return fmt.Errorf("invalid maxLifetime for GC: %d", maxLifetime)
	}

	cutoffTime := carbon.Now(carbon.UTC).SubSeconds(maxLifetime)

	var filesRemoved int
	var errorsEncountered int

	err := filepath.Walk(f.path, func(path string, info os.FileInfo, walkErr error) error {
		if walkErr != nil {
			// Log errors during walk (e.g., permissions) and potentially stop

			errorsEncountered++
			// Decide whether to stop walking on error. Returning the error stops.
			// Returning nil allows it to potentially continue with other files/subdirs.
			return nil
		}

		// Skip the root directory itself and any subdirectories immediately under it.
		// Session files should be directly in f.path, not in subdirs.
		if info.IsDir() {
			if path == f.path {
				return nil // Allow entering the root directory
			}

			return filepath.SkipDir
		}

		//Convert ModTime to utc
		modTime := info.ModTime().UTC()

		// Check modification time against cutoff for actual files
		if modTime.Before(cutoffTime.StdTime()) {

			removeErr := os.Remove(path)
			if removeErr != nil && !os.IsNotExist(removeErr) {

				errorsEncountered++
				// Continue GC even if one file fails to delete
			} else if removeErr == nil {
				filesRemoved++
			}
		}
		return nil // Continue walking
	})

	// Return the error from filepath.Walk if it terminated due to an issue.
	if err != nil {

		return fmt.Errorf("session GC failed for path '%s': %w", f.path, err)
	}
	return nil
}

// Open initializes the session driver. No action needed for file driver.
func (f *File) Open(savePath string, sessionName string) error {
	return nil
}

// Read retrieves session data from a file by ID. Returns empty string if not found or expired.
func (f *File) Read(id string) (string, error) {
	f.mu.RLock() // Read lock is sufficient
	defer f.mu.RUnlock()

	if f.path == "" {

		// Consistent behavior: return "" and no error if path is missing.
		return "", nil
	}

	filePath := f.getFilePath(id)

	// 1. Check existence first (optimization)
	if !file.Exists(filePath) {
		return "", nil
	}

	// 2. Check if expired (based on modification time)
	if f.minutes > 0 {
		modified, err := file.LastModified(filePath, carbon.UTC)
		if err != nil {

			// Treat as unreadable if mod time fails
			return "", fmt.Errorf("failed to check session expiry for '%s': %w", id, err)
		}

		expiryTime := carbon.Now(carbon.UTC).SubMinutes(f.minutes)
		if modified.Before(expiryTime.StdTime()) {

			return "", nil
		}
	}

	// 3. Read the file content
	data, err := os.ReadFile(filePath)
	if err != nil {

		return "", fmt.Errorf("failed to read session data for '%s': %w", id, err)
	}

	return string(data), nil
}

// Write saves session data to a file by ID.
func (f *File) Write(id string, data string) error {
	f.mu.Lock() // Exclusive lock for write operation
	defer f.mu.Unlock()

	if f.path == "" {
		return fmt.Errorf("cannot write session file: session path is not configured")
	}

	filePath := f.getFilePath(id)

	err := file.PutContent(filePath, data)
	if err != nil {
		return fmt.Errorf("failed to write session data for '%s': %w", id, err)
	}
	return nil
}

func (f *File) getFilePath(id string) string {
	return filepath.Join(f.path, id)
}

// Ensure File implements the Driver interface at compile time.
var _ sessioncontract.Driver = (*File)(nil)
