package filesystem

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	configmock "github.com/goravel/framework/contracts/config/mocks"
	"github.com/goravel/framework/support/file"
)

type FileTestSuite struct {
	suite.Suite
	file       *File
	mockConfig *configmock.Config
}

func TestFileTestSuite(t *testing.T) {
	suite.Run(t, new(FileTestSuite))

	assert.Nil(t, file.Remove("test.txt"))
}

func (s *FileTestSuite) SetupTest() {
	s.mockConfig = &configmock.Config{}
	s.mockConfig.On("GetString", "filesystems.default").Return("local").Once()
	ConfigFacade = s.mockConfig

	var err error
	s.file, err = NewFile("./file.go")
	s.Nil(err)
	s.NotNil(s.file)
}

func (s *FileTestSuite) TestNewFile_Error() {
	file, err := NewFile("./file1.go")
	s.EqualError(err, "file doesn't exist")
	s.Nil(file)
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
	mockConfig := &configmock.Config{}
	ConfigFacade = mockConfig
	mockConfig.On("GetString", "app.name").Return("goravel").Once()
	mockConfig.On("GetString", "filesystems.default").Return("local").Once()

	buf := new(bytes.Buffer)
	mw := multipart.NewWriter(buf)
	w, err := mw.CreateFormFile("file", "test.txt")
	if assert.NoError(t, err) {
		_, err = w.Write([]byte("test"))
		assert.NoError(t, err)
	}
	assert.Nil(t, mw.Close())
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request, _ = http.NewRequest("POST", "/", buf)
	c.Request.Header.Set("Content-Type", mw.FormDataContentType())
	f, err := c.FormFile("file")
	assert.Nil(t, err)
	requestFile, err := NewFileFromRequest(f)
	assert.Nil(t, err)
	assert.Equal(t, ".txt", filepath.Ext(requestFile.path))

	mockConfig.AssertExpectations(t)
}
