package foundation

import (
	"github.com/goravel/framework/contracts/binding"
	contractsmodular "github.com/goravel/framework/contracts/modular"
	mocksconfig "github.com/goravel/framework/mocks/config"
)

// testModule has only a name; tests embed it to add capabilities.
type testModule struct{ name string }

func (m *testModule) Name() string { return m.name }

type orderedTestModule struct {
	testModule
	order int
}

func (m *orderedTestModule) Order() int { return m.order }

func moduleNames(modules []contractsmodular.Module) []string {
	var names []string
	for _, module := range modules {
		names = append(names, module.Name())
	}

	return names
}

func (s *ApplicationTestSuite) TestConfigureModules() {
	s.Run("without WithModules an empty registry is bound", func() {
		s.SetupTest()
		s.app.builder = NewApplicationBuilder(s.app)

		s.app.configureModules()

		registry := s.app.MakeModular()
		s.Require().NotNil(registry)
		s.Empty(registry.All())
		s.Empty(s.app.modules)
	})

	s.Run("modules are registered and kept in order", func() {
		s.SetupTest()
		builder := NewApplicationBuilder(s.app)
		builder.modules = func() []contractsmodular.Module {
			return []contractsmodular.Module{
				&orderedTestModule{testModule{"last"}, -1},
				&testModule{"blog"},
				&orderedTestModule{testModule{"shop"}, 5},
			}
		}
		s.app.builder = builder

		s.app.configureModules()

		s.Equal([]string{"blog", "shop", "last"}, moduleNames(s.app.modules))
		s.True(s.app.MakeModular().Has("blog"))
		instance, err := s.app.Make(binding.Modular)
		s.NoError(err)
		s.Same(s.app.MakeModular(), instance)
	})

	s.Run("module directories are resolved, third-party ones under modular.path", func() {
		s.SetupTest()
		mockConfig := mocksconfig.NewConfig(s.T())
		mockConfig.EXPECT().GetString("modular.path", "modules").Return("app/modules").Once()
		s.app.Instance(binding.Config, mockConfig)
		builder := NewApplicationBuilder(s.app)
		builder.modules = func() []contractsmodular.Module {
			return []contractsmodular.Module{&testModule{"blog"}}
		}
		s.app.builder = builder

		s.app.configureModules()

		// Under go test the main module is the framework.
		dir, ok := s.app.MakeModular().Path("blog")
		s.True(ok)
		s.Equal("foundation", dir)
	})

	s.Run("a duplicate name is reported and the modules before it are kept", func() {
		s.SetupTest()
		builder := NewApplicationBuilder(s.app)
		builder.modules = func() []contractsmodular.Module {
			return []contractsmodular.Module{&testModule{"blog"}, &testModule{"blog"}, &testModule{"shop"}}
		}
		s.app.builder = builder

		s.app.configureModules()

		s.Equal([]string{"blog"}, moduleNames(s.app.modules))
	})

	s.Run("a rebuild binds a fresh registry", func() {
		s.SetupTest()
		builder := NewApplicationBuilder(s.app)
		builder.modules = func() []contractsmodular.Module {
			return []contractsmodular.Module{&testModule{"blog"}}
		}
		s.app.builder = builder

		s.app.configureModules()
		first := s.app.MakeModular()
		s.app.configureModules()

		s.NotSame(first, s.app.MakeModular())
		s.Len(s.app.MakeModular().All(), 1)
	})

	s.Run("without Build there is no registry", func() {
		s.SetupTest()

		s.Nil(s.app.MakeModular())
	})
}
