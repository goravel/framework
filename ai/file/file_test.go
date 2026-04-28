package file_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/goravel/framework/ai/file"
	contractsai "github.com/goravel/framework/contracts/ai"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/foundation"
	mocksfilesystem "github.com/goravel/framework/mocks/filesystem"
	mocksfoundation "github.com/goravel/framework/mocks/foundation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestImageFromByte(t *testing.T) {
	attachment := file.ImageFromByte([]byte("png"), file.WithMimeType("image/png"))

	assert.Equal(t, contractsai.AttachmentKindImage, attachment.Kind())
	assert.Equal(t, "image/png", attachment.MimeType())
	content, err := attachment.Content(context.Background())
	require.NoError(t, err)
	assert.Equal(t, []byte("png"), content)
	assert.Equal(t, "", attachment.FileName())
}

func TestDocumentFromByteAndString_LeaveFileNameEmpty(t *testing.T) {
	attachment := file.DocumentFromByte([]byte("report"))
	content, err := attachment.Content(context.Background())
	require.NoError(t, err)
	assert.Equal(t, []byte("report"), content)
	assert.Equal(t, "", attachment.FileName())

	attachment = file.DocumentFromString("report")
	content, err = attachment.Content(context.Background())
	require.NoError(t, err)
	assert.Equal(t, []byte("report"), content)
	assert.Equal(t, "", attachment.FileName())
}

func TestDocumentFromStringAndImageFromBase64(t *testing.T) {
	attachment := file.DocumentFromString("report")
	content, err := attachment.Content(context.Background())
	require.NoError(t, err)
	assert.Equal(t, []byte("report"), content)

	attachment = file.ImageFromBase64("aW1hZ2U=", file.WithMimeType("image/png"))
	content, err = attachment.Content(context.Background())
	require.NoError(t, err)
	assert.Equal(t, []byte("image"), content)
	assert.Equal(t, "image/png", attachment.MimeType())
}

func TestDocumentFromReader_BuffersContentOnce(t *testing.T) {
	reader := bytes.NewBufferString("document")
	attachment := file.DocumentFromReader(reader)

	first, err := attachment.Content(context.Background())
	require.NoError(t, err)
	second, err := attachment.Content(context.Background())
	require.NoError(t, err)

	assert.Equal(t, []byte("document"), first)
	assert.Equal(t, first, second)
	assert.Equal(t, "", attachment.FileName())
	assert.Equal(t, "text/plain; charset=utf-8", attachment.MimeType())
}

func TestDocumentFromPath_UsesBasename(t *testing.T) {
	tempFile, err := os.CreateTemp(t.TempDir(), "report-*.txt")
	require.NoError(t, err)
	_, err = tempFile.WriteString("report")
	require.NoError(t, err)
	require.NoError(t, tempFile.Close())

	attachment := file.DocumentFromPath(tempFile.Name())
	content, err := attachment.Content(context.Background())
	require.NoError(t, err)

	assert.Equal(t, []byte("report"), content)
	assert.Equal(t, filepathBase(tempFile.Name()), attachment.FileName())
	assert.Equal(t, "text/plain; charset=utf-8", attachment.MimeType())
}

func TestImageFromPath_UsesBasename(t *testing.T) {
	tempFile, err := os.CreateTemp(t.TempDir(), "chart-*.png")
	require.NoError(t, err)
	_, err = tempFile.Write([]byte("image"))
	require.NoError(t, err)
	require.NoError(t, tempFile.Close())

	attachment := file.ImageFromPath(tempFile.Name())
	content, err := attachment.Content(context.Background())
	require.NoError(t, err)

	assert.Equal(t, []byte("image"), content)
	assert.Equal(t, filepathBase(tempFile.Name()), attachment.FileName())
}

func TestDocumentFromStorage_ResolvesOnce(t *testing.T) {
	originalApp := foundation.App
	t.Cleanup(func() {
		foundation.App = originalApp
	})

	ctx := context.Background()
	driver := mocksfilesystem.NewDriver(t)
	storage := mocksfilesystem.NewStorage(t)
	storage.EXPECT().Disk("docs").Return(driver).Once()
	driver.EXPECT().WithContext(ctx).Return(driver).Once()
	driver.EXPECT().GetBytes("report.txt").Return([]byte("report"), nil).Once()
	driver.EXPECT().MimeType("report.txt").Return("text/plain", nil).Once()

	app := mocksfoundation.NewApplication(t)
	app.EXPECT().MakeStorage().Return(storage).Once()
	foundation.App = app

	attachment := file.DocumentFromStorage("report.txt", file.WithDisk("docs"))
	first, err := attachment.Content(ctx)
	require.NoError(t, err)
	second, err := attachment.Content(ctx)
	require.NoError(t, err)

	assert.Equal(t, []byte("report"), first)
	assert.Equal(t, first, second)
	assert.Equal(t, "report.txt", attachment.FileName())
	assert.Equal(t, "text/plain", attachment.MimeType())
}

func TestDocumentFromStorage_UsesDefaultDisk(t *testing.T) {
	originalApp := foundation.App
	t.Cleanup(func() {
		foundation.App = originalApp
	})

	ctx := context.Background()
	storage := mocksfilesystem.NewStorage(t)
	storage.EXPECT().WithContext(ctx).Return(storage).Once()
	storage.EXPECT().GetBytes("report.txt").Return([]byte("report"), nil).Once()
	storage.EXPECT().MimeType("report.txt").Return("text/plain", nil).Once()

	app := mocksfoundation.NewApplication(t)
	app.EXPECT().MakeStorage().Return(storage).Once()
	foundation.App = app

	attachment := file.DocumentFromStorage("report.txt")
	content, err := attachment.Content(ctx)
	require.NoError(t, err)
	assert.Equal(t, []byte("report"), content)
	assert.Equal(t, "report.txt", attachment.FileName())
	assert.Equal(t, "text/plain", attachment.MimeType())
}

func TestDocumentFromUpload_ResolvesOnce(t *testing.T) {
	ctx := context.Background()
	tempFile, err := os.CreateTemp(t.TempDir(), "report-*.txt")
	require.NoError(t, err)
	_, err = tempFile.WriteString("report")
	require.NoError(t, err)
	require.NoError(t, tempFile.Close())

	upload := mocksfilesystem.NewFile(t)
	upload.EXPECT().File().Return(tempFile.Name()).Once()
	upload.EXPECT().MimeType().Return("text/plain", nil).Once()
	upload.EXPECT().GetClientOriginalName().Return("report.txt").Once()

	attachment := file.DocumentFromUpload(upload)
	content, err := attachment.Content(ctx)
	require.NoError(t, err)
	assert.Equal(t, []byte("report"), content)
	assert.Equal(t, "report.txt", attachment.FileName())
	assert.Equal(t, "text/plain", attachment.MimeType())
}

func TestDocumentFromURL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		assert.Equal(t, "/files/report.txt", request.URL.Path)
		writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, err := io.WriteString(writer, "report")
		require.NoError(t, err)
	}))
	defer server.Close()

	attachment := file.DocumentFromURL(server.URL + "/files/report.txt")
	content, err := attachment.Content(context.Background())
	require.NoError(t, err)
	assert.Equal(t, []byte("report"), content)
	assert.Equal(t, "report.txt", attachment.FileName())
	assert.Equal(t, "text/plain", attachment.MimeType())
}

func TestDocumentFromURL_WithoutPathLeavesFileNameEmpty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "application/octet-stream")
		_, err := io.WriteString(writer, "data")
		require.NoError(t, err)
	}))
	defer server.Close()

	attachment := file.DocumentFromURL(server.URL)
	content, err := attachment.Content(context.Background())
	require.NoError(t, err)
	assert.Equal(t, []byte("data"), content)
	assert.Equal(t, "", attachment.FileName())
	assert.Equal(t, "application/octet-stream", attachment.MimeType())
}

func TestDocumentFromURL_UsesDetectedMimeTypeWhenHeaderMissing(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		_, err := writer.Write([]byte("plain text"))
		require.NoError(t, err)
	}))
	defer server.Close()

	attachment := file.DocumentFromURL(server.URL + "/report.txt")
	content, err := attachment.Content(context.Background())
	require.NoError(t, err)
	assert.Equal(t, []byte("plain text"), content)
	assert.Equal(t, "report.txt", attachment.FileName())
	assert.Equal(t, "text/plain", attachment.MimeType())
}

func TestDocumentFromURL_ReturnsErrorWhenResponseTooLarge(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Content-Type", "text/plain")
		writer.Header().Set("Content-Length", "20971521")
		writer.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	attachment := file.DocumentFromURL(server.URL + "/report.txt")
	content, err := attachment.Content(context.Background())

	assert.Nil(t, content)
	assert.Equal(t, errors.AIAttachmentUrlResponseTooLarge.Args(20<<20), err)
}

func filepathBase(path string) string {
	index := bytes.LastIndexByte([]byte(path), os.PathSeparator)
	if index == -1 {
		return path
	}

	return path[index+1:]
}
