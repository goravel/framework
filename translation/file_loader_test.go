package translation

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/support/env"
	"github.com/goravel/framework/support/file"
	"github.com/goravel/framework/support/path"
)

type FileLoaderTestSuite struct {
	suite.Suite
}

func TestFileLoaderTestSuite(t *testing.T) {
	assert.Nil(t, file.Create("lang/en/test.json", `{"foo": "bar", "baz": {"foo": "bar"}}`))
	assert.Nil(t, file.Create("lang/en/another/test.json", `{"foo": "backagebar", "baz": "backagesplash"}`))
	assert.Nil(t, file.Create("lang/another/en/test.json", `{"foo": "backagebar", "baz": "backagesplash"}`))
	assert.Nil(t, file.Create("lang/en/invalid/test.json", `{"foo": "bar",}`))
	// We should adapt this situation.
	assert.Nil(t, file.Create("lang/cn.json", `{"foo": "bar", "baz": {"foo": "bar"}}`))
	restrictedFilePath := "lang/en/restricted/test.json"
	assert.Nil(t, file.Create(restrictedFilePath, `{"foo": "restricted"}`))
	assert.Nil(t, os.Chmod(restrictedFilePath, 0000))
	suite.Run(t, &FileLoaderTestSuite{})
	assert.Nil(t, file.Remove("lang"))
}

func (f *FileLoaderTestSuite) SetupTest() {
}

func (f *FileLoaderTestSuite) TestLoad() {
	executable, err := path.Executable()
	assert.NoError(f.T(), err)

	paths := []string{filepath.Join(executable, "lang")}
	loader := NewFileLoader(paths)
	translations, err := loader.Load("en", "test")
	f.NoError(err)
	f.NotNil(translations)
	f.Equal("bar", translations["foo"])
	f.Equal("bar", translations["baz"].(map[string]any)["foo"])

	translations, err = loader.Load("cn", "*")
	f.NoError(err)
	f.NotNil(translations)
	f.Equal("bar", translations["foo"])
	f.Equal("bar", translations["baz"].(map[string]any)["foo"])

	paths = []string{filepath.Join(executable, "lang", "another"), filepath.Join(executable, "lang")}
	loader = NewFileLoader(paths)
	translations, err = loader.Load("en", "test")
	f.NoError(err)
	f.NotNil(translations)
	f.Equal("backagebar", translations["foo"])

	paths = []string{filepath.Join(executable, "lang")}
	loader = NewFileLoader(paths)
	translations, err = loader.Load("en", "another/test")
	f.NoError(err)
	f.NotNil(translations)
	f.Equal("backagebar", translations["foo"])

	paths = []string{filepath.Join(executable, "lang")}
	loader = NewFileLoader(paths)
	translations, err = loader.Load("en", "restricted/test")
	if env.IsWindows() {
		f.NoError(err)
		f.NotNil(translations)
		f.Equal("restricted", translations["foo"])
	} else {
		f.Error(err)
		f.Nil(translations)
	}
}

func (f *FileLoaderTestSuite) TestLoadNonExistentFile() {
	executable, err := path.Executable()
	assert.NoError(f.T(), err)

	paths := []string{filepath.Join(executable, "lang")}
	loader := NewFileLoader(paths)
	translations, err := loader.Load("hi", "test")

	f.Error(err)
	f.Nil(translations)
	f.Equal(ErrFileNotExist, err)
}

func (f *FileLoaderTestSuite) TestLoadInvalidJSON() {
	executable, err := path.Executable()
	assert.NoError(f.T(), err)

	paths := []string{filepath.Join(executable, "lang")}
	loader := NewFileLoader(paths)
	translations, err := loader.Load("en", "invalid/test")

	f.Error(err)
	f.Nil(translations)
}
