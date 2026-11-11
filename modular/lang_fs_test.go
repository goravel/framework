package modular

import (
	"io/fs"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/foundation/json"
	"github.com/goravel/framework/translation"
)

func TestLangFS(t *testing.T) {
	modules := map[string]fs.FS{
		"blog": fstest.MapFS{
			"en/messages.json":    {Data: []byte(`{"title": "Blog"}`)},
			"en/admin/posts.json": {Data: []byte(`{"title": "Posts"}`)},
			"en/overridden.json":  {Data: []byte(`{"title": "Module"}`)},
			"fr/messages.json":    {Data: []byte(`{"title": "Blogue"}`)},
		},
	}
	fallback := fstest.MapFS{
		"en.json":                 {Data: []byte(`{"hello": "Hello"}`)},
		"en/app.json":             {Data: []byte(`{"name": "App"}`)},
		"en/blog/overridden.json": {Data: []byte(`{"title": "App"}`)},
	}

	loader := translation.NewFSLoader(NewLangFS(modules, fallback), json.New())

	tests := []struct {
		locale, group string
		expected      map[string]any
	}{
		{"en", "blog/messages", map[string]any{"title": "Blog"}},
		{"fr", "blog/messages", map[string]any{"title": "Blogue"}},
		{"en", "blog/admin/posts", map[string]any{"title": "Posts"}},
		{"en", "blog/overridden", map[string]any{"title": "App"}},
		{"en", "app", map[string]any{"name": "App"}},
		{"en", "*", map[string]any{"hello": "Hello"}},
	}
	for _, tt := range tests {
		values, err := loader.Load(tt.locale, tt.group)
		require.NoError(t, err, tt.group)
		assert.Equal(t, tt.expected, values, tt.group)
	}

	for _, group := range []string{"blog/missing", "shop/messages", "missing"} {
		_, err := loader.Load("en", group)
		assert.ErrorIs(t, err, errors.LangFileNotExist, group)
	}
	_, err := loader.Load("de", "blog/messages")
	assert.ErrorIs(t, err, errors.LangFileNotExist)
}

func TestLangFS_WithoutFallback(t *testing.T) {
	fsys := NewLangFS(map[string]fs.FS{"blog": fstest.MapFS{"en/messages.json": {Data: []byte(`{}`)}}}, nil)

	_, err := fs.ReadFile(fsys, "en/blog/messages.json")
	assert.NoError(t, err)

	for _, name := range []string{"en.json", "en/app.json", "en/shop/messages.json"} {
		_, err := fs.ReadFile(fsys, name)
		assert.ErrorIs(t, err, fs.ErrNotExist, name)

		_, err = fsys.Open(name)
		assert.ErrorIs(t, err, fs.ErrNotExist, name)
	}

	_, err = fsys.Open("../en.json")
	assert.ErrorIs(t, err, fs.ErrInvalid)
}

func TestLangFS_UnwrapsPrevious(t *testing.T) {
	fallback := fstest.MapFS{"en/app.json": {Data: []byte(`{"name": "App"}`)}}
	first := NewLangFS(map[string]fs.FS{
		"blog": fstest.MapFS{"en/messages.json": {Data: []byte(`{"title": "Old"}`)}},
		"shop": fstest.MapFS{"en/messages.json": {Data: []byte(`{}`)}},
	}, fallback)
	second := NewLangFS(map[string]fs.FS{
		"blog": fstest.MapFS{"en/messages.json": {Data: []byte(`{"title": "New"}`)}},
	}, first)

	assert.Equal(t, fs.FS(fallback), second.(*langFS).fallback)

	data, err := fs.ReadFile(second, "en/blog/messages.json")
	require.NoError(t, err)
	assert.JSONEq(t, `{"title": "New"}`, string(data))

	_, err = fs.ReadFile(second, "en/app.json")
	assert.NoError(t, err)

	_, err = fs.ReadFile(second, "en/shop/messages.json")
	assert.ErrorIs(t, err, fs.ErrNotExist)
}

func TestLangFS_Open(t *testing.T) {
	fsys := NewLangFS(map[string]fs.FS{
		"blog": fstest.MapFS{"en/messages.json": {Data: []byte(`{"title": "Blog"}`)}},
	}, nil)

	opened, err := fsys.Open("en/blog/messages.json")
	require.NoError(t, err)
	info, err := opened.Stat()
	require.NoError(t, err)
	assert.Equal(t, "messages.json", info.Name())
	assert.NoError(t, opened.Close())

	_, err = fsys.Open("en/shop/messages.json")
	assert.ErrorIs(t, err, fs.ErrNotExist)
	_, err = fsys.Open("../en/blog/messages.json")
	assert.ErrorIs(t, err, fs.ErrInvalid)
}
