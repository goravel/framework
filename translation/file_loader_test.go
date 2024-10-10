package translation

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/foundation/json"
	"github.com/goravel/framework/support/env"
	"github.com/goravel/framework/support/file"
)

type FileLoaderTestSuite struct {
	suite.Suite
	json foundation.Json
}

func TestFileLoaderTestSuite(t *testing.T) {
	suite.Run(t, &FileLoaderTestSuite{})
}

func (f *FileLoaderTestSuite) SetupSuite() {
	assert.Nil(f.T(), file.Create("lang/en/test.json", `{"foo": "bar", "baz": {"foo": "bar"}}`))
	assert.Nil(f.T(), file.Create("lang/en/another/test.json", `{"foo": "backagebar", "baz": "backagesplash"}`))
	assert.Nil(f.T(), file.Create("lang/another/en/test.json", `{"foo": "backagebar", "baz": "backagesplash"}`))
	assert.Nil(f.T(), file.Create("lang/en/invalid/test.json", `{"foo": "bar",}`))
	assert.Nil(f.T(), file.Create("lang/cn.json", `{"foo": "bar", "baz": {"foo": "bar"}}`))
	restrictedFilePath := "lang/en/restricted/test.json"
	assert.Nil(f.T(), file.Create(restrictedFilePath, `{"foo": "restricted"}`))
	assert.Nil(f.T(), os.Chmod(restrictedFilePath, 0000))
}

func (f *FileLoaderTestSuite) TearDownSuite() {
	f.Nil(file.Remove("lang"))
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
	f.EqualError(err, errors.LangFileNotExist.Error())
}

func (f *FileLoaderTestSuite) TestLoadInvalidJSON() {
	paths := []string{"./lang"}
	loader := NewFileLoader(paths, f.json)
	translations, err := loader.Load("en", "invalid/test")

	f.Error(err)
	f.Nil(translations)
}

func Benchmark_Load(b *testing.B) {
	s := new(FileLoaderTestSuite)
	s.SetT(&testing.T{})
	s.SetupSuite()
	s.SetupTest()
	b.StartTimer()
	b.ResetTimer()

	paths := []string{"./lang"}
	loader := NewFileLoader(paths, s.json)
	for i := 0; i < b.N; i++ {
		_, err := loader.Load("en", "test")
		s.NoError(err)
	}

	b.StopTimer()
	s.TearDownSuite()
}
