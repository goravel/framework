package driver

import (
	"os"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/support/carbon"
	"github.com/goravel/framework/support/env"
	"github.com/goravel/framework/support/file"
)

type FileDriverTestSuite struct {
	suite.Suite
}

func TestFileDriverTestSuite(t *testing.T) {
	suite.Run(t, &FileDriverTestSuite{})
}

func (f *FileDriverTestSuite) AfterTest() {
	f.Nil(file.Remove(f.getPath()))
}

func (f *FileDriverTestSuite) BeforeTest() {
	f.Nil(file.Remove(f.getPath()))
}

func (f *FileDriverTestSuite) TestClose() {
	driver := f.getDriver()
	f.True(driver.Close())
}

func (f *FileDriverTestSuite) TestDestroy() {
	driver := f.getDriver()
	f.Nil(driver.Destroy("foo"))

	err := driver.Write("foo", "bar")
	f.Nil(err)

	f.Equal("bar", driver.Read("foo"))
	f.Nil(driver.Destroy("foo"))
	f.Equal("", driver.Read("foo"))
}

func (f *FileDriverTestSuite) TestGc() {
	driver := f.getDriver()

	f.Equal(0, driver.Gc(300))

	f.Nil(driver.Write("foo", "bar"))
	f.Equal(0, driver.Gc(300))
	f.Equal("bar", driver.Read("foo"))

	carbon.SetTestNow(carbon.Now().AddSecond())
	f.Nil(driver.Write("baz", "qux"))
	carbon.UnsetTestNow()

	carbon.SetTestNow(carbon.Now(carbon.UTC).AddMinutes(6).AddSecond())
	f.Equal(2, driver.Gc(300))
	f.Equal("", driver.Read("foo"))
	f.Equal("", driver.Read("baz"))
	carbon.UnsetTestNow()
}

func (f *FileDriverTestSuite) TestOpen() {
	driver := f.getDriver()
	f.True(driver.Open("", ""))
}

func (f *FileDriverTestSuite) TestRead() {
	driver := f.getDriver()
	err := driver.Write("foo", "bar")
	f.Nil(err)

	f.Equal("bar", driver.Read("foo"))

	carbon.SetTestNow(carbon.Now(carbon.UTC).AddMinutes(f.getMinutes()))
	f.Equal("bar", driver.Read("foo"))
	carbon.UnsetTestNow()

	carbon.SetTestNow(carbon.Now().AddMinutes(f.getMinutes()).AddSecond())
	f.Equal("", driver.Read("foo"))
	carbon.UnsetTestNow()

	// error when reading file content
	restrictedFilePath := f.getPath() + "/foo"
	f.Nil(os.Chmod(restrictedFilePath, 0000))
	if env.IsWindows() {
		f.Equal("bar", driver.Read("foo"))
	} else {
		f.Equal("", driver.Read("foo"))
	}
	f.Nil(os.Chmod(restrictedFilePath, 0777))
}

func (f *FileDriverTestSuite) TestWrite() {
	driver := f.getDriver()

	f.Nil(driver.Write("bar", "baz"))
	f.Equal("baz", driver.Read("bar"))

	f.Nil(driver.Write("bar", "qux"))
	f.Equal("qux", driver.Read("bar"))

	f.Nil(file.Remove(f.getPath()))
}

func (f *FileDriverTestSuite) getDriver() *FileDriver {
	return NewFileDriver(f.getPath(), f.getMinutes())
}

func (f *FileDriverTestSuite) getPath() string {
	return "test"
}

func (f *FileDriverTestSuite) getMinutes() int {
	return 10
}
