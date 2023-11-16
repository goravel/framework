package translation

import (
	"testing"

	"github.com/goravel/framework/support/file"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type FileLoaderTestSuite struct {
	suite.Suite
}

func TestFileLoaderTestSuite(t *testing.T) {
	assert.Nil(t, file.Create("lang/en.json", `{"foo": "bar"}`))
	assert.Nil(t, file.Create("lang/another/en.json", `{"foo": "backagebar", "baz": "backagesplash"}`))
	assert.Nil(t, file.Create("lang/invalid/en.json", `{"foo": "bar",}`))
	suite.Run(t, &FileLoaderTestSuite{})
	assert.Nil(t, file.Remove("lang"))
}

func (f *FileLoaderTestSuite) SetupTest() {
}

func (f *FileLoaderTestSuite) TestLoad() {
	paths := []string{"./lang"}
	loader := NewFileLoader(paths)
	translations, err := loader.Load("*", "en")
	f.NoError(err)
	f.NotNil(translations)
	f.Equal("bar", translations["en"]["foo"])

	paths = []string{"./lang/another", "./lang"}
	loader = NewFileLoader(paths)
	translations, err = loader.Load("*", "en")
	f.NoError(err)
	f.NotNil(translations)
	f.Equal("bar", translations["en"]["foo"])

	paths = []string{"./lang"}
	loader = NewFileLoader(paths)
	translations, err = loader.Load("another", "en")
	f.NoError(err)
	f.NotNil(translations)
	f.Equal("backagebar", translations["en"]["foo"])
}

func (f *FileLoaderTestSuite) TestLoadNonExistentFile() {
	paths := []string{"./lang"}
	loader := NewFileLoader(paths)
	translations, err := loader.Load("*", "hi")

	f.Error(err)
	f.Nil(translations)
	f.Equal(ErrFileNotExist, err)
}

func (f *FileLoaderTestSuite) TestLoadInvalidJSON() {
	paths := []string{"./lang/invalid"}
	loader := NewFileLoader(paths)
	translations, err := loader.Load("*", "en")

	f.Error(err)
	f.Nil(translations)
}

func (f *FileLoaderTestSuite) TestMergeMaps() {
	dst := map[string]string{
		"foo": "bar",
	}
	src := map[string]string{
		"baz": "backage",
	}
	mergeMaps(dst, src)
	f.Equal(map[string]string{
		"foo": "bar",
		"baz": "backage",
	}, dst)
}
