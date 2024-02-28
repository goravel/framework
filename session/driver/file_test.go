package driver

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	mockfilesystem "github.com/goravel/framework/mocks/filesystem"
	"github.com/goravel/framework/support/carbon"
)

type FileDriverTestSuite struct {
	suite.Suite
	mockFiles *mockfilesystem.Storage
}

func TestFileDriverTestSuite(t *testing.T) {
	suite.Run(t, &FileDriverTestSuite{})
}

func (f *FileDriverTestSuite) SetupTest() {
	f.mockFiles = mockfilesystem.NewStorage(f.T())
}

func (f *FileDriverTestSuite) TestClose() {
	driver := f.getDriver()
	f.True(driver.Close())
}

func (f *FileDriverTestSuite) TestDestroy() {
	driver := f.getDriver()

	f.mockFiles.On("Delete", f.getPath()+"/foo").Return(nil).Once()
	f.Nil(driver.Destroy("foo"))

	f.mockFiles.On("Put", f.getPath()+"/foo", "bar").Return(nil).Once()
	err := driver.Write("foo", "bar")
	f.Nil(err)

	f.mockFiles.On("Exists", f.getPath()+"/foo").Return(true).Once()
	f.mockFiles.On("LastModified", f.getPath()+"/foo").Once().Return(time.Now(), nil)
	f.mockFiles.On("Get", f.getPath()+"/foo").Once().Return("bar", nil)
	f.Equal("bar", driver.Read("foo"))

	f.mockFiles.On("Delete", f.getPath()+"/foo").Return(nil).Once()
	f.Nil(driver.Destroy("foo"))

	f.mockFiles.On("Exists", f.getPath()+"/foo").Return(false).Once()
	f.Equal("", driver.Read("foo"))
}

func (f *FileDriverTestSuite) TestGc() {
	driver := f.getDriver()

	f.mockFiles.On("Files", f.getPath()).Return([]string{}, errors.New("error")).Once()
	f.Equal(0, driver.Gc(300))

	f.mockFiles.On("Files", f.getPath()).Return([]string{"foo"}, nil).Once()
	f.mockFiles.On("LastModified", f.getPath()+"/foo").Return(time.Now(), nil).Once()
	f.Equal(0, driver.Gc(300))

	f.mockFiles.On("Files", f.getPath()).Return([]string{"foo"}, nil).Once()
	f.mockFiles.On("LastModified", f.getPath()+"/foo").Return(time.Time{}, errors.New("error")).Once()
	f.Equal(0, driver.Gc(300))

	f.mockFiles.On("Files", f.getPath()).Return([]string{"foo"}, nil).Once()
	f.mockFiles.On("LastModified", f.getPath()+"/foo").Return(carbon.Now().SubMinutes(f.getMinutes()).StdTime(), nil).Once()
	f.mockFiles.On("Delete", f.getPath()+"/foo").Return(nil).Once()
	f.Equal(1, driver.Gc(300))
}

func (f *FileDriverTestSuite) TestOpen() {
	driver := f.getDriver()
	f.True(driver.Open("", ""))
}

func (f *FileDriverTestSuite) TestRead() {
	driver := f.getDriver()

	f.mockFiles.On("Exists", f.getPath()+"/foo").Return(true).Once()
	f.mockFiles.On("LastModified", f.getPath()+"/foo").Once().Return(time.Now(), nil)
	f.mockFiles.On("Get", f.getPath()+"/foo").Once().Return("bar", nil)
	f.Equal("bar", driver.Read("foo"))

	f.mockFiles.On("Exists", f.getPath()+"/foo").Return(true).Once()
	f.mockFiles.On("LastModified", f.getPath()+"/foo").
		Return(carbon.Now().SubMinutes(f.getMinutes()).AddSecond().StdTime(), nil).Once()
	f.mockFiles.On("Get", f.getPath()+"/foo").Return("bar", nil).Once()
	f.Equal("bar", driver.Read("foo"))

	f.mockFiles.On("Exists", f.getPath()+"/foo").Return(true).Once()
	f.mockFiles.On("LastModified", f.getPath()+"/foo").
		Return(carbon.Now().SubMinutes(f.getMinutes()).StdTime(), nil).Once()
	f.Equal("", driver.Read("foo"))

	f.mockFiles.On("Exists", f.getPath()+"/foo").Return(true).Once()
	f.mockFiles.On("LastModified", f.getPath()+"/foo").Return(time.Time{}, errors.New("error")).Once()
	f.Equal("", driver.Read("foo"))

	// error when reading file content
	f.mockFiles.On("Exists", f.getPath()+"/foo").Return(true).Once()
	f.mockFiles.On("LastModified", f.getPath()+"/foo").Once().Return(time.Now(), nil)
	f.mockFiles.On("Get", f.getPath()+"/foo").Once().Return("", errors.New("error"))
	f.Equal("", driver.Read("foo"))
}

func (f *FileDriverTestSuite) TestWrite() {
	driver := f.getDriver()

	f.mockFiles.On("Put", f.getPath()+"/foo", "bar").Return(nil).Once()
	f.Nil(driver.Write("foo", "bar"))
}

func (f *FileDriverTestSuite) getDriver() *FileDriver {
	return NewFileDriver(f.mockFiles, f.getPath(), f.getMinutes())
}

func (f *FileDriverTestSuite) getPath() string {
	return "test"
}

func (f *FileDriverTestSuite) getMinutes() int {
	return 10
}
