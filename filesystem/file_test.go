package filesystem

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/errors"
	mocksconfig "github.com/goravel/framework/mocks/config"
	mocksfilesystem "github.com/goravel/framework/mocks/filesystem"
	"github.com/goravel/framework/support/file"
)

type FileTestSuite struct {
	suite.Suite
	file       *File
	mockConfig *mocksconfig.Config
}

func TestFileTestSuite(t *testing.T) {
	suite.Run(t, new(FileTestSuite))

	assert.Nil(t, file.Remove("test.txt"))
}

func (s *FileTestSuite) SetupTest() {
	s.mockConfig = &mocksconfig.Config{}
	s.mockConfig.On("GetString", "filesystems.default").Return("local").Once()
	ConfigFacade = s.mockConfig

	f, err := NewFile("./file.go")
	s.Nil(err)
	s.NotNil(f)

	s.file = f
}

func (s *FileTestSuite) TestNewFile_Error() {
	f, err := NewFile("./file1.go")
	s.EqualError(err, "file doesn't exist")
	s.Nil(f)
}

func (s *FileTestSuite) TestGetClientOriginalName() {
	s.Equal("file.go", s.file.GetClientOriginalName())
}

func (s *FileTestSuite) TestGetClientOriginalExtension() {
	s.Equal("go", s.file.GetClientOriginalExtension())
}

func (s *FileTestSuite) TestHashName() {
	s.Len(s.file.HashName("goravel"), 52)
}

func (s *FileTestSuite) TestExtension() {
	extension, err := s.file.Extension()
	s.Equal("txt", extension)
	s.Nil(err)
}

func TestNewFileFromRequest(t *testing.T) {
	mockConfig := &mocksconfig.Config{}
	ConfigFacade = mockConfig
	mockConfig.On("GetString", "app.name").Return("goravel").Once()
	mockConfig.On("GetString", "filesystems.default").Return("local").Once()

	buf := new(bytes.Buffer)
	mw := multipart.NewWriter(buf)
	w, err := mw.CreateFormFile("file", "test.txt")
	assert.NotNil(t, w)
	if assert.NoError(t, err) {
		_, err = w.Write([]byte("test"))
		assert.NoError(t, err)
	}
	assert.Nil(t, mw.Close())
	req, err := http.NewRequest("POST", "/", buf)
	assert.NoError(t, err)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	err = req.ParseMultipartForm(10 << 20) // 10 MB
	assert.NoError(t, err)
	_, fileHeader, err := req.FormFile("file")
	assert.NoError(t, err)
	requestFile, err := NewFileFromRequest(fileHeader)
	assert.NoError(t, err)
	assert.Equal(t, ".txt", filepath.Ext(requestFile.path))

	mockConfig.AssertExpectations(t)
}

func TestNewFile_ConfigFacadeNotSet(t *testing.T) {
	originConfigFacade := ConfigFacade
	t.Cleanup(func() {
		ConfigFacade = originConfigFacade
	})
	ConfigFacade = nil

	f, err := NewFile("./file.go")

	assert.Nil(t, f)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), errors.ConfigFacadeNotSet.Error())
}

func TestFileStore(t *testing.T) {
	storage := mocksfilesystem.NewStorage(t)
	driver := mocksfilesystem.NewDriver(t)
	file := &File{
		storage: storage,
		disk:    "s3",
		path:    "./file.go",
		name:    "file.go",
	}

	storage.EXPECT().Disk("s3").Return(driver).Once()
	driver.EXPECT().PutFile("uploads", file).Return("uploads/hash.go", nil).Once()

	path, err := file.Store("uploads")
	assert.NoError(t, err)
	assert.Equal(t, "uploads/hash.go", path)
}

func TestFileStoreAs(t *testing.T) {
	storage := mocksfilesystem.NewStorage(t)
	driver := mocksfilesystem.NewDriver(t)
	file := &File{
		storage: storage,
		disk:    "s3",
		path:    "./file.go",
		name:    "file.go",
	}

	storage.EXPECT().Disk("s3").Return(driver).Once()
	driver.EXPECT().PutFileAs("uploads", file, "goravel.go").Return("uploads/goravel.go", nil).Once()

	path, err := file.StoreAs("uploads", "goravel.go")
	assert.NoError(t, err)
	assert.Equal(t, "uploads/goravel.go", path)
}

func TestFileStore_ErrorWhenStorageFacadeMissing(t *testing.T) {
	file := &File{
		path: "./file.go",
	}

	path, err := file.Store("uploads")
	assert.Empty(t, path)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), errors.StorageFacadeNotSet.Error())

	path, err = file.StoreAs("uploads", "goravel.go")
	assert.Empty(t, path)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), errors.StorageFacadeNotSet.Error())
}

func TestFileMetadataAndDisk(t *testing.T) {
	mockConfig := mocksconfig.NewConfig(t)
	tempFile := filepath.Join(t.TempDir(), "goravel.txt")
	assert.NoError(t, os.WriteFile(tempFile, []byte("framework"), 0o644))

	testFile := &File{
		config: mockConfig,
		path:   tempFile,
		name:   "goravel.txt",
		disk:   "local",
	}

	assert.Same(t, testFile, testFile.Disk("s3"))
	assert.Equal(t, "s3", testFile.disk)
	assert.Equal(t, tempFile, testFile.File())

	mockConfig.EXPECT().GetString("app.timezone").Return("UTC").Once()
	lastModified, err := testFile.LastModified()
	assert.NoError(t, err)
	assert.False(t, lastModified.IsZero())

	mimeType, err := testFile.MimeType()
	assert.NoError(t, err)
	assert.NotEmpty(t, mimeType)

	size, err := testFile.Size()
	assert.NoError(t, err)
	assert.Equal(t, int64(len("framework")), size)
}
