package ai

import (
	"bytes"
	"context"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	contractsai "github.com/goravel/framework/contracts/ai"
	"github.com/goravel/framework/errors"
	mocksfilesystem "github.com/goravel/framework/mocks/filesystem"
	mockshttpclient "github.com/goravel/framework/mocks/http/client"
)

func TestImageFromByte(t *testing.T) {
	attachment := ImageFromByte([]byte("png"), WithMimeType("image/png"))

	assert.Equal(t, contractsai.AttachmentKindImage, attachment.Kind())
	assert.Equal(t, "image/png", attachment.MimeType())
	content, err := attachment.Content(context.Background())
	require.NoError(t, err)
	assert.Equal(t, []byte("png"), content)
	assert.Equal(t, "", attachment.FileName())
}

func TestDocumentFromByteAndStringLeaveFileNameEmpty(t *testing.T) {
	attachment := DocumentFromByte([]byte("report"))
	content, err := attachment.Content(context.Background())
	require.NoError(t, err)
	assert.Equal(t, []byte("report"), content)
	assert.Equal(t, "", attachment.FileName())

	attachment = DocumentFromString("report")
	content, err = attachment.Content(context.Background())
	require.NoError(t, err)
	assert.Equal(t, []byte("report"), content)
	assert.Equal(t, "", attachment.FileName())
}

func TestDocumentFromStringAndImageFromBase64(t *testing.T) {
	attachment := DocumentFromString("report")
	content, err := attachment.Content(context.Background())
	require.NoError(t, err)
	assert.Equal(t, []byte("report"), content)

	attachment = ImageFromBase64("aW1hZ2U=", WithMimeType("image/png"))
	content, err = attachment.Content(context.Background())
	require.NoError(t, err)
	assert.Equal(t, []byte("image"), content)
	assert.Equal(t, "image/png", attachment.MimeType())
}

func TestDocumentFromReaderBuffersContentOnce(t *testing.T) {
	reader := bytes.NewBufferString("document")
	attachment := DocumentFromReader(reader)

	first, err := attachment.Content(context.Background())
	require.NoError(t, err)
	second, err := attachment.Content(context.Background())
	require.NoError(t, err)

	assert.Equal(t, []byte("document"), first)
	assert.Equal(t, first, second)
	assert.Equal(t, "", attachment.FileName())
	assert.Equal(t, "text/plain; charset=utf-8", attachment.MimeType())
}

func TestDocumentFromPathUsesBasename(t *testing.T) {
	tempFile, err := os.CreateTemp(t.TempDir(), "report-*.txt")
	require.NoError(t, err)
	_, err = tempFile.WriteString("report")
	require.NoError(t, err)
	require.NoError(t, tempFile.Close())

	attachment := DocumentFromPath(tempFile.Name())
	content, err := attachment.Content(context.Background())
	require.NoError(t, err)

	assert.Equal(t, []byte("report"), content)
	assert.Equal(t, filepathBase(tempFile.Name()), attachment.FileName())
	assert.Equal(t, "text/plain; charset=utf-8", attachment.MimeType())
}

func TestImageFromPathUsesBasename(t *testing.T) {
	tempFile, err := os.CreateTemp(t.TempDir(), "chart-*.png")
	require.NoError(t, err)
	_, err = tempFile.Write([]byte("image"))
	require.NoError(t, err)
	require.NoError(t, tempFile.Close())

	attachment := ImageFromPath(tempFile.Name())
	content, err := attachment.Content(context.Background())
	require.NoError(t, err)

	assert.Equal(t, []byte("image"), content)
	assert.Equal(t, filepathBase(tempFile.Name()), attachment.FileName())
}

func TestDocumentFromStorageResolvesOnce(t *testing.T) {
	originalStorageFacade := storageFacade
	t.Cleanup(func() {
		storageFacade = originalStorageFacade
	})

	ctx := context.Background()
	driver := mocksfilesystem.NewDriver(t)
	storage := mocksfilesystem.NewStorage(t)
	storage.EXPECT().Disk("docs").Return(driver).Once()
	driver.EXPECT().WithContext(ctx).Return(driver).Once()
	driver.EXPECT().GetBytes("report.txt").Return([]byte("report"), nil).Once()
	driver.EXPECT().MimeType("report.txt").Return("text/plain", nil).Once()
	storageFacade = storage

	attachment := DocumentFromStorage("report.txt", WithDisk("docs"))
	first, err := attachment.Content(ctx)
	require.NoError(t, err)
	second, err := attachment.Content(ctx)
	require.NoError(t, err)

	assert.Equal(t, []byte("report"), first)
	assert.Equal(t, first, second)
	assert.Equal(t, "report.txt", attachment.FileName())
	assert.Equal(t, "text/plain", attachment.MimeType())
}

func TestDocumentFromStorageUsesDefaultDisk(t *testing.T) {
	originalStorageFacade := storageFacade
	t.Cleanup(func() {
		storageFacade = originalStorageFacade
	})

	ctx := context.Background()
	storage := mocksfilesystem.NewStorage(t)
	storage.EXPECT().WithContext(ctx).Return(storage).Once()
	storage.EXPECT().GetBytes("report.txt").Return([]byte("report"), nil).Once()
	storage.EXPECT().MimeType("report.txt").Return("text/plain", nil).Once()
	storageFacade = storage

	attachment := DocumentFromStorage("report.txt")
	content, err := attachment.Content(ctx)
	require.NoError(t, err)
	assert.Equal(t, []byte("report"), content)
	assert.Equal(t, "report.txt", attachment.FileName())
	assert.Equal(t, "text/plain", attachment.MimeType())
}

func TestDocumentFromUploadResolvesOnce(t *testing.T) {
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

	attachment := DocumentFromUpload(upload)
	content, err := attachment.Content(ctx)
	require.NoError(t, err)
	assert.Equal(t, []byte("report"), content)
	assert.Equal(t, "report.txt", attachment.FileName())
	assert.Equal(t, "text/plain", attachment.MimeType())
}

func TestDocumentFromURL(t *testing.T) {
	originalHTTPFacade := httpFacade
	t.Cleanup(func() {
		httpFacade = originalHTTPFacade
	})

	request := mockshttpclient.NewRequest(t)
	response := mockshttpclient.NewResponse(t)
	responseStream := io.NopCloser(bytes.NewBufferString("report"))
	ctx := context.Background()

	httpFactory := mockshttpclient.NewFactory(t)
	httpFactory.EXPECT().WithContext(ctx).Return(request).Once()
	request.EXPECT().Get("https://example.com/files/report.txt").Return(response, nil).Once()
	response.EXPECT().Successful().Return(true).Once()
	response.EXPECT().Header("Content-Length").Return("").Once()
	response.EXPECT().Stream().Return(responseStream, nil).Once()
	response.EXPECT().Header("Content-Type").Return("text/plain; charset=utf-8").Once()
	httpFacade = httpFactory

	attachment := DocumentFromURL("https://example.com/files/report.txt")
	content, err := attachment.Content(ctx)
	require.NoError(t, err)
	assert.Equal(t, []byte("report"), content)
	assert.Equal(t, "report.txt", attachment.FileName())
	assert.Equal(t, "text/plain", attachment.MimeType())
}

func TestDocumentFromURLWithoutPathLeavesFileNameEmpty(t *testing.T) {
	originalHTTPFacade := httpFacade
	t.Cleanup(func() {
		httpFacade = originalHTTPFacade
	})

	request := mockshttpclient.NewRequest(t)
	response := mockshttpclient.NewResponse(t)
	ctx := context.Background()

	httpFactory := mockshttpclient.NewFactory(t)
	httpFactory.EXPECT().WithContext(ctx).Return(request).Once()
	request.EXPECT().Get("https://example.com").Return(response, nil).Once()
	response.EXPECT().Successful().Return(true).Once()
	response.EXPECT().Header("Content-Length").Return("").Once()
	response.EXPECT().Stream().Return(io.NopCloser(bytes.NewBufferString("data")), nil).Once()
	response.EXPECT().Header("Content-Type").Return("application/octet-stream").Once()
	httpFacade = httpFactory

	attachment := DocumentFromURL("https://example.com")
	content, err := attachment.Content(ctx)
	require.NoError(t, err)
	assert.Equal(t, []byte("data"), content)
	assert.Equal(t, "", attachment.FileName())
	assert.Equal(t, "application/octet-stream", attachment.MimeType())
}

func TestDocumentFromURLUsesDetectedMimeTypeWhenHeaderMissing(t *testing.T) {
	originalHTTPFacade := httpFacade
	t.Cleanup(func() {
		httpFacade = originalHTTPFacade
	})

	request := mockshttpclient.NewRequest(t)
	response := mockshttpclient.NewResponse(t)
	ctx := context.Background()

	httpFactory := mockshttpclient.NewFactory(t)
	httpFactory.EXPECT().WithContext(ctx).Return(request).Once()
	request.EXPECT().Get("https://example.com/report.txt").Return(response, nil).Once()
	response.EXPECT().Successful().Return(true).Once()
	response.EXPECT().Header("Content-Length").Return("").Once()
	response.EXPECT().Stream().Return(io.NopCloser(bytes.NewBufferString("plain text")), nil).Once()
	response.EXPECT().Header("Content-Type").Return("").Once()
	httpFacade = httpFactory

	attachment := DocumentFromURL("https://example.com/report.txt")
	content, err := attachment.Content(ctx)
	require.NoError(t, err)
	assert.Equal(t, []byte("plain text"), content)
	assert.Equal(t, "report.txt", attachment.FileName())
	assert.Equal(t, "text/plain; charset=utf-8", attachment.MimeType())
}

func TestDocumentFromURLReturnsErrorWhenResponseTooLarge(t *testing.T) {
	originalHTTPFacade := httpFacade
	t.Cleanup(func() {
		httpFacade = originalHTTPFacade
	})

	request := mockshttpclient.NewRequest(t)
	response := mockshttpclient.NewResponse(t)
	ctx := context.Background()

	httpFactory := mockshttpclient.NewFactory(t)
	httpFactory.EXPECT().WithContext(ctx).Return(request).Once()
	request.EXPECT().Get("https://example.com/report.txt").Return(response, nil).Once()
	response.EXPECT().Successful().Return(true).Once()
	response.EXPECT().Header("Content-Length").Return("20971521").Once()
	httpFacade = httpFactory

	attachment := DocumentFromURL("https://example.com/report.txt")
	content, err := attachment.Content(ctx)

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
