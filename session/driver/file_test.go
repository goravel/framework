package driver

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/support/carbon"
	"github.com/goravel/framework/support/file"
)

type FileTestSuite struct {
	suite.Suite
}

func TestFileTestSuite(t *testing.T) {
	suite.Run(t, &FileTestSuite{})
}

func (f *FileTestSuite) BeforeTest() {
	f.Nil(file.Remove(f.getPath()))
}

func (f *FileTestSuite) TestClose() {
	driver := f.getDriver()
	f.Nil(driver.Close())
}

func (f *FileTestSuite) TestDestroy() {
	driver := f.getDriver()

	f.Nil(driver.Destroy("foo"))

	f.Nil(driver.Write("foo", "bar"))
	value, err := driver.Read("foo")
	f.Nil(err)
	f.Equal("bar", value)

	f.Nil(driver.Destroy("foo"))

	value, err = driver.Read("foo")
	f.NotNil(err)
	f.Equal("", value)
}

func (f *FileTestSuite) TestGc() {
	driver := f.getDriver()

	f.Nil(driver.Gc(300))

	f.Nil(driver.Write("foo", "bar"))
	f.Nil(driver.Gc(300))
	value, err := driver.Read("foo")
	f.Nil(err)
	f.Equal("bar", value)

	carbon.SetTestNow(carbon.Now(carbon.UTC).AddMinutes(5).AddSecond())
	f.Nil(driver.Gc(300))

	value, err = driver.Read("foo")
	f.NotNil(err)
	f.Equal("", value)
	carbon.UnsetTestNow()

	// file does not exist
	driver = NewFile("foo", 300)
	f.Nil(driver.Gc(300))
}

func (f *FileTestSuite) TestOpen() {
	driver := f.getDriver()
	f.Nil(driver.Open("", ""))
}

func (f *FileTestSuite) TestRead() {
	driver := f.getDriver()
	f.Nil(driver.Write("foo", "bar"))

	value, err := driver.Read("foo")
	f.Nil(err)
	f.Equal("bar", value)

	carbon.SetTestNow(carbon.Now(carbon.UTC).AddMinutes(f.getMinutes()).SubSecond())
	value, err = driver.Read("foo")
	f.Nil(err)
	f.Equal("bar", value)
	carbon.UnsetTestNow()

	carbon.SetTestNow(carbon.Now(carbon.UTC).AddMinutes(f.getMinutes()).AddSecond())
	value, err = driver.Read("foo")
	f.NotNil(err)
	f.Equal("", value)
	carbon.UnsetTestNow()
}

func (f *FileTestSuite) TestWrite() {
	driver := f.getDriver()
	f.Nil(driver.Write("foo", "bar"))
	value, err := driver.Read("foo")
	f.Nil(err)
	f.Equal("bar", value)

	f.Nil(driver.Write("foo", "qux"))
	value, err = driver.Read("foo")
	f.Nil(err)
	f.Equal("qux", value)

	f.Nil(file.Remove(f.getPath()))
}

func BenchmarkFile_ReadWrite(b *testing.B) {
	f := new(FileTestSuite)
	f.SetT(&testing.T{})

	driver := f.getDriver()
	f.Nil(driver.Write("foo", "bar"))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		f.Nil(driver.Write("foo", "bar"))

		value, err := driver.Read("foo")
		f.Nil(err)
		f.Equal("bar", value)
	}
	b.StopTimer()

	f.BeforeTest()
}

func (f *FileTestSuite) getDriver() *File {
	return NewFile(f.getPath(), f.getMinutes())
}

func (f *FileTestSuite) getPath() string {
	return "test"
}

func (f *FileTestSuite) getMinutes() int {
	return 10
}
