package ai_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	frameworkai "github.com/goravel/framework/ai"
	"github.com/goravel/framework/ai/document"
	"github.com/goravel/framework/ai/image"
	contractsai "github.com/goravel/framework/contracts/ai"
	mocksfilesystem "github.com/goravel/framework/mocks/filesystem"
)

func TestAttachmentConstructors(t *testing.T) {
	image := image.New([]byte("png"), frameworkai.WithFilename("avatar.png"), frameworkai.WithMimeType("image/png"))

	assert.Equal(t, contractsai.AttachmentKindImage, image.Kind())
	assert.Equal(t, "avatar.png", image.Filename())
	assert.Equal(t, "image/png", image.MimeType())
	content, err := image.Content(context.Background())
	require.NoError(t, err)
	assert.Equal(t, []byte("png"), content)
}

func TestAttachmentReaderBuffersContentOnce(t *testing.T) {
	reader := bytes.NewBufferString("document")
	file := document.FromReader(reader, frameworkai.WithFilename("report.txt"))

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
	ctx := context.Background()
	storage := mocksfilesystem.NewStorage(t)
	storage.EXPECT().WithContext(ctx).Return(storage).Once()
	storage.EXPECT().GetBytes("docs/report.txt").Return([]byte("report"), nil).Once()
	storage.EXPECT().MimeType("docs/report.txt").Return("text/plain", nil).Once()

	file := document.FromStorage(storage, "docs/report.txt")
	first, err := file.Content(ctx)
	require.NoError(t, err)
	second, err := file.Content(ctx)
	require.NoError(t, err)

	assert.Equal(t, []byte("report"), first)
	assert.Equal(t, first, second)
	assert.Equal(t, "report.txt", file.Filename())
	assert.Equal(t, "text/plain", file.MimeType())
}
