package transcription

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	frameworkai "github.com/goravel/framework/ai"
	contractsai "github.com/goravel/framework/contracts/ai"
	contractsfilesystem "github.com/goravel/framework/contracts/filesystem"
	"github.com/goravel/framework/foundation"
	mocksai "github.com/goravel/framework/mocks/ai"
	mocksfoundation "github.com/goravel/framework/mocks/foundation"
)

func TestOf(t *testing.T) {
	mockApp := mocksfoundation.NewApplication(t)
	mockAI := mocksai.NewAI(t)
	mockRequest := mocksai.NewTranscriptionRequest(t)
	attachment := frameworkai.DocumentFromByte([]byte("audio"))
	previousApp := foundation.App
	foundation.App = mockApp
	t.Cleanup(func() {
		foundation.App = previousApp
	})

	mockApp.EXPECT().MakeAI().Return(mockAI).Once()
	mockAI.EXPECT().Transcription(attachment).Return(mockRequest).Once()

	request := Of(attachment)

	assert.Same(t, mockRequest, request)
	assert.Implements(t, (*contractsai.TranscriptionRequest)(nil), request)
}

func TestFromPath(t *testing.T) {
	mockApp := mocksfoundation.NewApplication(t)
	mockAI := mocksai.NewAI(t)
	mockRequest := mocksai.NewTranscriptionRequest(t)
	tempFile, err := os.CreateTemp(t.TempDir(), "audio-*.mp3")
	require.NoError(t, err)
	_, err = tempFile.Write([]byte("audio"))
	require.NoError(t, err)
	require.NoError(t, tempFile.Close())
	previousApp := foundation.App
	foundation.App = mockApp
	t.Cleanup(func() {
		foundation.App = previousApp
	})

	mockApp.EXPECT().MakeAI().Return(mockAI).Once()
	mockAI.EXPECT().Transcription(mock.MatchedBy(func(file contractsai.StorableFile) bool {
		return file != nil
	})).RunAndReturn(func(file contractsai.StorableFile) contractsai.TranscriptionRequest {
		content, err := file.Content(context.Background())
		require.NoError(t, err)
		assert.Equal(t, []byte("audio"), content)
		assert.Equal(t, filepath.Base(tempFile.Name()), file.FileName())
		assert.Equal(t, "text/plain; charset=utf-8", file.MimeType())
		return mockRequest
	}).Once()

	request := FromPath(tempFile.Name())

	assert.Same(t, mockRequest, request)
}

func TestFromStorage(t *testing.T) {
	mockApp := mocksfoundation.NewApplication(t)
	mockAI := mocksai.NewAI(t)
	mockRequest := mocksai.NewTranscriptionRequest(t)
	previousApp := foundation.App
	foundation.App = mockApp
	t.Cleanup(func() {
		foundation.App = previousApp
	})

	mockApp.EXPECT().MakeAI().Return(mockAI).Once()
	mockAI.EXPECT().Transcription(mock.MatchedBy(func(file contractsai.StorableFile) bool {
		return file != nil
	})).RunAndReturn(func(file contractsai.StorableFile) contractsai.TranscriptionRequest {
		assert.NotNil(t, file)
		return mockRequest
	}).Once()

	request := FromStorage("call.mp3", WithDisk("audio"))

	assert.Same(t, mockRequest, request)
}

func TestFromUpload(t *testing.T) {
	mockApp := mocksfoundation.NewApplication(t)
	mockAI := mocksai.NewAI(t)
	mockRequest := mocksai.NewTranscriptionRequest(t)
	tempDir := t.TempDir()
	audioPath := tempDir + "/upload.mp3"
	require.NoError(t, os.WriteFile(audioPath, []byte("audio"), 0o644))
	upload := &transcriptionTestFile{path: audioPath, originalName: "upload.mp3", mimeType: "audio/mpeg"}
	previousApp := foundation.App
	foundation.App = mockApp
	t.Cleanup(func() {
		foundation.App = previousApp
	})

	mockApp.EXPECT().MakeAI().Return(mockAI).Once()
	mockAI.EXPECT().Transcription(mock.MatchedBy(func(file contractsai.StorableFile) bool {
		return file != nil
	})).RunAndReturn(func(file contractsai.StorableFile) contractsai.TranscriptionRequest {
		content, err := file.Content(context.Background())
		require.NoError(t, err)
		assert.Equal(t, []byte("audio"), content)
		assert.Equal(t, "upload.mp3", file.FileName())
		assert.Equal(t, "audio/mpeg", file.MimeType())
		return mockRequest
	}).Once()

	request := FromUpload(upload)

	assert.Same(t, mockRequest, request)
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
