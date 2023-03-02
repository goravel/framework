package filesystem

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"path"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/goravel/framework/testing/mock"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

var testFile *File

type FileTestSuite struct {
	suite.Suite
}

func TestFileTestSuite(t *testing.T) {
	mockConfig := mock.Config()
	mockConfig.On("GetString", "filesystems.default").Return("local").Once()

	var err error
	testFile, err = NewFile("./file.go")
	assert.Nil(t, err)
	assert.NotNil(t, testFile)

	suite.Run(t, new(FileTestSuite))
	mockConfig.AssertExpectations(t)
}

func (s *FileTestSuite) SetupTest() {

}

func (s *FileTestSuite) TestNewFile_Error() {
	file, err := NewFile("./file1.go")
	s.EqualError(err, "file doesn't exist")
	s.Nil(file)
}

func (s *FileTestSuite) TestGetClientOriginalName() {
	s.Equal("file.go", testFile.GetClientOriginalName())
}

func (s *FileTestSuite) TestGetClientOriginalExtension() {
	s.Equal("go", testFile.GetClientOriginalExtension())
}

func (s *FileTestSuite) TestHashName() {
	s.Len(testFile.HashName("goravel"), 51)
}

func (s *FileTestSuite) TestExtension() {
	extension, err := testFile.Extension()
	s.Empty(extension)
	s.EqualError(err, "unknown file extension")
}

func TestNewFileFromRequest(t *testing.T) {
	mockConfig := mock.Config()
	mockConfig.On("GetString", "app.name").Return("goravel").Once()
	mockConfig.On("GetString", "filesystems.default").Return("local").Once()

	buf := new(bytes.Buffer)
	mw := multipart.NewWriter(buf)
	w, err := mw.CreateFormFile("file", "test.txt")
	if assert.NoError(t, err) {
		_, err = w.Write([]byte("test"))
		assert.NoError(t, err)
	}
	mw.Close()
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request, _ = http.NewRequest("POST", "/", buf)
	c.Request.Header.Set("Content-Type", mw.FormDataContentType())
	f, err := c.FormFile("file")
	assert.Nil(t, err)
	file, err := NewFileFromRequest(f)
	assert.Nil(t, err)
	assert.Equal(t, ".txt", path.Ext(file.file))

	mockConfig.AssertExpectations(t)
}
