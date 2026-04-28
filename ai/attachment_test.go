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

	aifile "github.com/goravel/framework/ai/file"
	contractsai "github.com/goravel/framework/contracts/ai"
	mocksfilesystem "github.com/goravel/framework/mocks/filesystem"
	mocksfoundation "github.com/goravel/framework/mocks/foundation"
)

func TestAttachmentConstructors(t *testing.T) {
	attachment := aifile.ImageFromByte([]byte("png"), aifile.WithFilename("avatar.png"), aifile.WithMimeType("image/png"))

	assert.Equal(t, contractsai.AttachmentKindImage, attachment.Kind())
	assert.Equal(t, "avatar.png", attachment.Filename())
	assert.Equal(t, "image/png", attachment.MimeType())
	content, err := attachment.Content(context.Background())
	require.NoError(t, err)
	assert.Equal(t, []byte("png"), content)
}

func TestAttachmentStringAndBase64Constructors(t *testing.T) {
	attachment := aifile.DocumentFromString("report", aifile.WithFilename("report.txt"))
	content, err := attachment.Content(context.Background())
	require.NoError(t, err)
	assert.Equal(t, []byte("report"), content)
	assert.Equal(t, "report.txt", attachment.Filename())

	attachment = aifile.ImageFromBase64("aW1hZ2U=", aifile.WithMimeType("image/png"))
	content, err = attachment.Content(context.Background())
	require.NoError(t, err)
	assert.Equal(t, []byte("image"), content)
	assert.Equal(t, "image/png", attachment.MimeType())
}

func TestAttachmentReaderBuffersContentOnce(t *testing.T) {
	reader := bytes.NewBufferString("document")
	attachment := aifile.DocumentFromReader(reader, aifile.WithFilename("report.txt"))

	first, err := attachment.Content(context.Background())
	require.NoError(t, err)
	second, err := attachment.Content(context.Background())
	require.NoError(t, err)

	assert.Equal(t, []byte("document"), first)
	assert.Equal(t, first, second)
	assert.Equal(t, "report.txt", attachment.Filename())
	assert.Equal(t, "text/plain; charset=utf-8", attachment.MimeType())
}

func TestAttachmentFromStorageResolvesOnce(t *testing.T) {
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

	attachment := aifile.DocumentFromStorage("report.txt", aifile.WithDisk("docs"))
	first, err := attachment.Content(ctx)
	require.NoError(t, err)
	second, err := attachment.Content(ctx)
	require.NoError(t, err)

	assert.Equal(t, []byte("report"), first)
	assert.Equal(t, first, second)
	assert.Equal(t, "report.txt", attachment.Filename())
	assert.Equal(t, "text/plain", attachment.MimeType())
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

	upload := aifile.DocumentFromUpload(file)
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

	attachment := aifile.DocumentFromURL(server.URL + "/files/report.txt")
	content, err := attachment.Content(context.Background())
	require.NoError(t, err)
	assert.Equal(t, []byte("report"), content)
	assert.Equal(t, "report.txt", attachment.Filename())
	assert.Equal(t, "text/plain", attachment.MimeType())
}
