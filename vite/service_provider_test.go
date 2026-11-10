package vite

import (
	"html/template"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/binding"
	contractsconsole "github.com/goravel/framework/contracts/console"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/errors"
	mocksconfig "github.com/goravel/framework/mocks/config"
	mocksfoundation "github.com/goravel/framework/mocks/foundation"
	mocksview "github.com/goravel/framework/mocks/view"
	"github.com/goravel/framework/support/color"
)

type ServiceProviderTestSuite struct {
	suite.Suite
	app      *mocksfoundation.Application
	provider *ServiceProvider
}

func TestServiceProviderTestSuite(t *testing.T) {
	suite.Run(t, new(ServiceProviderTestSuite))
}

func (s *ServiceProviderTestSuite) SetupTest() {
	s.app = mocksfoundation.NewApplication(s.T())
	s.provider = &ServiceProvider{}
}

func (s *ServiceProviderTestSuite) TestRelationship() {
	relationship := s.provider.Relationship()

	s.Equal([]string{binding.Vite}, relationship.Bindings)
	s.Equal([]string{binding.Config, binding.View}, relationship.Dependencies)
	s.Empty(relationship.ProvideFor)
}

func (s *ServiceProviderTestSuite) TestRegister() {
	config := mocksconfig.NewConfig(s.T())
	config.EXPECT().GetString("vite.hot_file", DefaultHotFile()).Return(DefaultHotFile()).Once()
	config.EXPECT().GetString("vite.build_directory", DefaultBuildDirectory()).Return(DefaultBuildDirectory()).Once()
	config.EXPECT().GetString("vite.manifest", DefaultManifestFilename).Return(DefaultManifestFilename).Once()

	s.app.EXPECT().Singleton(binding.Vite, mock.Anything).Run(func(_ any, callback func(app foundation.Application) (any, error)) {
		s.app.EXPECT().MakeConfig().Return(config).Once()
		instance, err := callback(s.app)
		s.NoError(err)
		s.IsType(&Vite{}, instance)

		s.app.EXPECT().MakeConfig().Return(nil).Once()
		instance, err = callback(s.app)
		s.Nil(instance)
		s.EqualError(err, errors.ConfigFacadeNotSet.SetModule(errors.ModuleVite).Error())
	}).Once()

	s.provider.Register(s.app)
}

func (s *ServiceProviderTestSuite) TestBoot() {
	s.Run("registers the template functions", func() {
		view := mocksview.NewView(s.T())
		s.app.EXPECT().MakeView().Return(view).Once()
		s.app.EXPECT().MakeVite().Return(NewVite(nil)).Once()
		s.app.EXPECT().MakeConfig().Return(mocksconfig.NewConfig(s.T())).Once()
		view.EXPECT().AddFuncs(mock.Anything).Run(func(funcs template.FuncMap) {
			s.Len(funcs, 3)
			s.Contains(funcs, "vite")
			s.Contains(funcs, "viteAsset")
			s.Contains(funcs, "viteReactRefresh")
		}).Once()
		s.expectInputsCommand()

		s.provider.Boot(s.app)
	})

	s.Run("without the view registry nothing is registered and the cause is reported", func() {
		s.app.EXPECT().MakeView().Return(nil).Once()

		output := color.CaptureOutput(func(io.Writer) {
			s.provider.Boot(s.app)
		})
		s.Contains(output, errors.ViewFacadeNotSet.SetModule(errors.ModuleVite).Error())
	})

	s.Run("without the vite binding the functions are skipped and the cause is reported", func() {
		view := mocksview.NewView(s.T())
		s.app.EXPECT().MakeView().Return(view).Once()
		s.app.EXPECT().MakeVite().Return(nil).Once()
		s.app.EXPECT().MakeConfig().Return(mocksconfig.NewConfig(s.T())).Once()
		s.expectInputsCommand()

		output := color.CaptureOutput(func(io.Writer) {
			s.provider.Boot(s.app)
		})
		s.Contains(output, errors.ViteFacadeNotSet.Error())
	})
}

// expectInputsCommand expects Boot to register vite:inputs, the only command of the module.
func (s *ServiceProviderTestSuite) expectInputsCommand() {
	s.app.EXPECT().Commands(mock.Anything).Run(func(commands []contractsconsole.Command) {
		s.Len(commands, 1)
		s.Equal("vite:inputs", commands[0].Signature())
	}).Once()
}

func TestFuncsParse(t *testing.T) {
	// The registered functions must have signatures html/template accepts.
	_, err := template.New("x").Funcs(Funcs(NewVite(nil))).Parse(`{{ vite "a" }}{{ viteAsset "a" }}{{ viteReactRefresh }}`)
	require.NoError(t, err)
	assert.NotPanics(t, func() { Funcs(NewVite(nil)) })
}
