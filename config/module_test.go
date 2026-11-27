package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/goravel/framework/support"
)

func TestModule(t *testing.T) {
	original := support.DontVerifyAppKey
	support.DontVerifyAppKey = true
	defer func() { support.DontVerifyAppKey = original }()

	app := NewApplication("")
	app.Add("modules.blog", map[string]any{
		"path":     "blog",
		"per_page": 10,
		"ratio":    1.5,
		"enabled":  true,
		"timeout":  "10s",
		"since":    "2026-01-02",
		"tags":     []string{"a", "b"},
		"sizes":    []int{1, 2},
		"nested":   map[string]any{"key": "value"},
		"labels":   map[string]string{"title": "Blog"},
	})
	app.Add("path", "root")

	blog := app.Module("blog")

	// Both ways read the same value.
	assert.Equal(t, "blog", blog.GetString("path"))
	assert.Equal(t, app.GetString("modules.blog.path"), blog.GetString("path"))

	assert.Equal(t, "blog", blog.Get("path"))
	assert.Equal(t, 10, blog.GetInt("per_page"))
	assert.Equal(t, int64(10), blog.GetInt64("per_page"))
	assert.Equal(t, 1.5, blog.GetFloat64("ratio"))
	assert.True(t, blog.GetBool("enabled"))
	assert.Equal(t, 10*time.Second, blog.GetDuration("timeout"))
	assert.Equal(t, time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC), blog.GetTime("since"))
	assert.Equal(t, []string{"a", "b"}, blog.GetStringSlice("tags"))
	assert.Equal(t, []int{1, 2}, blog.GetIntSlice("sizes"))
	assert.Equal(t, map[string]any{"key": "value"}, blog.GetStringMap("nested"))
	assert.Equal(t, map[string]string{"title": "Blog"}, blog.GetStringMapString("labels"))
	assert.Equal(t, "value", blog.GetString("nested.key"))

	// Defaults apply to missing paths and to modules without config.
	assert.Equal(t, "default", blog.GetString("missing", "default"))
	assert.Equal(t, "admin", app.Module("admin").GetString("path", "admin"))
	assert.False(t, app.Module("admin").Has("path"))
	assert.Equal(t, map[string]any{}, app.Module("admin").All())

	// An empty path is the module's whole config.
	assert.True(t, blog.Has(""))
	assert.Equal(t, "blog", blog.GetStringMap("")["path"])

	var nested struct {
		Key string `json:"key"`
	}
	assert.NoError(t, blog.UnmarshalKey("nested", &nested))
	assert.Equal(t, "value", nested.Key)

	// Add and Forget are scoped.
	blog.Add("per_page", 20)
	assert.Equal(t, 20, app.GetInt("modules.blog.per_page"))
	assert.Equal(t, "root", app.GetString("path"))

	all := blog.All()
	assert.Equal(t, 20, all["per_page"])
	assert.Equal(t, "blog", all["path"])

	blog.Forget("path")
	assert.False(t, blog.Has("path"))
	assert.False(t, app.Has("modules.blog.path"))
	assert.NotContains(t, blog.All(), "path")
	assert.True(t, app.Has("path"))
}
