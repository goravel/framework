package filesystem

import (
	"context"
	"mime"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	configmock "github.com/goravel/framework/contracts/config/mocks"
	"github.com/goravel/framework/support/carbon"
	"github.com/goravel/framework/support/env"
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
	s.Nil(s.local.Put("AllDirectories/1.txt", "Goravel"))
	s.Nil(s.local.Put("AllDirectories/2.txt", "Goravel"))
	s.Nil(s.local.Put("AllDirectories/3/3.txt", "Goravel"))
	s.Nil(s.local.Put("AllDirectories/3/5/6/6.txt", "Goravel"))
	s.Nil(s.local.MakeDirectory("AllDirectories/3/4"))
	s.True(s.local.Exists("AllDirectories/1.txt"))
	s.True(s.local.Exists("AllDirectories/2.txt"))
	s.True(s.local.Exists("AllDirectories/3/3.txt"))
	s.True(s.local.Exists("AllDirectories/3/4/"))
	s.True(s.local.Exists("AllDirectories/3/5/6/6.txt"))
	files, err := s.local.AllDirectories("AllDirectories")
	s.Nil(err)
	if env.IsWindows() {
		s.Equal([]string{"3\\", "3\\4\\", "3\\5\\", "3\\5\\6\\"}, files)
	} else {
		s.Equal([]string{"3/", "3/4/", "3/5/", "3/5/6/"}, files)
	}
	files, err = s.local.AllDirectories("./AllDirectories")
	s.Nil(err)
	if env.IsWindows() {
		s.Equal([]string{"3\\", "3\\4\\", "3\\5\\", "3\\5\\6\\"}, files)
	} else {
		s.Equal([]string{"3/", "3/4/", "3/5/", "3/5/6/"}, files)
	}
	files, err = s.local.AllDirectories("/AllDirectories")
	s.Nil(err)
	if env.IsWindows() {
		s.Equal([]string{"3\\", "3\\4\\", "3\\5\\", "3\\5\\6\\"}, files)
	} else {
		s.Equal([]string{"3/", "3/4/", "3/5/", "3/5/6/"}, files)
	}
	files, err = s.local.AllDirectories("./AllDirectories/")
	s.Nil(err)
	if env.IsWindows() {
		s.Equal([]string{"3\\", "3\\4\\", "3\\5\\", "3\\5\\6\\"}, files)
	} else {
		s.Equal([]string{"3/", "3/4/", "3/5/", "3/5/6/"}, files)
	}
	s.Nil(s.local.DeleteDirectory("AllDirectories"))
}

func (s *LocalTestSuite) TestAllFiles() {
	s.Nil(s.local.Put("AllFiles/1.txt", "Goravel"))
	s.Nil(s.local.Put("AllFiles/2.txt", "Goravel"))
	s.Nil(s.local.Put("AllFiles/3/3.txt", "Goravel"))
	s.Nil(s.local.Put("AllFiles/3/4/4.txt", "Goravel"))
	s.True(s.local.Exists("AllFiles/1.txt"))
	s.True(s.local.Exists("AllFiles/2.txt"))
	s.True(s.local.Exists("AllFiles/3/3.txt"))
	s.True(s.local.Exists("AllFiles/3/4/4.txt"))
	files, err := s.local.AllFiles("AllFiles")
	s.Nil(err)
	if env.IsWindows() {
		s.Equal([]string{"1.txt", "2.txt", "3\\3.txt", "3\\4\\4.txt"}, files)
	} else {
		s.Equal([]string{"1.txt", "2.txt", "3/3.txt", "3/4/4.txt"}, files)
	}
	files, err = s.local.AllFiles("./AllFiles")
	s.Nil(err)
	if env.IsWindows() {
		s.Equal([]string{"1.txt", "2.txt", "3\\3.txt", "3\\4\\4.txt"}, files)
	} else {
		s.Equal([]string{"1.txt", "2.txt", "3/3.txt", "3/4/4.txt"}, files)
	}
	files, err = s.local.AllFiles("/AllFiles")
	s.Nil(err)
	if env.IsWindows() {
		s.Equal([]string{"1.txt", "2.txt", "3\\3.txt", "3\\4\\4.txt"}, files)
	} else {
		s.Equal([]string{"1.txt", "2.txt", "3/3.txt", "3/4/4.txt"}, files)
	}
	files, err = s.local.AllFiles("./AllFiles/")
	s.Nil(err)
	if env.IsWindows() {
		s.Equal([]string{"1.txt", "2.txt", "3\\3.txt", "3\\4\\4.txt"}, files)
	} else {
		s.Equal([]string{"1.txt", "2.txt", "3/3.txt", "3/4/4.txt"}, files)
	}
	s.Nil(s.local.DeleteDirectory("AllFiles"))
}

func (s *LocalTestSuite) TestCopy() {
	s.Nil(s.local.Put("Copy/1.txt", "Goravel"))
	s.True(s.local.Exists("Copy/1.txt"))
	s.Nil(s.local.Copy("Copy/1.txt", "Copy1/1.txt"))
	s.True(s.local.Exists("Copy/1.txt"))
	s.True(s.local.Exists("Copy1/1.txt"))
	s.Nil(s.local.DeleteDirectory("Copy"))
	s.Nil(s.local.DeleteDirectory("Copy1"))
}

func (s *LocalTestSuite) TestDelete() {
	s.Nil(s.local.Put("Delete/1.txt", "Goravel"))
	s.True(s.local.Exists("Delete/1.txt"))
	s.Nil(s.local.Delete("Delete/1.txt"))
	s.True(s.local.Missing("Delete/1.txt"))
	s.Nil(s.local.DeleteDirectory("Delete"))
}

func (s *LocalTestSuite) TestDeleteDirectory() {
	s.Nil(s.local.Put("DeleteDirectory/1.txt", "Goravel"))
	s.True(s.local.Exists("DeleteDirectory/1.txt"))
	s.Nil(s.local.DeleteDirectory("DeleteDirectory"))
	s.True(s.local.Missing("DeleteDirectory/1.txt"))
	s.Nil(s.local.DeleteDirectory("DeleteDirectory"))
}

func (s *LocalTestSuite) TestDirectories() {
	s.Nil(s.local.Put("Directories/1.txt", "Goravel"))
	s.Nil(s.local.Put("Directories/2.txt", "Goravel"))
	s.Nil(s.local.Put("Directories/3/3.txt", "Goravel"))
	s.Nil(s.local.Put("Directories/3/5/5.txt", "Goravel"))
	s.Nil(s.local.MakeDirectory("Directories/3/4"))
	s.True(s.local.Exists("Directories/1.txt"))
	s.True(s.local.Exists("Directories/2.txt"))
	s.True(s.local.Exists("Directories/3/3.txt"))
	s.True(s.local.Exists("Directories/3/4/"))
	s.True(s.local.Exists("Directories/3/5/5.txt"))
	files, err := s.local.Directories("Directories")
	s.Nil(err)
	if env.IsWindows() {
		s.Equal([]string{"3\\"}, files)
	} else {
		s.Equal([]string{"3/"}, files)
	}
	files, err = s.local.Directories("./Directories")
	s.Nil(err)
	if env.IsWindows() {
		s.Equal([]string{"3\\"}, files)
	} else {
		s.Equal([]string{"3/"}, files)
	}
	files, err = s.local.Directories("/Directories")
	s.Nil(err)
	if env.IsWindows() {
		s.Equal([]string{"3\\"}, files)
	} else {
		s.Equal([]string{"3/"}, files)
	}
	files, err = s.local.Directories("./Directories/")
	s.Nil(err)
	if env.IsWindows() {
		s.Equal([]string{"3\\"}, files)
	} else {
		s.Equal([]string{"3/"}, files)
	}
	s.Nil(s.local.DeleteDirectory("Directories"))
}

func (s *LocalTestSuite) TestExists() {
	exists := s.local.Exists("test.txt")
	s.True(exists)

	exists = s.local.Exists("test1.txt")
	s.False(exists)
}

func (s *LocalTestSuite) TestFiles() {
	s.Nil(s.local.Put("Files/1.txt", "Goravel"))
	s.Nil(s.local.Put("Files/2.txt", "Goravel"))
	s.Nil(s.local.Put("Files/3/3.txt", "Goravel"))
	s.Nil(s.local.Put("Files/3/4/4.txt", "Goravel"))
	s.True(s.local.Exists("Files/1.txt"))
	s.True(s.local.Exists("Files/2.txt"))
	s.True(s.local.Exists("Files/3/3.txt"))
	s.True(s.local.Exists("Files/3/4/4.txt"))
	files, err := s.local.Files("Files")
	s.Nil(err)
	s.Equal([]string{"1.txt", "2.txt"}, files)
	files, err = s.local.Files("./Files")
	s.Nil(err)
	s.Equal([]string{"1.txt", "2.txt"}, files)
	files, err = s.local.Files("/Files")
	s.Nil(err)
	s.Equal([]string{"1.txt", "2.txt"}, files)
	files, err = s.local.Files("./Files/")
	s.Nil(err)
	s.Equal([]string{"1.txt", "2.txt"}, files)
	s.Nil(s.local.DeleteDirectory("Files"))
}

func (s *LocalTestSuite) TestGet() {
	s.Nil(s.local.Put("Get/1.txt", "Goravel"))
	s.True(s.local.Exists("Get/1.txt"))
	data, err := s.local.Get("Get/1.txt")
	s.Nil(err)
	s.Equal("Goravel", data)
	length, err := s.local.Size("Get/1.txt")
	s.Nil(err)
	s.Equal(int64(7), length)
	s.Nil(s.local.DeleteDirectory("Get"))
}

func (s *LocalTestSuite) TestGetBytes() {
	s.Nil(s.local.Put("Get/1.txt", "Goravel"))
	s.True(s.local.Exists("Get/1.txt"))
	data, err := s.local.GetBytes("Get/1.txt")
	s.Nil(err)
	s.Equal([]byte("Goravel"), data)
	length, err := s.local.Size("Get/1.txt")
	s.Nil(err)
	s.Equal(int64(7), length)
	s.Nil(s.local.DeleteDirectory("Get"))
}

func (s *LocalTestSuite) TestLastModified() {
	s.mockConfig.On("GetString", "app.timezone").Return("UTC").Once()

	s.Nil(s.local.Put("LastModified/1.txt", "Goravel"))
	s.True(s.local.Exists("LastModified/1.txt"))
	date, err := s.local.LastModified("LastModified/1.txt")
	s.Nil(err)

	s.Nil(err)
	s.Equal(carbon.Now().ToDateString(), carbon.FromStdTime(date).ToDateString())
	s.Nil(s.local.DeleteDirectory("LastModified"))
}

func (s *LocalTestSuite) TestMakeDirectory() {
	s.Nil(s.local.MakeDirectory("MakeDirectory1/"))
	s.Nil(s.local.MakeDirectory("MakeDirectory2"))
	s.Nil(s.local.MakeDirectory("MakeDirectory3/MakeDirectory4"))
	s.Nil(s.local.DeleteDirectory("MakeDirectory1"))
	s.Nil(s.local.DeleteDirectory("MakeDirectory2"))
	s.Nil(s.local.DeleteDirectory("MakeDirectory3"))
	s.Nil(s.local.DeleteDirectory("MakeDirectory4"))
}

func (s *LocalTestSuite) TestMimeType_File() {
	s.Nil(s.local.Put("MimeType/1.txt", "Goravel"))
	s.True(s.local.Exists("MimeType/1.txt"))
	mimeType, err := s.local.MimeType("MimeType/1.txt")
	s.Nil(err)
	mediaType, _, err := mime.ParseMediaType(mimeType)
	s.Nil(err)
	s.Equal("text/plain", mediaType)
}

func (s *LocalTestSuite) TestMimeType_Image() {
	s.mockConfig.On("GetString", "filesystems.default").Return("local").Once()

	fileInfo, err := NewFile("../logo.png")
	s.Nil(err)
	path, err := s.local.PutFile("MimeType", fileInfo)
	s.Nil(err)
	s.True(s.local.Exists(path))
	mimeType, err := s.local.MimeType(path)
	s.Nil(err)
	s.Equal("image/png", mimeType)
}

func (s *LocalTestSuite) TestMissing() {
	missing := s.local.Missing("test.txt")
	s.False(missing)

	missing = s.local.Missing("test1.txt")
	s.True(missing)
}

func (s *LocalTestSuite) TestMove() {
	s.Nil(s.local.Put("Move/1.txt", "Goravel"))
	s.True(s.local.Exists("Move/1.txt"))
	s.Nil(s.local.Move("Move/1.txt", "Move1/1.txt"))
	s.True(s.local.Missing("Move/1.txt"))
	s.True(s.local.Exists("Move1/1.txt"))
	s.Nil(s.local.DeleteDirectory("Move"))
	s.Nil(s.local.DeleteDirectory("Move1"))
}

func (s *LocalTestSuite) TestPath() {
	path := s.local.Path("test.txt")
	s.Equal(filepath.Join(s.local.root, "test.txt"), path)
}

func (s *LocalTestSuite) TestPut() {
	s.Nil(s.local.Put("Put/1.txt", "Goravel"))
	s.True(s.local.Exists("Put/1.txt"))
	s.True(s.local.Missing("Put/2.txt"))
	s.Nil(s.local.DeleteDirectory("Put"))
}

func (s *LocalTestSuite) TestPutFile_Text() {
	path, err := s.local.PutFile("PutFile", s.file)
	s.Nil(err)
	s.True(s.local.Exists(path))
	data, err := s.local.Get(path)
	s.Nil(err)
	s.NotEmpty(data)
	s.Nil(s.local.DeleteDirectory("PutFile"))
}

func (s *LocalTestSuite) TestPutFile_Image() {
	s.mockConfig.On("GetString", "filesystems.default").Return("local").Once()

	fileInfo, err := NewFile("../logo.png")
	s.Nil(err)
	path, err := s.local.PutFile("PutFile1", fileInfo)
	s.Nil(err)
	s.True(s.local.Exists(path))
	s.Nil(s.local.DeleteDirectory("PutFile1"))
}

func (s *LocalTestSuite) TestPutFileAs_Text() {
	path, err := s.local.PutFileAs("PutFileAs", s.file, "text")
	s.Nil(err)
	s.Equal(filepath.Join("PutFileAs", "text.txt"), path)
	s.True(s.local.Exists(path))
	data, err := s.local.Get(path)
	s.Nil(err)
	s.NotEmpty(data)

	path, err = s.local.PutFileAs("PutFileAs", s.file, "text1.txt")
	s.Nil(err)
	s.Equal(filepath.Join("PutFileAs", "text1.txt"), path)
	s.True(s.local.Exists(path))
	data, err = s.local.Get(path)
	s.Nil(err)
	s.NotEmpty(data)

	s.Nil(s.local.DeleteDirectory("PutFileAs"))
}

func (s *LocalTestSuite) TestPutFileAs_Image() {
	s.mockConfig.On("GetString", "filesystems.default").Return("local").Once()

	fileInfo, err := NewFile("../logo.png")
	s.Nil(err)
	path, err := s.local.PutFileAs("PutFileAs1", fileInfo, "image")
	s.Nil(err)
	s.Equal(filepath.Join("PutFileAs1", "image.png"), path)
	s.True(s.local.Exists(path))

	path, err = s.local.PutFileAs("PutFileAs1", fileInfo, "image1.png")
	s.Nil(err)
	s.Equal(filepath.Join("PutFileAs1", "image1.png"), path)
	s.True(s.local.Exists(path))

	s.Nil(s.local.DeleteDirectory("PutFileAs1"))
}

func (s *LocalTestSuite) TestSize() {
	s.Nil(s.local.Put("Size/1.txt", "Goravel"))
	s.True(s.local.Exists("Size/1.txt"))
	length, err := s.local.Size("Size/1.txt")
	s.Nil(err)
	s.Equal(int64(7), length)
	s.Nil(s.local.DeleteDirectory("Size"))
}

func (s *LocalTestSuite) TestTemporaryUrl() {
	s.Nil(s.local.Put("TemporaryUrl/1.txt", "Goravel"))
	s.True(s.local.Exists("TemporaryUrl/1.txt"))
	url, err := s.local.TemporaryUrl("TemporaryUrl/1.txt", carbon.Now().AddSeconds(5).ToStdTime())
	s.Nil(err)
	s.NotEmpty(url)
	s.Nil(s.local.DeleteDirectory("TemporaryUrl"))
}

func (s *LocalTestSuite) TestWithContext() {
	driver := s.local.WithContext(context.Background())
	s.Equal(s.local, driver)
}

func (s *LocalTestSuite) TestUrl() {
	s.Equal("https://goravel.dev/Url/1.txt", s.local.Url("Url/1.txt"))

	if env.IsWindows() {
		s.Equal("https://goravel.dev/Url/2.txt", s.local.Url(`Url\2.txt`))
	}
}
