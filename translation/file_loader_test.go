package translation

import (
	"github.com/goravel/framework/support/file"
	"github.com/stretchr/testify/suite"
	"testing"
)

type FileLoaderTestSuite struct {
	suite.Suite
}

func TestFileLoaderTestSuite(t *testing.T) {
	_ = file.Create("lang/en.json", `{"foo": "bar"}`)
	_ = file.Create("lang/another/en.json", `{"foo": "backagebar", "baz": "backagesplash"}`)
	_ = file.Create("lang/invalid/en.json", `{"foo": "bar",}`)
	suite.Run(t, &FileLoaderTestSuite{})
	_ = file.Remove("lang")
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
