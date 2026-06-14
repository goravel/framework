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
	mocksai "github.com/goravel/framework/mocks/ai"
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
	_, ok := any(attachment).(contractsai.Attachment)
	assert.True(t, ok)

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

func TestDocumentFromURLReturnsErrorWhenHttpFacadeNotSet(t *testing.T) {
	originalHTTPFacade := httpFacade
	t.Cleanup(func() {
		httpFacade = originalHTTPFacade
	})
	httpFacade = nil

	attachment := DocumentFromURL("https://example.com/files/report.txt")
	content, err := attachment.Content(context.Background())

	assert.Nil(t, content)
	assert.Equal(t, errors.HttpFacadeNotSet, err)
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

func TestResolved_Put(t *testing.T) {
	originalAIFacade := aiFacade
	t.Cleanup(func() {
		aiFacade = originalAIFacade
	})

	tests := []struct {
		name        string
		setup       func(t *testing.T, ctx context.Context, attachment contractsai.Attachment) contractsai.FileResponse
		expectError error
		expectID    string
	}{
		{
			name: "success",
			setup: func(t *testing.T, ctx context.Context, attachment contractsai.Attachment) contractsai.FileResponse {
				fileProvider := mocksai.NewFileProvider(t)
				response := mocksai.NewFileResponse(t)
				response.EXPECT().ID().Return("file-456").Once()
				response.EXPECT().MimeType().Return("").Once()
				response.EXPECT().Content(ctx).Return(nil, nil).Once()

				fileProvider.EXPECT().PutFile(ctx, attachment).Return(response, nil).Once()
				aiFacade = &Application{
					ctx: context.Background(),
					config: contractsai.Config{
						Default: "openai",
						Providers: map[string]contractsai.ProviderConfig{
							"openai": {Via: uploadTestProvider{fileProvider: fileProvider}},
						},
					},
					resolver: NewProviderResolver(contractsai.Config{
						Default: "openai",
						Providers: map[string]contractsai.ProviderConfig{
							"openai": {Via: uploadTestProvider{fileProvider: fileProvider}},
						},
					}),
				}

				return response
			},
			expectID: "file-456",
		},
		{
			name: "facade not set",
			setup: func(t *testing.T, _ context.Context, _ contractsai.Attachment) contractsai.FileResponse {
				aiFacade = nil
				return nil
			},
			expectError: errors.AIFacadeNotSet,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.WithValue(context.Background(), testCtxKey("resolved-put"), tt.name)
			attachment := DocumentFromString("report")

			response := tt.setup(t, ctx, attachment)

			stored, err := attachment.Put(ctx)
			assert.Equal(t, tt.expectError, err)
			if tt.expectError != nil {
				assert.Nil(t, stored)
				return
			}

			require.NotNil(t, stored)
			assert.Equal(t, response, stored)
			assert.Equal(t, tt.expectID, stored.ID())
			assert.Empty(t, stored.MimeType())
			content, err := stored.Content(ctx)
			require.NoError(t, err)
			assert.Nil(t, content)
		})
	}
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

func TestDocumentFromIDGetAndDelete(t *testing.T) {
	originalAIFacade := aiFacade
	t.Cleanup(func() {
		aiFacade = originalAIFacade
	})

	ctx := context.Background()
	fileProvider := mocksai.NewFileProvider(t)
	response := mocksai.NewFileResponse(t)
	response.EXPECT().ID().Return("file-123").Once()
	response.EXPECT().MimeType().Return("text/plain").Once()
	response.EXPECT().Content(ctx).Return([]byte("report"), nil).Once()

	fileProvider.EXPECT().GetFile(ctx, "file-123").Return(response, nil).Once()
	fileProvider.EXPECT().DeleteFile(ctx, "file-123").Return(nil).Once()

	config := contractsai.Config{
		Default: "openai",
		Providers: map[string]contractsai.ProviderConfig{
			"openai": {Via: uploadTestProvider{fileProvider: fileProvider}},
		},
	}
	aiFacade = &Application{ctx: context.Background(), config: config, resolver: NewProviderResolver(config)}

	attachment := DocumentFromID("file-123")
	assert.Equal(t, "file-123", attachment.ID())

	file, err := attachment.Get(ctx)
	require.NoError(t, err)
	assert.Equal(t, "file-123", file.ID())

	content, err := attachment.Content(ctx)
	require.NoError(t, err)
	assert.Equal(t, []byte("report"), content)
	assert.Equal(t, "text/plain", attachment.MimeType())

	assert.NoError(t, attachment.Delete(ctx))
}

func TestDocumentFromIDPutRequiresID(t *testing.T) {
	attachment := DocumentFromID("")

	stored, err := attachment.Put(context.Background())
	assert.Nil(t, stored)
	assert.Equal(t, errors.AIStoredFileIDEmpty, err)
}

func TestDocumentFromIDContentCachesResolvedFile(t *testing.T) {
	originalAIFacade := aiFacade
	t.Cleanup(func() {
		aiFacade = originalAIFacade
	})

	ctx := context.Background()
	fileProvider := mocksai.NewFileProvider(t)
	response := mocksai.NewFileResponse(t)
	response.EXPECT().MimeType().Return("text/plain").Twice()
	response.EXPECT().Content(ctx).Return([]byte("report"), nil).Once()

	fileProvider.EXPECT().GetFile(ctx, "file-123").Return(response, nil).Once()

	config := contractsai.Config{
		Default: "openai",
		Providers: map[string]contractsai.ProviderConfig{
			"openai": {Via: uploadTestProvider{fileProvider: fileProvider}},
		},
	}
	aiFacade = &Application{ctx: context.Background(), config: config, resolver: NewProviderResolver(config)}

	attachment := DocumentFromID("file-123")
	content, err := attachment.Content(ctx)
	require.NoError(t, err)
	assert.Equal(t, []byte("report"), content)
	assert.Equal(t, "text/plain", attachment.MimeType())

	file, err := attachment.Get(ctx)
	require.NoError(t, err)
	assert.Equal(t, response, file)
	assert.Equal(t, "text/plain", file.MimeType())
}

func TestDocumentFromIDReturnsErrorWhenFacadeNotSet(t *testing.T) {
	originalAIFacade := aiFacade
	t.Cleanup(func() {
		aiFacade = originalAIFacade
	})
	aiFacade = nil

	attachment := DocumentFromID("file-123")
	file, err := attachment.Get(context.Background())
	assert.Nil(t, file)
	assert.Equal(t, errors.AIFacadeNotSet, err)
	assert.Equal(t, errors.AIFacadeNotSet, attachment.Delete(context.Background()))
}

func TestDocumentFromPathOverridesTitle(t *testing.T) {
	tempFile, err := os.CreateTemp(t.TempDir(), "report-*.txt")
	require.NoError(t, err)
	_, err = tempFile.WriteString("report")
	require.NoError(t, err)
	require.NoError(t, tempFile.Close())

	attachment := DocumentFromPath(tempFile.Name(), WithTitle("custom-title"))
	content, err := attachment.Content(context.Background())
	require.NoError(t, err)
	assert.Equal(t, []byte("report"), content)
	assert.Equal(t, filepathBase(tempFile.Name()), attachment.FileName())
}

func TestDocumentFromStringWithTitle(t *testing.T) {
	attachment := DocumentFromString("report", WithTitle("report-2025-06-14.txt"), WithMimeType("text/plain"))

	assert.Equal(t, "report-2025-06-14.txt", attachment.FileName())
	assert.Equal(t, "text/plain", attachment.MimeType())
	content, err := attachment.Content(context.Background())
	require.NoError(t, err)
	assert.Equal(t, []byte("report"), content)
	assert.Equal(t, "report-2025-06-14.txt", attachment.FileName())
}

func filepathBase(path string) string {
	index := bytes.LastIndexByte([]byte(path), os.PathSeparator)
	if index == -1 {
		return path
	}

	return path[index+1:]
}
