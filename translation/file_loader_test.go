package translation

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/support/env"
	"github.com/goravel/framework/support/file"
)

type FileLoaderTestSuite struct {
	suite.Suite
}

func TestFileLoaderTestSuite(t *testing.T) {
	assert.Nil(t, file.Create("lang/en/test.json", `{"foo": "bar", "baz": {"foo": "bar"}}`))
	assert.Nil(t, file.Create("lang/en/another/test.json", `{"foo": "backagebar", "baz": "backagesplash"}`))
	assert.Nil(t, file.Create("lang/another/en/test.json", `{"foo": "backagebar", "baz": "backagesplash"}`))
	assert.Nil(t, file.Create("lang/en/invalid/test.json", `{"foo": "bar",}`))
	restrictedFilePath := "lang/en/restricted/test.json"
	assert.Nil(t, file.Create(restrictedFilePath, `{"foo": "restricted"}`))
	assert.Nil(t, os.Chmod(restrictedFilePath, 0000))
	suite.Run(t, &FileLoaderTestSuite{})
	assert.Nil(t, file.Remove("lang"))
}

func (f *FileLoaderTestSuite) SetupTest() {
}

func (f *FileLoaderTestSuite) TestLoad() {
	paths := []string{"./lang"}
	loader := NewFileLoader(paths)
	translations, err := loader.Load("test", "en")
	f.NoError(err)
	f.NotNil(translations)
	f.Equal("bar", translations["en"]["foo"])

	paths = []string{"./lang/another", "./lang"}
	loader = NewFileLoader(paths)
	translations, err = loader.Load("test", "en")
	f.NoError(err)
	f.NotNil(translations)
	f.Equal("bar", translations["en"]["foo"])

	paths = []string{"./lang"}
	loader = NewFileLoader(paths)
	translations, err = loader.Load("another/test", "en")
	f.NoError(err)
	f.NotNil(translations)
	f.Equal("backagebar", translations["en"]["foo"])

	paths = []string{"./lang"}
	loader = NewFileLoader(paths)
	translations, err = loader.Load("restricted/test", "en")
	if env.IsWindows() {
		f.NoError(err)
		f.NotNil(translations)
		f.Equal("restricted", translations["en"]["foo"])
	} else {
		f.Error(err)
		f.Nil(translations)
	}
}

func (f *FileLoaderTestSuite) TestLoadNonExistentFile() {
	paths := []string{"./lang"}
	loader := NewFileLoader(paths)
	translations, err := loader.Load("test", "hi")

	f.Error(err)
	f.Nil(translations)
	f.Equal(ErrFileNotExist, err)
}

func (f *FileLoaderTestSuite) TestLoadInvalidJSON() {
	paths := []string{"./lang"}
	loader := NewFileLoader(paths)
	translations, err := loader.Load("invalid/test", "en")

	f.Error(err)
	f.Nil(translations)
}

func TestMergeMaps(t *testing.T) {
	dst := map[string]string{
		"foo": "bar",
	}
	src := map[string]string{
		"baz": "backage",
	}
	mergeMaps(dst, src)
	assert.Equal(t, map[string]string{
		"foo": "bar",
		"baz": "backage",
	}, dst)
}
