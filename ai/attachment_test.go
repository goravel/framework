package ai_test

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/goravel/framework/foundation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/goravel/framework/ai/document"
	"github.com/goravel/framework/ai/image"
	contractsai "github.com/goravel/framework/contracts/ai"
	mocksfilesystem "github.com/goravel/framework/mocks/filesystem"
	mocksfilesystemdriver "github.com/goravel/framework/mocks/filesystem"
	mocksfoundation "github.com/goravel/framework/mocks/foundation"
)

func TestAttachmentConstructors(t *testing.T) {
	image := image.FromByte([]byte("png"), image.WithFilename("avatar.png"), image.WithMimeType("image/png"))

	assert.Equal(t, contractsai.AttachmentKindImage, image.Kind())
	assert.Equal(t, "avatar.png", image.Filename())
	assert.Equal(t, "image/png", image.MimeType())
	content, err := image.Content(context.Background())
	require.NoError(t, err)
	assert.Equal(t, []byte("png"), content)
}

func TestAttachmentStringAndBase64Constructors(t *testing.T) {
	file := document.FromString("report", document.WithFilename("report.txt"))
	content, err := file.Content(context.Background())
	require.NoError(t, err)
	assert.Equal(t, []byte("report"), content)
	assert.Equal(t, "report.txt", file.Filename())

	image := image.FromBase64("aW1hZ2U=", image.WithMimeType("image/png"))
	content, err = image.Content(context.Background())
	require.NoError(t, err)
	assert.Equal(t, []byte("image"), content)
	assert.Equal(t, "image/png", image.MimeType())
}

func TestAttachmentReaderBuffersContentOnce(t *testing.T) {
	reader := bytes.NewBufferString("document")
	file := document.FromReader(reader, document.WithFilename("report.txt"))

	first, err := file.Content(context.Background())
	require.NoError(t, err)
	second, err := file.Content(context.Background())
	require.NoError(t, err)

	assert.Equal(t, []byte("document"), first)
	assert.Equal(t, first, second)
	assert.Equal(t, "report.txt", file.Filename())
	assert.Equal(t, "text/plain; charset=utf-8", file.MimeType())
}

func TestAttachmentFromStorageResolvesOnce(t *testing.T) {
	originalApp := foundation.App
	t.Cleanup(func() {
		foundation.App = originalApp
	})

	ctx := context.Background()
	driver := mocksfilesystemdriver.NewDriver(t)
	storage := mocksfilesystem.NewStorage(t)
	storage.EXPECT().Disk("docs").Return(driver).Once()
	driver.EXPECT().WithContext(ctx).Return(driver).Once()
	driver.EXPECT().GetBytes("report.txt").Return([]byte("report"), nil).Once()
	driver.EXPECT().MimeType("report.txt").Return("text/plain", nil).Once()

	app := mocksfoundation.NewApplication(t)
	app.EXPECT().MakeStorage().Return(storage).Once()
	foundation.App = app

	file := document.FromStorage("report.txt", "docs")
	first, err := file.Content(ctx)
	require.NoError(t, err)
	second, err := file.Content(ctx)
	require.NoError(t, err)

	assert.Equal(t, []byte("report"), first)
	assert.Equal(t, first, second)
	assert.Equal(t, "report.txt", file.Filename())
	assert.Equal(t, "text/plain", file.MimeType())
}

func TestAttachmentFromUploadResolvesOnce(t *testing.T) {
	ctx := context.Background()
	tempFile, err := os.CreateTemp(t.TempDir(), "report-*.txt")
	require.NoError(t, err)
	_, err = tempFile.WriteString("report")
	require.NoError(t, err)
	require.NoError(t, tempFile.Close())

	file := mocksfilesystem.NewFile(t)
	file.EXPECT().File().Return(tempFile.Name()).Once()
	file.EXPECT().MimeType().Return("text/plain", nil).Once()
	file.EXPECT().GetClientOriginalName().Return("report.txt").Once()

	upload := document.FromUpload(file)
	content, err := upload.Content(ctx)
	require.NoError(t, err)
	assert.Equal(t, []byte("report"), content)
	assert.Equal(t, "report.txt", upload.Filename())
	assert.Equal(t, "text/plain", upload.MimeType())
}

func TestAttachmentFromUrl(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		assert.Equal(t, "/files/report.txt", request.URL.Path)
		writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
		_, err := io.WriteString(writer, "report")
		require.NoError(t, err)
	}))
	defer server.Close()

	attachment := document.FromUrl(server.URL + "/files/report.txt")
	content, err := attachment.Content(context.Background())
	require.NoError(t, err)
	assert.Equal(t, []byte("report"), content)
	assert.Equal(t, "report.txt", attachment.Filename())
	assert.Equal(t, "text/plain", attachment.MimeType())
}
