package translation

import (
	"io/fs"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/foundation/json"
)

type FSLoaderTestSuite struct {
	suite.Suite
	json   foundation.Json
	mockFS fs.FS
}

func TestFSLoaderTestSuite(t *testing.T) {
	suite.Run(t, &FSLoaderTestSuite{})
}

func (f *FSLoaderTestSuite) SetupTest() {
	f.json = json.New()

	f.mockFS = fstest.MapFS{
		"en/test.json":         &fstest.MapFile{Data: []byte(`{"foo": "bar", "baz": {"foo": "bar"}}`)},
		"en/another/test.json": &fstest.MapFile{Data: []byte(`{"foo": "backagebar", "baz": "backagesplash"}`)},
		"another/en/test.json": &fstest.MapFile{Data: []byte(`{"foo": "backagebar", "baz": "backagesplash"}`)},
		"en/invalid/test.json": &fstest.MapFile{Data: []byte(`{"foo": "bar",}`)},
		"cn.json":              &fstest.MapFile{Data: []byte(`{"foo": "bar", "baz": {"foo": "bar"}}`)},
	}
}

func (f *FSLoaderTestSuite) TestLoad() {
	loader := NewFSLoader(f.mockFS, f.json)
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

	translations, err = loader.Load("en", "another/test")
	f.NoError(err)
	f.NotNil(translations)
	f.Equal("backagebar", translations["foo"])
}

func (f *FSLoaderTestSuite) TestLoadNonExistentFile() {
	loader := NewFSLoader(f.mockFS, f.json)
	translations, err := loader.Load("hi", "test")

	f.Error(err)
	f.Nil(translations)
	f.EqualError(err, errors.LangFileNotExist.Error())
}

func (f *FSLoaderTestSuite) TestLoadInvalidJSON() {
	loader := NewFSLoader(f.mockFS, f.json)
	translations, err := loader.Load("en", "invalid/test")

	f.Error(err)
	f.Nil(translations)
}

func Benchmark_FSLoad(b *testing.B) {
	s := new(FSLoaderTestSuite)
	s.SetT(&testing.T{})
	s.SetupTest()
	b.StartTimer()
	b.ResetTimer()

	loader := NewFSLoader(s.mockFS, s.json)
	for i := 0; i < b.N; i++ {
		_, err := loader.Load("en", "test")
		s.NoError(err)
	}

	b.StopTimer()
}
