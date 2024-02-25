package handler

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/support/carbon"
	"github.com/goravel/framework/support/file"
)

type FileHandlerTestSuite struct {
	suite.Suite
}

func TestFileHandlerTestSuite(t *testing.T) {
	suite.Run(t, &FileHandlerTestSuite{})
}

func (f *FileHandlerTestSuite) AfterTest() {
	f.Nil(file.Remove(f.getPath()))
}

func (f *FileHandlerTestSuite) BeforeTest() {
	f.Nil(file.Remove(f.getPath()))
}

func (f *FileHandlerTestSuite) TestClose() {
	handler := f.getHandler()
	f.True(handler.Close())
}

func (f *FileHandlerTestSuite) TestDestroy() {
	handler := f.getHandler()
	f.True(handler.Destroy("foo"))

	err := handler.Write("foo", "bar")
	f.Nil(err)

	f.Equal("bar", handler.Read("foo"))
	f.True(handler.Destroy("foo"))
	f.Equal("", handler.Read("foo"))
}

func (f *FileHandlerTestSuite) TestGc() {
	handler := f.getHandler()

	f.Equal(0, handler.Gc(300))

	f.Nil(handler.Write("foo", "bar"))
	f.Equal(0, handler.Gc(300))
	f.Equal("bar", handler.Read("foo"))

	carbon.SetTestNow(carbon.Now(carbon.UTC).AddMinutes(5))
	f.Nil(handler.Write("baz", "qux"))
	carbon.UnsetTestNow()

	carbon.SetTestNow(carbon.Now(carbon.UTC).AddMinutes(6).AddSecond())
	f.Equal(2, handler.Gc(300))
	f.Equal("", handler.Read("foo"))
	f.Equal("", handler.Read("baz"))
	carbon.UnsetTestNow()
}

func (f *FileHandlerTestSuite) TestOpen() {
	handler := f.getHandler()
	f.True(handler.Open("", ""))
}

func (f *FileHandlerTestSuite) TestRead() {
	handler := f.getHandler()
	err := handler.Write("foo", "bar")
	f.Nil(err)

	f.Equal("bar", handler.Read("foo"))

	carbon.SetTestNow(carbon.Now(carbon.UTC).AddMinutes(f.getMinutes()))
	f.Equal("bar", handler.Read("foo"))
	carbon.UnsetTestNow()

	carbon.SetTestNow(carbon.Now().AddMinutes(f.getMinutes()).AddSecond())
	f.Equal("", handler.Read("foo"))
	carbon.UnsetTestNow()
}

func (f *FileHandlerTestSuite) TestWrite() {
	handler := f.getHandler()

	f.Nil(handler.Write("bar", "baz"))
	f.Equal("baz", handler.Read("bar"))

	f.Nil(handler.Write("bar", "qux"))
	f.Equal("qux", handler.Read("bar"))

	f.Nil(file.Remove(f.getPath()))
}

func (f *FileHandlerTestSuite) getHandler() *FileHandler {
	return NewFileHandler(f.getPath(), f.getMinutes())
}

func (f *FileHandlerTestSuite) getPath() string {
	return "test"
}

func (f *FileHandlerTestSuite) getMinutes() int {
	return 10
}
