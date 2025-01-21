package file

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gabriel-vasile/mimetype"

	"github.com/goravel/framework/errors"
)

// Option represents an option for FilePutContents
type Option func(*fileOptions)

type fileOptions struct {
	mode   os.FileMode
	append bool
}

// WithMode sets the file mode for FilePutContents
func WithMode(mode os.FileMode) Option {
	return func(opts *fileOptions) {
		opts.mode = mode
	}
}

// WithAppend sets the append mode for FilePutContents
func WithAppend(append bool) Option {
	return func(opts *fileOptions) {
		opts.append = append
	}
}

func ClientOriginalExtension(file string) string {
	return strings.ReplaceAll(filepath.Ext(file), ".", "")
}

func Contain(file string, search string) bool {
	if Exists(file) {
		data, err := GetContent(file)
		if err != nil {
			return false
		}

		return strings.Contains(data, search)
	}

	return false
}

// Create a file with the given content
// Deprecated: Use PutContent instead
func Create(file string, content string) error {
	if err := os.MkdirAll(filepath.Dir(file), os.ModePerm); err != nil {
		return err
	}

	f, err := os.Create(file)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err = f.WriteString(content); err != nil {
		return err
	}

	return nil
}

func Exists(file string) bool {
	_, err := os.Stat(file)
	return err == nil
}

// Extension Supported types: https://github.com/gabriel-vasile/mimetype/blob/master/supported_mimes.md
func Extension(file string, originalWhenUnknown ...bool) (string, error) {
	mtype, err := mimetype.DetectFile(file)
	if err != nil {
		return "", err
	}

	if mtype.String() == "" {
		if len(originalWhenUnknown) > 0 {
			if originalWhenUnknown[0] {
				return ClientOriginalExtension(file), nil
			}
		}

		return "", errors.UnknownFileExtension
	}

	return strings.TrimPrefix(mtype.Extension(), "."), nil
}

func LastModified(file, timezone string) (time.Time, error) {
	fileInfo, err := os.Stat(file)
	if err != nil {
		return time.Time{}, err
	}

	l, err := time.LoadLocation(timezone)
	if err != nil {
		return time.Time{}, err
	}

	return fileInfo.ModTime().In(l), nil
}

func MimeType(file string) (string, error) {
	mtype, err := mimetype.DetectFile(file)
	if err != nil {
		return "", err
	}

	return mtype.String(), nil
}

func Remove(file string) error {
	_, err := os.Stat(file)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}

		return err
	}

	return os.RemoveAll(file)
}

func Size(file string) (int64, error) {
	fileInfo, err := os.Open(file)
	if err != nil {
		return 0, err
	}
	defer fileInfo.Close()

	fi, err := fileInfo.Stat()
	if err != nil {
		return 0, err
	}

	return fi.Size(), nil
}

func GetContent(file string) (string, error) {
	// Read the entire file
	data, err := os.ReadFile(file)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func PutContent(file string, content string, options ...Option) error {
	// Default options
	opts := &fileOptions{
		mode:   os.ModePerm,
		append: false,
	}

	// Apply options
	for _, option := range options {
		option(opts)
	}

	// Create the directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(file), opts.mode); err != nil {
		return err
	}

	// Open file with appropriate flags
	flag := os.O_CREATE | os.O_WRONLY
	if opts.append {
		flag |= os.O_APPEND
	} else {
		flag |= os.O_TRUNC
	}

	// Open the file
	f, err := os.OpenFile(file, flag, opts.mode)
	if err != nil {
		return err
	}
	defer f.Close()

	// Write the content
	if _, err = f.WriteString(content); err != nil {
		return err
	}

	return nil
}
