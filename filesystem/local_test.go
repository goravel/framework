package filesystem

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	configmock "github.com/goravel/framework/contracts/config/mocks"
	"github.com/goravel/framework/support/file"
)

type LocalTestSuite struct {
	suite.Suite
	local      *Local
	file       *File
	mockConfig *configmock.Config
}

func TestLocalTestSuite(t *testing.T) {
	suite.Run(t, new(LocalTestSuite))

	assert.Nil(t, file.Remove("test.txt"))
}

func (s *LocalTestSuite) SetupTest() {
	s.mockConfig = &configmock.Config{}

	dir, err := os.MkdirTemp("", "local-test")
	s.Nil(err)

	err = os.WriteFile(dir+"/test.txt", []byte("goravel"), 0644)
	s.Nil(err)

	err = os.Mkdir(dir+"/test", 0755)
	s.Nil(err)

	s.mockConfig.On("GetString", "filesystems.default").Return("local").Once()
	s.mockConfig.On("GetString", "filesystems.disks.local.root").Return(dir).Once()
	s.mockConfig.On("GetString", "filesystems.disks.local.url").Return("https://goravel.dev").Once()
	ConfigFacade = s.mockConfig

	s.local, err = NewLocal(s.mockConfig, "local")
	s.Nil(err)
	s.NotNil(s.local)

	s.file, err = NewFile("./file.go")
	s.Nil(err)
	s.NotNil(s.file)

	s.mockConfig.AssertExpectations(s.T())
}

func (s *LocalTestSuite) TestAllDirectories() {
	directories, err := s.local.AllDirectories("")
	s.Nil(err)
	s.Len(directories, 1)
}

func (s *LocalTestSuite) TestAllFiles() {
	files, err := s.local.AllFiles("")
	s.Nil(err)
	s.Len(files, 1)
}

func (s *LocalTestSuite) TestCopy() {
	err := s.local.Copy("test.txt", "test1.txt")
	s.Nil(err)

	_, err = os.Stat(s.local.fullPath("test1.txt"))
	s.Nil(err)

	err = os.Remove(s.local.fullPath("test1.txt"))
	s.Nil(err)
}

func (s *LocalTestSuite) TestDelete() {
	err := s.local.Copy("test.txt", "test1.txt")
	s.Nil(err)

	err = s.local.Delete("test1.txt")
	s.Nil(err)

	_, err = os.Stat(s.local.fullPath("test1.txt"))
	s.True(os.IsNotExist(err))
}

func (s *LocalTestSuite) TestDirectories() {
	directories, err := s.local.Directories("")
	s.Nil(err)
	s.Len(directories, 1)
}

func (s *LocalTestSuite) TestDeleteDirectory() {
	err := s.local.DeleteDirectory("test")
	s.Nil(err)

	_, err = os.Stat(s.local.fullPath("test"))
	s.True(os.IsNotExist(err))
}

func (s *LocalTestSuite) TestExists() {
	exists := s.local.Exists("test.txt")
	s.True(exists)

	exists = s.local.Exists("test1.txt")
	s.False(exists)
}

func (s *LocalTestSuite) TestFiles() {
	files, err := s.local.Files("")
	s.Nil(err)
	s.Len(files, 1)
}

func (s *LocalTestSuite) TestGet() {
	content, err := s.local.Get("test.txt")
	s.Nil(err)
	s.Equal("goravel", content)
}

func (s *LocalTestSuite) TestGetBytes() {
	content, err := s.local.GetBytes("test.txt")
	s.Nil(err)
	s.Equal([]byte("goravel"), content)
}

func (s *LocalTestSuite) TestLastModified() {
	s.mockConfig.On("GetString", "app.timezone").Return("UTC").Once()

	lastModified, err := s.local.LastModified("test.txt")
	s.Nil(err)
	s.NotNil(lastModified)
}

func (s *LocalTestSuite) TestMakeDirectory() {
	err := s.local.MakeDirectory("test1")
	s.Nil(err)

	_, err = os.Stat(s.local.fullPath("test1"))
	s.Nil(err)

	err = os.Remove(s.local.fullPath("test1"))
	s.Nil(err)
}

func (s *LocalTestSuite) TestMimeType() {
	mimeType, err := s.local.MimeType("test.txt")
	s.Nil(err)
	s.Equal("text/plain; charset=utf-8", mimeType)
}

func (s *LocalTestSuite) TestMissing() {
	missing := s.local.Missing("test.txt")
	s.False(missing)

	missing = s.local.Missing("test1.txt")
	s.True(missing)
}

func (s *LocalTestSuite) TestMove() {
	err := s.local.Move("test.txt", "test1.txt")
	s.Nil(err)

	_, err = os.Stat(s.local.fullPath("test1.txt"))
	s.Nil(err)
}

func (s *LocalTestSuite) TestPath() {
	path := s.local.Path("test.txt")
	s.Equal(filepath.Join(s.local.root, "test.txt"), path)
}

func (s *LocalTestSuite) TestPut() {
	err := s.local.Put("test1.txt", "goravel")
	s.Nil(err)

	content, err := s.local.Get("test1.txt")
	s.Nil(err)
	s.Equal("goravel", content)

	err = os.Remove(s.local.fullPath("test1.txt"))
	s.Nil(err)
}

func (s *LocalTestSuite) TestPutFile() {
	path, err := s.local.PutFile("put", s.file)
	s.Nil(err)
	s.NotEmpty(path)

	content, err := s.local.Get(path)
	s.Nil(err)
	s.NotEmpty(content)

	err = os.Remove(s.local.fullPath(path))
	s.Nil(err)
}

func (s *LocalTestSuite) TestPutFileAs() {
	path, err := s.local.PutFileAs("put", s.file, "goravel")
	s.Nil(err)
	s.Equal(filepath.Join("put", "goravel.txt"), path)

	content, err := s.local.Get("put/goravel.txt")
	s.Nil(err)
	s.NotEmpty(content)

	err = os.Remove(s.local.fullPath("put/goravel.txt"))
	s.Nil(err)
}

func (s *LocalTestSuite) TestSize() {
	size, err := s.local.Size("test.txt")
	s.Nil(err)
	s.Equal(int64(7), size)
}

func (s *LocalTestSuite) TestTemporaryUrl() {
	url, err := s.local.TemporaryUrl("test.txt", time.Now().Add(1*time.Minute))
	s.Nil(err)
	s.Equal("https://goravel.dev/test.txt", url)
}

func (s *LocalTestSuite) TestWithContext() {
	driver := s.local.WithContext(context.Background())
	s.Equal(s.local, driver)
}

func (s *LocalTestSuite) TestUrl() {
	url := s.local.Url("test.txt")
	s.Equal("https://goravel.dev/test.txt", url)
}
