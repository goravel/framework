package translation

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/foundation/json"
	"github.com/goravel/framework/support/env"
	"github.com/goravel/framework/support/file"
)

type FileLoaderTestSuite struct {
	suite.Suite
	executable string
	json       foundation.Json
}

func TestFileLoaderTestSuite(t *testing.T) {
	assert.Nil(t, file.Create("lang/en/test.json", `{"foo": "bar", "baz": {"foo": "bar"}}`))
	assert.Nil(t, file.Create("lang/en/another/test.json", `{"foo": "backagebar", "baz": "backagesplash"}`))
	assert.Nil(t, file.Create("lang/another/en/test.json", `{"foo": "backagebar", "baz": "backagesplash"}`))
	assert.Nil(t, file.Create("lang/en/invalid/test.json", `{"foo": "bar",}`))
	assert.Nil(t, file.Create("lang/cn.json", `{"foo": "bar", "baz": {"foo": "bar"}}`))
	restrictedFilePath := "lang/en/restricted/test.json"
	assert.Nil(t, file.Create(restrictedFilePath, `{"foo": "restricted"}`))
	assert.Nil(t, os.Chmod(restrictedFilePath, 0000))

	// path2
	executable, err := executablePath()
	assert.NoError(t, err)
	assert.Nil(t, file.Create(filepath.Join(executable, "lang/en/test.json"), `{"foo": "bar", "baz": {"foo": "bar"}}`))
	assert.Nil(t, file.Create(filepath.Join(executable, "lang/en/another/test.json"), `{"foo": "backagebar", "baz": "backagesplash"}`))
	assert.Nil(t, file.Create(filepath.Join(executable, "lang/another/en/test.json"), `{"foo": "backagebar", "baz": "backagesplash"}`))
	assert.Nil(t, file.Create(filepath.Join(executable, "lang/en/invalid/test.json"), `{"foo": "bar",}`))
	assert.Nil(t, file.Create(filepath.Join(executable, "lang/cn.json"), `{"foo": "bar", "baz": {"foo": "bar"}}`))
	restrictedFilePath = filepath.Join(executable, "lang/en/restricted/test.json")
	assert.Nil(t, file.Create(restrictedFilePath, `{"foo": "restricted"}`))
	assert.Nil(t, os.Chmod(restrictedFilePath, 0000))

	suite.Run(t, &FileLoaderTestSuite{})
	assert.Nil(t, file.Remove("lang"))
	assert.Nil(t, file.Remove(filepath.Join(executable, "lang")))
}

func (f *FileLoaderTestSuite) SetupTest() {
	f.json = json.NewJson()
}

func (f *FileLoaderTestSuite) TestLoad() {
	paths := []string{"./lang"}
	loader := NewFileLoader(paths, f.json)
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

	paths = []string{"./lang/another", "./lang"}
	loader = NewFileLoader(paths, f.json)
	translations, err = loader.Load("en", "test")
	f.NoError(err)
	f.NotNil(translations)
	f.Equal("backagebar", translations["foo"])

	paths = []string{"./lang"}
	loader = NewFileLoader(paths, f.json)
	translations, err = loader.Load("en", "another/test")
	f.NoError(err)
	f.NotNil(translations)
	f.Equal("backagebar", translations["foo"])

	paths = []string{"./lang"}
	loader = NewFileLoader(paths, f.json)
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
	paths := []string{"./lang"}
	loader := NewFileLoader(paths, f.json)
	translations, err := loader.Load("hi", "test")

	f.Error(err)
	f.Nil(translations)
	f.Equal(ErrFileNotExist, err)
}

func (f *FileLoaderTestSuite) TestLoadInvalidJSON() {
	paths := []string{"./lang"}
	loader := NewFileLoader(paths, f.json)
	translations, err := loader.Load("en", "invalid/test")

	f.Error(err)
	f.Nil(translations)
}

func (f *FileLoaderTestSuite) TestLoadByExecutable() {
	paths := []string{filepath.Join(f.executable, "lang")}
	loader := NewFileLoader(paths, f.json)
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

	paths = []string{filepath.Join(f.executable, "lang", "another"), filepath.Join(f.executable, "lang")}
	loader = NewFileLoader(paths, f.json)
	translations, err = loader.Load("en", "test")
	f.NoError(err)
	f.NotNil(translations)
	f.Equal("backagebar", translations["foo"])

	paths = []string{filepath.Join(f.executable, "lang")}
	loader = NewFileLoader(paths, f.json)
	translations, err = loader.Load("en", "another/test")
	f.NoError(err)
	f.NotNil(translations)
	f.Equal("backagebar", translations["foo"])

	paths = []string{filepath.Join(f.executable, "lang")}
	loader = NewFileLoader(paths, f.json)
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

func (f *FileLoaderTestSuite) TestLoadNonExistentFileByExecutable() {
	paths := []string{filepath.Join(f.executable, "lang")}
	loader := NewFileLoader(paths, f.json)
	translations, err := loader.Load("hi", "test")

	f.Error(err)
	f.Nil(translations)
	f.Equal(ErrFileNotExist, err)
}

func (f *FileLoaderTestSuite) TestLoadInvalidJSONByExecutable() {
	paths := []string{filepath.Join(f.executable, "lang")}
	loader := NewFileLoader(paths, f.json)
	translations, err := loader.Load("en", "invalid/test")

	f.Error(err)
	f.Nil(translations)
}

func executablePath() (string, error) {
	executable, err := os.Executable()
	if err != nil {
		return "", err
	}

	return filepath.Dir(executable), nil
}
