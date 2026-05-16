package transcription

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	contractsfilesystem "github.com/goravel/framework/contracts/filesystem"
	"github.com/goravel/framework/errors"
)

func TestFromPath(t *testing.T) {
	tempFile, err := os.CreateTemp(t.TempDir(), "audio-*.mp3")
	require.NoError(t, err)
	_, err = tempFile.Write([]byte("audio"))
	require.NoError(t, err)
	require.NoError(t, tempFile.Close())
	attachment := FromPath(tempFile.Name())
	content, err := attachment.Content(context.Background())
	require.NoError(t, err)

	assert.Equal(t, []byte("audio"), content)
	assert.Equal(t, filepath.Base(tempFile.Name()), attachment.FileName())
	assert.Equal(t, "text/plain; charset=utf-8", attachment.MimeType())
}

func TestFromUpload(t *testing.T) {
	tempDir := t.TempDir()
	audioPath := tempDir + "/upload.mp3"
	require.NoError(t, os.WriteFile(audioPath, []byte("audio"), 0o644))
	upload := &transcriptionTestFile{path: audioPath, originalName: "upload.mp3", mimeType: "audio/mpeg"}
	attachment := FromUpload(upload)
	content, err := attachment.Content(context.Background())
	require.NoError(t, err)

	assert.Equal(t, []byte("audio"), content)
	assert.Equal(t, "upload.mp3", attachment.FileName())
	assert.Equal(t, "audio/mpeg", attachment.MimeType())
}

func TestFromStorage(t *testing.T) {
	attachment := FromStorage("audio.mp3")
	require.NotNil(t, attachment)

	// Without a configured storage facade the resolver must return the expected error.
	_, err := attachment.Content(context.Background())
	assert.ErrorIs(t, err, errors.StorageFacadeNotSet)
}

func TestFromStorageWithDisk(t *testing.T) {
	attachment := FromStorage("audio.mp3", WithDisk("media"))
	require.NotNil(t, attachment)

	// WithDisk is forwarded; content resolution still fails without a real storage backend.
	_, err := attachment.Content(context.Background())
	assert.ErrorIs(t, err, errors.StorageFacadeNotSet)
}

type transcriptionTestFile struct {
	path         string
	originalName string
	mimeType     string
}

func (f *transcriptionTestFile) File() string                           { return f.path }
func (f *transcriptionTestFile) GetClientOriginalName() string          { return f.originalName }
func (f *transcriptionTestFile) MimeType() (string, error)              { return f.mimeType, nil }
func (f *transcriptionTestFile) Extension() (string, error)             { return ".mp3", nil }
func (f *transcriptionTestFile) Store(string) (string, error)           { return "", nil }
func (f *transcriptionTestFile) StoreAs(string, string) (string, error) { return "", nil }
func (f *transcriptionTestFile) StorePublicly(string) (string, error)   { return "", nil }
func (f *transcriptionTestFile) StorePubliclyAs(string, string) (string, error) {
	return "", nil
}
func (f *transcriptionTestFile) Disk(string) contractsfilesystem.File { return f }
func (f *transcriptionTestFile) GetClientOriginalExtension() string   { return "mp3" }
func (f *transcriptionTestFile) HashName(...string) string            { return "hash.mp3" }
func (f *transcriptionTestFile) LastModified() (time.Time, error)     { return time.Time{}, nil }
func (f *transcriptionTestFile) Size() (int64, error)                 { return 5, nil }
