package console_test

import (
	"encoding/json"
	"html/template"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	mocksconfig "github.com/goravel/framework/mocks/config"
	mocksconsole "github.com/goravel/framework/mocks/console"
	mocksview "github.com/goravel/framework/mocks/view"
	"github.com/goravel/framework/support"
	"github.com/goravel/framework/vite"
	"github.com/goravel/framework/vite/console"
)

// TestInputsCommand_LegacyHotFile pins that a "public/hot" hot file, restored from
// its absolute path as goravel/example does, is still reported relative to the
// project root.
func TestInputsCommand_LegacyHotFile(t *testing.T) {
	relativePath := support.RelativePath
	t.Cleanup(func() { support.RelativePath = relativePath })
	support.RelativePath = t.TempDir()

	config := mocksconfig.NewConfig(t)
	config.EXPECT().GetString("vite.hot_file", vite.DefaultHotFile()).Return("public/hot").Once()
	config.EXPECT().GetString("vite.build_directory", vite.DefaultBuildDirectory()).Return(vite.DefaultBuildDirectory()).Once()
	config.EXPECT().GetString("vite.manifest", vite.DefaultManifestFilename).Return(vite.DefaultManifestFilename).Once()
	config.EXPECT().GetStringSlice("vite.inputs", []string{}).Return(nil).Once()
	config.EXPECT().GetStringSlice("vite.assets", []string{}).Return(nil).Once()

	views := mocksview.NewView(t)
	views.EXPECT().Funcs().Return(template.FuncMap{}).Once()
	views.EXPECT().RegisteredViews().Return(nil).Once()
	views.EXPECT().RegisteredViewFS().Return(nil).Once()

	var printed string
	ctx := mocksconsole.NewContext(t)
	ctx.EXPECT().Line(mock.Anything).Run(func(line string) { printed = line }).Once()

	shared := vite.NewVite(config)
	shared.UseHotFile(shared.HotFile())
	require.NoError(t, console.NewInputsCommand(views, shared, config).Handle(ctx))

	var inputs console.Inputs
	require.NoError(t, json.Unmarshal([]byte(printed), &inputs))
	assert.Equal(t, "public/hot", inputs.Paths.HotFile)
}
