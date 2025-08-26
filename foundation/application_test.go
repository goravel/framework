package foundation

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/auth"
	"github.com/goravel/framework/cache"
	frameworkconfig "github.com/goravel/framework/config"
	"github.com/goravel/framework/console"
	"github.com/goravel/framework/contracts/binding"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/crypt"
	"github.com/goravel/framework/event"
	"github.com/goravel/framework/filesystem"
	"github.com/goravel/framework/grpc"
	"github.com/goravel/framework/hash"
	"github.com/goravel/framework/http"
	frameworklog "github.com/goravel/framework/log"
	"github.com/goravel/framework/mail"
	mockscache "github.com/goravel/framework/mocks/cache"
	mocksconfig "github.com/goravel/framework/mocks/config"
	mocksconsole "github.com/goravel/framework/mocks/console"
	mocksdb "github.com/goravel/framework/mocks/database/db"
	mocksorm "github.com/goravel/framework/mocks/database/orm"
	mocksfoundation "github.com/goravel/framework/mocks/foundation"
	mockslog "github.com/goravel/framework/mocks/log"
	mocksqueue "github.com/goravel/framework/mocks/queue"
	mocksroute "github.com/goravel/framework/mocks/route"
	"github.com/goravel/framework/queue"
	"github.com/goravel/framework/schedule"
	frameworksession "github.com/goravel/framework/session"
	"github.com/goravel/framework/support"
	"github.com/goravel/framework/support/file"
	frameworktranslation "github.com/goravel/framework/translation"
	"github.com/goravel/framework/validation"
)

type ApplicationTestSuite struct {
	suite.Suite
	app *Application
}

func TestApplicationTestSuite(t *testing.T) {
	assert.Nil(t, file.PutContent(support.EnvFilePath, "APP_KEY=12345678901234567890123456789012"))

	suite.Run(t, new(ApplicationTestSuite))

	assert.Nil(t, file.Remove(support.EnvFilePath))
}

func (s *ApplicationTestSuite) SetupTest() {
	s.app = &Application{
		Container:     NewContainer(),
		publishes:     make(map[string]map[string]string),
		publishGroups: make(map[string]map[string]string),
	}
	App = s.app
}

func (s *ApplicationTestSuite) TestPath() {
	s.Equal(filepath.Join(support.RootPath, "app", "goravel.go"), s.app.Path("goravel.go"))
}

func (s *ApplicationTestSuite) TestBasePath() {
	s.Equal(filepath.Join(support.RootPath, "goravel.go"), s.app.BasePath("goravel.go"))
}

func (s *ApplicationTestSuite) TestConfigPath() {
	s.Equal(filepath.Join(support.RootPath, "config", "goravel.go"), s.app.ConfigPath("goravel.go"))
}

func (s *ApplicationTestSuite) TestDatabasePath() {
	s.Equal(filepath.Join(support.RootPath, "database", "goravel.go"), s.app.DatabasePath("goravel.go"))
}

func (s *ApplicationTestSuite) TestGetInstalledFacades() {
	mockConfig := mocksconfig.NewConfig(s.T())
	mockConfig.EXPECT().Get("app.providers").Return([]foundation.ServiceProvider{
		&auth.ServiceProvider{},
	}).Once()

	s.app.Singleton(binding.Config, func(app foundation.Application) (any, error) {
		return mockConfig, nil
	})

	s.ElementsMatch([]string{"Auth", "Gate"}, s.app.getInstalledFacades())
}

func (s *ApplicationTestSuite) TestStoragePath() {
	s.Equal(filepath.Join(support.RootPath, "storage", "goravel.go"), s.app.StoragePath("goravel.go"))
}

func (s *ApplicationTestSuite) TestResourcePath() {
	s.Equal(filepath.Join(support.RootPath, "resources", "goravel.go"), s.app.ResourcePath("goravel.go"))
}

func (s *ApplicationTestSuite) TestLangPath() {
	mockConfig := mocksconfig.NewConfig(s.T())
	mockConfig.EXPECT().GetString("app.lang_path", "lang").Return("test").Once()

	s.app.Singleton(binding.Config, func(app foundation.Application) (any, error) {
		return mockConfig, nil
	})

	s.Equal(filepath.Join(support.RootPath, "test", "goravel.go"), s.app.LangPath("goravel.go"))
}

func (s *ApplicationTestSuite) TestPublicPath() {
	s.Equal(filepath.Join(support.RootPath, "public", "goravel.go"), s.app.PublicPath("goravel.go"))
}

func (s *ApplicationTestSuite) TestExecutablePath() {
	path, err := os.Getwd()
	s.NoError(err)

	executable := s.app.ExecutablePath()
	s.NotEmpty(executable)
	executable2 := s.app.ExecutablePath("test")
	s.Equal(filepath.Join(path, "test"), executable2)
	executable3 := s.app.ExecutablePath("test", "test2/test3")
	s.Equal(filepath.Join(path, "test", "test2/test3"), executable3)
}

func (s *ApplicationTestSuite) TestPublishes() {
	s.app.Publishes("github.com/goravel/sms", map[string]string{
		"config.go": "config.go",
	})
	s.Equal(1, len(s.app.publishes["github.com/goravel/sms"]))
	s.Equal(0, len(s.app.publishGroups))

	s.app.Publishes("github.com/goravel/sms", map[string]string{
		"config.go":  "config1.go",
		"config1.go": "config1.go",
	}, "public", "private")
	s.Equal(2, len(s.app.publishes["github.com/goravel/sms"]))
	s.Equal("config1.go", s.app.publishes["github.com/goravel/sms"]["config.go"])
	s.Equal(2, len(s.app.publishGroups["public"]))
	s.Equal("config1.go", s.app.publishGroups["public"]["config.go"])
	s.Equal(2, len(s.app.publishGroups["private"]))
}

func (s *ApplicationTestSuite) TestAddPublishGroup() {
	s.app.addPublishGroup("public", map[string]string{
		"config.go": "config.go",
	})
	s.Equal(1, len(s.app.publishGroups["public"]))

	s.app.addPublishGroup("public", map[string]string{
		"config.go":  "config1.go",
		"config1.go": "config1.go",
	})
	s.Equal(2, len(s.app.publishGroups["public"]))
	s.Equal("config1.go", s.app.publishGroups["public"]["config.go"])
}

func (s *ApplicationTestSuite) TestMakeArtisan() {
	serviceProvider := &console.ServiceProvider{}
	serviceProvider.Register(s.app)

	s.NotNil(s.app.MakeArtisan())
}

func (s *ApplicationTestSuite) TestMakeAuth() {
	mockConfig := mocksconfig.NewConfig(s.T())
	mockConfig.EXPECT().GetString("auth.defaults.guard").Return("user").Once()
	mockConfig.EXPECT().GetString("auth.guards.user.driver").Return("jwt").Once()
	mockConfig.EXPECT().GetString("auth.guards.user.provider").Return("user").Once()
	mockConfig.EXPECT().GetString("auth.providers.user.driver").Return("orm").Once()
	mockConfig.EXPECT().GetString("auth.guards.user.secret").Return("secret").Once()
	mockConfig.EXPECT().GetInt("auth.guards.user.ttl").Return(100).Once()
	mockConfig.EXPECT().GetInt("auth.guards.user.refresh_ttl").Return(100).Once()

	s.app.Singleton(binding.Config, func(app foundation.Application) (any, error) {
		return mockConfig, nil
	})
	s.app.Singleton(binding.Cache, func(app foundation.Application) (any, error) {
		return &mockscache.Cache{}, nil
	})
	s.app.Singleton(binding.Orm, func(app foundation.Application) (any, error) {
		return &mocksorm.Orm{}, nil
	})
	s.app.Singleton(binding.Log, func(app foundation.Application) (any, error) {
		return &mockslog.Log{}, nil
	})

	serviceProvider := &auth.ServiceProvider{}
	serviceProvider.Register(s.app)
	serviceProvider.Boot(s.app)

	s.NotNil(s.app.MakeAuth(http.Background()))
}

func (s *ApplicationTestSuite) TestMakeCache() {
	mockConfig := mocksconfig.NewConfig(s.T())
	mockConfig.EXPECT().GetString("cache.default").Return("memory").Once()
	mockConfig.EXPECT().GetString("cache.stores.memory.driver").Return("memory").Once()
	mockConfig.EXPECT().GetString("cache.prefix").Return("goravel").Once()

	s.app.Singleton(binding.Config, func(app foundation.Application) (any, error) {
		return mockConfig, nil
	})
	s.app.Singleton(binding.Log, func(app foundation.Application) (any, error) {
		return &mockslog.Log{}, nil
	})

	serviceProvider := &cache.ServiceProvider{}
	serviceProvider.Register(s.app)

	s.NotNil(s.app.MakeCache())
}

func (s *ApplicationTestSuite) TestMakeConfig() {
	serviceProvider := &frameworkconfig.ServiceProvider{}
	serviceProvider.Register(s.app)

	s.NotNil(s.app.MakeConfig())
}

func (s *ApplicationTestSuite) TestMakeCrypt() {
	mockConfig := mocksconfig.NewConfig(s.T())
	mockConfig.EXPECT().GetString("app.key").Return("12345678901234567890123456789012").Once()

	s.app.Singleton(binding.Config, func(app foundation.Application) (any, error) {
		return mockConfig, nil
	})
	s.app.SetJson(mocksfoundation.NewJson(s.T()))

	serviceProvider := &crypt.ServiceProvider{}
	serviceProvider.Register(s.app)

	s.NotNil(s.app.MakeCrypt())
}

func (s *ApplicationTestSuite) TestMakeEvent() {
	s.app.Singleton(binding.Queue, func(app foundation.Application) (any, error) {
		return &mocksqueue.Queue{}, nil
	})

	serviceProvider := &event.ServiceProvider{}
	serviceProvider.Register(s.app)

	s.NotNil(s.app.MakeEvent())
}

func (s *ApplicationTestSuite) TestMakeGate() {
	serviceProvider := &auth.ServiceProvider{}
	serviceProvider.Register(s.app)

	s.NotNil(s.app.MakeGate())
}

func (s *ApplicationTestSuite) TestMakeGrpc() {
	s.app.Singleton(binding.Config, func(app foundation.Application) (any, error) {
		return &mocksconfig.Config{}, nil
	})

	serviceProvider := &grpc.ServiceProvider{}
	serviceProvider.Register(s.app)

	s.NotNil(s.app.MakeGrpc())
}

func (s *ApplicationTestSuite) TestMakeHash() {
	mockConfig := mocksconfig.NewConfig(s.T())
	mockConfig.EXPECT().GetString("hashing.driver", "argon2id").Return("argon2id").Once()
	mockConfig.EXPECT().GetInt("hashing.argon2id.time", 4).Return(4).Once()
	mockConfig.EXPECT().GetInt("hashing.argon2id.memory", 65536).Return(65536).Once()
	mockConfig.EXPECT().GetInt("hashing.argon2id.threads", 1).Return(1).Once()

	s.app.Singleton(binding.Config, func(app foundation.Application) (any, error) {
		return mockConfig, nil
	})

	serviceProvider := &hash.ServiceProvider{}
	serviceProvider.Register(s.app)

	s.NotNil(s.app.MakeHash())
}

func (s *ApplicationTestSuite) TestMakeLang() {
	mockConfig := mocksconfig.NewConfig(s.T())
	mockConfig.EXPECT().GetString("app.locale").Return("en").Once()
	mockConfig.EXPECT().GetString("app.fallback_locale").Return("en").Once()
	mockConfig.EXPECT().GetString("app.lang_path", "lang").Return("lang").Once()
	mockConfig.EXPECT().Get("app.lang_fs").Return(nil).Once()

	s.app.Singleton(binding.Config, func(app foundation.Application) (any, error) {
		return mockConfig, nil
	})
	s.app.Singleton(binding.Log, func(app foundation.Application) (any, error) {
		return &mockslog.Log{}, nil
	})

	serviceProvider := &frameworktranslation.ServiceProvider{}
	serviceProvider.Register(s.app)
	ctx := http.Background()

	s.NotNil(s.app.MakeLang(ctx))
}

func (s *ApplicationTestSuite) TestMakeLog() {
	mockConfig := mocksconfig.NewConfig(s.T())
	s.app.Singleton(binding.Config, func(app foundation.Application) (any, error) {
		return mockConfig, nil
	})

	mockConfig.EXPECT().GetString("logging.default").Return("").Once()

	s.app.SetJson(mocksfoundation.NewJson(s.T()))

	serviceProvider := &frameworklog.ServiceProvider{}
	serviceProvider.Register(s.app)

	s.NotNil(s.app.MakeLog())
}

func (s *ApplicationTestSuite) TestMakeMail() {
	mockConfig := mocksconfig.NewConfig(s.T())
	s.app.Singleton(binding.Config, func(app foundation.Application) (any, error) {
		return mockConfig, nil
	})
	s.app.Singleton(binding.Queue, func(app foundation.Application) (any, error) {
		return &mocksqueue.Queue{}, nil
	})

	mockConfig.EXPECT().GetString("mail.template.default", "html").Return("html").Once()
	mockConfig.EXPECT().GetString("mail.template.engines.html.driver", "html").
		Return("html").Once()
	mockConfig.EXPECT().GetString("mail.template.engines.html.path", "resources/views/mail").
		Return("resources/views/mail").Once()

	serviceProvider := &mail.ServiceProvider{}
	serviceProvider.Register(s.app)

	s.NotNil(s.app.MakeMail())
}

func (s *ApplicationTestSuite) TestMakeQueue() {
	mockConfig := mocksconfig.NewConfig(s.T())
	s.app.Singleton(binding.Config, func(app foundation.Application) (any, error) {
		return mockConfig, nil
	})
	s.app.Singleton(binding.DB, func(app foundation.Application) (any, error) {
		return &mocksdb.DB{}, nil
	})
	s.app.Singleton(binding.Log, func(app foundation.Application) (any, error) {
		return &mockslog.Log{}, nil
	})

	serviceProvider := &queue.ServiceProvider{}
	serviceProvider.Register(s.app)

	mockConfig.EXPECT().GetString("queue.default").Return("redis").Once()
	mockConfig.EXPECT().GetString("queue.connections.redis.queue", "default").Return("default").Once()
	mockConfig.EXPECT().GetInt("queue.connections.redis.concurrent", 1).Return(2).Once()
	mockConfig.EXPECT().GetString("app.name", "goravel").Return("goravel").Once()
	mockConfig.EXPECT().GetBool("app.debug").Return(true).Once()
	mockConfig.EXPECT().GetString("queue.failed.database").Return("mysql").Once()
	mockConfig.EXPECT().GetString("queue.failed.table").Return("failed_jobs").Once()

	s.NotNil(s.app.MakeQueue())
}

func (s *ApplicationTestSuite) TestMakeRateLimiter() {
	serviceProvider := &http.ServiceProvider{}
	serviceProvider.Register(s.app)

	s.NotNil(s.app.MakeRateLimiter())
}

func (s *ApplicationTestSuite) TestMakeRoute() {
	mockConfig := mocksconfig.NewConfig(s.T())

	s.app.Singleton(binding.Config, func(app foundation.Application) (any, error) {
		return mockConfig, nil
	})

	s.app.Singleton(binding.Route, func(app foundation.Application) (any, error) {
		return &mocksroute.Route{}, nil
	})

	s.NotNil(s.app.MakeRoute())
}

func (s *ApplicationTestSuite) TestMakeSchedule() {
	mockConfig := mocksconfig.NewConfig(s.T())
	mockConfig.EXPECT().GetBool("app.debug").Return(false).Once()

	s.app.Singleton(binding.Config, func(app foundation.Application) (any, error) {
		return mockConfig, nil
	})
	s.app.Singleton(binding.Artisan, func(app foundation.Application) (any, error) {
		return &mocksconsole.Artisan{}, nil
	})
	s.app.Singleton(binding.Log, func(app foundation.Application) (any, error) {
		return &mockslog.Log{}, nil
	})
	s.app.Singleton(binding.Cache, func(app foundation.Application) (any, error) {
		return &mockscache.Cache{}, nil
	})

	serviceProvider := &schedule.ServiceProvider{}
	serviceProvider.Register(s.app)

	s.NotNil(s.app.MakeSchedule())
}

func (s *ApplicationTestSuite) TestMakeSession() {
	mockConfig := mocksconfig.NewConfig(s.T())
	mockConfig.EXPECT().GetString("session.default", "file").Return("file").Once()
	mockConfig.EXPECT().GetString("session.drivers.file.driver").Return("file").Once()
	mockConfig.EXPECT().GetInt("session.lifetime", 120).Return(120).Once()
	mockConfig.EXPECT().GetInt("session.gc_interval", 30).Return(30).Once()
	mockConfig.EXPECT().GetString("session.files").Return("framework/sessions").Once()
	mockConfig.EXPECT().GetString("session.cookie").Return("goravel_session").Once()

	s.app.Singleton(binding.Config, func(app foundation.Application) (any, error) {
		return mockConfig, nil
	})
	s.app.SetJson(mocksfoundation.NewJson(s.T()))

	serviceProvider := &frameworksession.ServiceProvider{}
	// error
	s.Nil(s.app.MakeSession())

	serviceProvider.Register(s.app)
	s.NotNil(s.app.MakeSession())
}

func (s *ApplicationTestSuite) TestMakeStorage() {
	mockConfig := mocksconfig.NewConfig(s.T())
	mockConfig.EXPECT().GetString("filesystems.default").Return("local").Once()
	mockConfig.EXPECT().GetString("filesystems.disks.local.driver").Return("local").Once()
	mockConfig.EXPECT().GetString("filesystems.disks.local.root").Return("").Once()
	mockConfig.EXPECT().GetString("filesystems.disks.local.url").Return("").Once()

	s.app.Singleton(binding.Config, func(app foundation.Application) (any, error) {
		return mockConfig, nil
	})

	serviceProvider := &filesystem.ServiceProvider{}
	serviceProvider.Register(s.app)

	s.NotNil(s.app.MakeStorage())
}

func (s *ApplicationTestSuite) TestMakeValidation() {
	serviceProvider := &validation.ServiceProvider{}
	serviceProvider.Register(s.app)

	s.NotNil(s.app.MakeValidation())
}

func TestSortConfiguredServiceProviders(t *testing.T) {
	testCases := []struct {
		name      string
		providers []foundation.ServiceProvider
		expected  []foundation.ServiceProvider
		checkTopo bool
	}{
		{
			name: "not found basic dependency, should be sorted correctly",
			providers: []foundation.ServiceProvider{
				&CServiceProvider{},
				&BServiceProvider{},
			},
			expected: []foundation.ServiceProvider{
				&BServiceProvider{},
				&CServiceProvider{},
			},
		},
		{
			name: "BasicSorting",
			providers: []foundation.ServiceProvider{
				&BServiceProvider{},
				&CServiceProvider{},
				&AServiceProvider{},
			},
			expected: []foundation.ServiceProvider{
				&AServiceProvider{},
				&BServiceProvider{},
				&CServiceProvider{},
			},
			checkTopo: true,
		},
		{
			name: "SingleProvider",
			providers: []foundation.ServiceProvider{
				&BasicServiceProvider{},
			},
			expected: []foundation.ServiceProvider{
				&BasicServiceProvider{},
			},
			checkTopo: true,
		},
		{
			name:      "EmptyProviders",
			providers: []foundation.ServiceProvider{},
			expected:  []foundation.ServiceProvider{},
			checkTopo: true,
		},
		{
			name: "ProvideForRelationship",
			providers: []foundation.ServiceProvider{
				&ProvideForBServiceProvider{},
				&ProvideForAServiceProvider{},
			},
			expected: []foundation.ServiceProvider{
				&ProvideForAServiceProvider{},
				&ProvideForBServiceProvider{},
			},
			checkTopo: true,
		},
		{
			name: "SingleProviderWithMock",
			providers: []foundation.ServiceProvider{
				&MockProviderE{},
			},
			expected: []foundation.ServiceProvider{
				&MockProviderE{},
			},
		},
		{
			name: "EmptyDependencies",
			providers: []foundation.ServiceProvider{
				&MockProviderC{},
				&EmptyDependenciesProvider{},
			},
			expected: []foundation.ServiceProvider{
				&EmptyDependenciesProvider{},
				&MockProviderC{},
			},
			checkTopo: true,
		},
		{
			name: "EmptyProvideFor",
			providers: []foundation.ServiceProvider{
				&EmptyProvideForProvider{},
				&MockProviderA{},
			},
			expected: []foundation.ServiceProvider{
				&MockProviderA{},
				&EmptyProvideForProvider{},
			},
			checkTopo: true,
		},
		{
			name: "AllEmptyMethods",
			providers: []foundation.ServiceProvider{
				&AllEmptyProvider{},
				&MockProviderE{},
			},
			expected: []foundation.ServiceProvider{
				&MockProviderE{},
				&AllEmptyProvider{},
			},
			checkTopo: true,
		},
		{
			name: "MixedEmptyAndNonEmpty",
			providers: []foundation.ServiceProvider{
				&AllEmptyProvider{},
				&MockProviderC{},
				&EmptyDependenciesProvider{},
			},
			expected: []foundation.ServiceProvider{
				&EmptyDependenciesProvider{},
				&MockProviderC{},
				&AllEmptyProvider{},
			},
			checkTopo: true,
		},
		{
			name: "EmptyBindingsWithDependencies",
			providers: []foundation.ServiceProvider{
				&EmptyBindingsWithDependenciesProvider{},
				&MockProviderC{},
			},
			expected: []foundation.ServiceProvider{
				&MockProviderC{},
				&EmptyBindingsWithDependenciesProvider{},
			},
			checkTopo: true,
		},
		{
			name: "EmptyBindingsWithProvideFor",
			providers: []foundation.ServiceProvider{
				&MockProviderA{},
				&EmptyBindingsWithProvideForProvider{},
			},
			expected: []foundation.ServiceProvider{
				&EmptyBindingsWithProvideForProvider{},
				&MockProviderA{},
			},
			checkTopo: true,
		},
		{
			name: "EmptyBindingsWithBothDependenciesAndProvideFor",
			providers: []foundation.ServiceProvider{
				&EmptyBindingsWithBothProvider{},
				&MockProviderA{},
			},
			expected: []foundation.ServiceProvider{
				&EmptyBindingsWithBothProvider{},
				&MockProviderA{},
			},
		},
		{
			name: "MultipleEmptyBindingsProviders",
			providers: []foundation.ServiceProvider{
				&EmptyBindingsWithDependenciesProvider{},
				&EmptyBindingsWithProvideForProvider{},
				&MockProviderE{},
			},
			expected: []foundation.ServiceProvider{
				&EmptyBindingsWithDependenciesProvider{},
				&EmptyBindingsWithProvideForProvider{},
				&MockProviderE{},
			},
		},
		{
			name: "ComplexEmptyBindingsScenario",
			providers: []foundation.ServiceProvider{
				&EmptyBindingsWithDependenciesProvider{},
				&EmptyBindingsWithProvideForProvider{},
				&EmptyBindingsWithBothProvider{},
				&MockProviderE{},
				&AllEmptyProvider{},
			},
			expected: []foundation.ServiceProvider{
				&AllEmptyProvider{},
				&EmptyBindingsWithDependenciesProvider{},
				&EmptyBindingsWithProvideForProvider{},
				&MockProviderE{},
				&EmptyBindingsWithBothProvider{},
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			result := sortConfiguredServiceProviders(tt.providers)

			if tt.checkTopo {
				assert.Equal(t, tt.expected, result)
				assert.True(t, isTopologicalOrder(tt.providers, result), "Result is not a valid topological order")
			} else {
				assert.ElementsMatch(t, tt.expected, result)
			}
		})
	}
}

func TestSortConfiguredServiceProvidersWithCircularDependency(t *testing.T) {
	providers := []foundation.ServiceProvider{
		&ComplexProviderA{},
		&ComplexProviderB{},
		&ComplexProviderC{},
	}

	// 捕获 panic 并验证错误消息
	defer func() {
		if r := recover(); r != nil {
			err, ok := r.(error)
			assert.True(t, ok, "Expected panic to be an error")
			assert.Contains(t, err.Error(), "circular dependency detected between providers:")
			assert.Contains(t, err.Error(), "*foundation.ComplexProviderA")
			assert.Contains(t, err.Error(), "*foundation.ComplexProviderB")
			assert.Contains(t, err.Error(), "*foundation.ComplexProviderC")
		} else {
			t.Error("Expected panic but none occurred")
		}
	}()

	sortConfiguredServiceProviders(providers)
}

func TestSortConfiguredServiceProvidersWithEmptyBindingsCircularDependency(t *testing.T) {
	providers := []foundation.ServiceProvider{
		&EmptyBindingsCircularAProvider{},
		&EmptyBindingsCircularBProvider{},
		&CircularBindingAProvider{},
		&CircularBindingBProvider{},
	}

	defer func() {
		if r := recover(); r != nil {
			err, ok := r.(error)
			assert.True(t, ok, "Expected panic to be an error")
			assert.Contains(t, err.Error(), "circular dependency detected between providers:")
		} else {
			t.Error("Expected panic but none occurred")
		}
	}()

	sortConfiguredServiceProviders(providers)
}

func Test_detectCycle(t *testing.T) {
	testCases := []struct {
		name              string
		graph             map[string][]string
		bindingToProvider map[string]foundation.ServiceProvider
		expected          []string
	}{
		{
			name: "SimpleCycle",
			graph: map[string][]string{
				"A": {"B"},
				"B": {"A"},
			},
			bindingToProvider: map[string]foundation.ServiceProvider{
				"A": &MockProviderA{},
				"B": &MockProviderB{},
			},
			expected: []string{"*foundation.MockProviderA", "*foundation.MockProviderB", "*foundation.MockProviderA"},
		},
		{
			name: "ComplexCycle",
			graph: map[string][]string{
				"A": {"B"},
				"B": {"C"},
				"C": {"A"},
			},
			bindingToProvider: map[string]foundation.ServiceProvider{
				"A": &MockProviderA{},
				"B": &MockProviderB{},
				"C": &MockProviderC{},
			},
			expected: []string{"*foundation.MockProviderA", "*foundation.MockProviderB", "*foundation.MockProviderC", "*foundation.MockProviderA"},
		},
		{
			name:              "SelfLoop",
			graph:             map[string][]string{"A": {"A"}},
			bindingToProvider: map[string]foundation.ServiceProvider{"A": &MockProviderA{}},
			expected:          []string{"*foundation.MockProviderA", "*foundation.MockProviderA"},
		},
		{
			name:              "NoCycle",
			graph:             map[string][]string{"A": {"B"}, "B": {"C"}, "C": {}},
			bindingToProvider: map[string]foundation.ServiceProvider{"A": &MockProviderA{}, "B": &MockProviderB{}, "C": &MockProviderC{}},
			expected:          nil,
		},
		{
			name:              "DisconnectedComponents",
			graph:             map[string][]string{"A": {"B"}, "B": {"A"}, "C": {"D"}, "D": {}},
			bindingToProvider: map[string]foundation.ServiceProvider{"A": &MockProviderA{}, "B": &MockProviderB{}, "C": &MockProviderC{}, "D": &MockProviderD{}},
			expected:          []string{"*foundation.MockProviderA", "*foundation.MockProviderB", "*foundation.MockProviderA"},
		},
		{
			name:              "EmptyGraph",
			graph:             map[string][]string{},
			bindingToProvider: map[string]foundation.ServiceProvider{},
			expected:          nil,
		},
		{
			name:              "SingleNode",
			graph:             map[string][]string{"A": {}},
			bindingToProvider: map[string]foundation.ServiceProvider{"A": &MockProviderA{}},
			expected:          nil,
		},
		{
			name:              "MultipleCycles",
			graph:             map[string][]string{"A": {"B"}, "B": {"A"}, "C": {"D"}, "D": {"C"}},
			bindingToProvider: map[string]foundation.ServiceProvider{"A": &MockProviderA{}, "B": &MockProviderB{}, "C": &MockProviderC{}, "D": &MockProviderD{}},
			expected:          []string{"*foundation.MockProviderA", "*foundation.MockProviderB", "*foundation.MockProviderA"},
		},
		{
			name:              "ComplexPath",
			graph:             map[string][]string{"A": {"B"}, "B": {"C"}, "C": {"D"}, "D": {"B"}},
			bindingToProvider: map[string]foundation.ServiceProvider{"A": &MockProviderA{}, "B": &MockProviderB{}, "C": &MockProviderC{}, "D": &MockProviderD{}},
			expected:          []string{"*foundation.MockProviderB", "*foundation.MockProviderC", "*foundation.MockProviderD", "*foundation.MockProviderB"},
		},
		{
			name:              "IsolatedNodes",
			graph:             map[string][]string{"A": {"B"}, "B": {"A"}, "C": {}, "D": {}},
			bindingToProvider: map[string]foundation.ServiceProvider{"A": &MockProviderA{}, "B": &MockProviderB{}, "C": &MockProviderC{}, "D": &MockProviderD{}},
			expected:          []string{"*foundation.MockProviderA", "*foundation.MockProviderB", "*foundation.MockProviderA"},
		},
		{
			name:              "LongCycle",
			graph:             map[string][]string{"A": {"B"}, "B": {"C"}, "C": {"D"}, "D": {"E"}, "E": {"A"}},
			bindingToProvider: map[string]foundation.ServiceProvider{"A": &MockProviderA{}, "B": &MockProviderB{}, "C": &MockProviderC{}, "D": &MockProviderD{}, "E": &MockProviderE{}},
			expected:          []string{"*foundation.MockProviderA", "*foundation.MockProviderB", "*foundation.MockProviderC", "*foundation.MockProviderD", "*foundation.MockProviderE", "*foundation.MockProviderA"},
		},
		{
			name:              "MissingProviderMapping",
			graph:             map[string][]string{"A": {"B"}, "B": {"A"}},
			bindingToProvider: map[string]foundation.ServiceProvider{"A": &MockProviderA{}}, // B missing
			expected:          []string{"*foundation.MockProviderA"},
		},
		{
			name:              "DuplicateProviderNames",
			graph:             map[string][]string{"A1": {"B"}, "A2": {"C"}, "B": {"A1"}, "C": {"A2"}},
			bindingToProvider: map[string]foundation.ServiceProvider{"A1": &MockProviderA{}, "A2": &MockProviderA{}, "B": &MockProviderB{}, "C": &MockProviderC{}},
			expected:          []string{"*foundation.MockProviderA", "*foundation.MockProviderB", "*foundation.MockProviderA"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := detectCycle(tc.graph, tc.bindingToProvider)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// Helper to check topological order
func isTopologicalOrder(providers []foundation.ServiceProvider, sorted []foundation.ServiceProvider) bool {
	providerIndex := make(map[foundation.ServiceProvider]int)
	for i, p := range sorted {
		providerIndex[p] = i
	}

	getDependencies := func(provider foundation.ServiceProvider) []string {
		if p, ok := provider.(interface{ Dependencies() []string }); ok {
			return p.Dependencies()
		}
		return []string{}
	}

	getBindings := func(provider foundation.ServiceProvider) []string {
		if p, ok := provider.(interface{ Relationship() binding.Relationship }); ok {
			return p.Relationship().Bindings
		}
		return []string{}
	}

	getProvideFor := func(provider foundation.ServiceProvider) []string {
		if p, ok := provider.(interface{ Relationship() binding.Relationship }); ok {
			return p.Relationship().ProvideFor
		}
		return []string{}
	}

	// Build binding to provider mapping
	bindingToProvider := make(map[string]foundation.ServiceProvider)
	for _, p := range providers {
		for _, b := range getBindings(p) {
			bindingToProvider[b] = p
		}
	}

	// Build provideFor to provider mapping (reverse relationship)
	provideForToProvider := make(map[string]foundation.ServiceProvider)
	for _, p := range providers {
		for _, pf := range getProvideFor(p) {
			provideForToProvider[pf] = p
		}
	}

	// Check all dependency relationships
	for _, p := range providers {
		// Check explicit dependencies (this provider depends on others)
		for _, dep := range getDependencies(p) {
			if depProvider, ok := bindingToProvider[dep]; ok {
				if providerIndex[depProvider] > providerIndex[p] {
					return false
				}
			}
		}

		// Check implicit dependencies through ProvideFor (others depend on this provider)
		for _, pf := range getProvideFor(p) {
			if dependentProvider, ok := bindingToProvider[pf]; ok {
				if providerIndex[p] > providerIndex[dependentProvider] {
					return false
				}
			}
		}
	}

	return true
}

type AServiceProvider struct{}

func (r *AServiceProvider) Relationship() binding.Relationship {
	return binding.Relationship{
		Bindings:     []string{"A"},
		Dependencies: []string{},
		ProvideFor:   []string{},
	}
}
func (r *AServiceProvider) Register(app foundation.Application) {}
func (r *AServiceProvider) Boot(app foundation.Application)     {}

type BServiceProvider struct{}

func (r *BServiceProvider) Relationship() binding.Relationship {
	return binding.Relationship{
		Bindings:     []string{"B"},
		Dependencies: []string{"A"},
		ProvideFor:   []string{"C"},
	}
}
func (r *BServiceProvider) Register(app foundation.Application) {}
func (r *BServiceProvider) Boot(app foundation.Application)     {}

type CServiceProvider struct{}

func (r *CServiceProvider) Relationship() binding.Relationship {
	return binding.Relationship{
		Bindings:     []string{"C"},
		Dependencies: []string{"A"},
		ProvideFor:   []string{},
	}
}
func (r *CServiceProvider) Register(app foundation.Application) {}
func (r *CServiceProvider) Boot(app foundation.Application)     {}

type CircularAServiceProvider struct{}

func (r *CircularAServiceProvider) Relationship() binding.Relationship {
	return binding.Relationship{
		Bindings:     []string{"CircularA"},
		Dependencies: []string{"CircularB"},
		ProvideFor:   []string{},
	}
}
func (r *CircularAServiceProvider) Register(app foundation.Application) {}
func (r *CircularAServiceProvider) Boot(app foundation.Application)     {}

type CircularBServiceProvider struct{}

func (r *CircularBServiceProvider) Relationship() binding.Relationship {
	return binding.Relationship{
		Bindings:     []string{"CircularB"},
		Dependencies: []string{"CircularA"},
		ProvideFor:   []string{},
	}
}
func (r *CircularBServiceProvider) Register(app foundation.Application) {}
func (r *CircularBServiceProvider) Boot(app foundation.Application)     {}

type BasicServiceProvider struct{}

func (r *BasicServiceProvider) Relationship() binding.Relationship {
	return binding.Relationship{
		Bindings:     []string{"Basic"},
		Dependencies: []string{},
		ProvideFor:   []string{},
	}
}
func (r *BasicServiceProvider) Register(app foundation.Application) {}
func (r *BasicServiceProvider) Boot(app foundation.Application)     {}

type ProvideForBServiceProvider struct{}

func (r *ProvideForBServiceProvider) Relationship() binding.Relationship {
	return binding.Relationship{
		Bindings:     []string{"ProvideForB"},
		Dependencies: []string{"ProvideForA"},
		ProvideFor:   []string{},
	}
}
func (r *ProvideForBServiceProvider) Register(app foundation.Application) {}
func (r *ProvideForBServiceProvider) Boot(app foundation.Application)     {}

type ProvideForAServiceProvider struct{}

func (r *ProvideForAServiceProvider) Relationship() binding.Relationship {
	return binding.Relationship{
		Bindings:     []string{"ProvideForA"},
		Dependencies: []string{},
		ProvideFor:   []string{"ProvideForB"},
	}
}
func (r *ProvideForAServiceProvider) Register(app foundation.Application) {}
func (r *ProvideForAServiceProvider) Boot(app foundation.Application)     {}

type MockProviderA struct{}

func (p *MockProviderA) Register(app foundation.Application) {}
func (p *MockProviderA) Boot(app foundation.Application)     {}
func (p *MockProviderA) Relationship() binding.Relationship {
	return binding.Relationship{
		Bindings:     []string{"provider_a"},
		Dependencies: []string{"provider_b"},
		ProvideFor:   []string{},
	}
}

type MockProviderB struct{}

func (p *MockProviderB) Register(app foundation.Application) {}
func (p *MockProviderB) Boot(app foundation.Application)     {}
func (p *MockProviderB) Relationship() binding.Relationship {
	return binding.Relationship{
		Bindings:     []string{"provider_b"},
		Dependencies: []string{"provider_a"},
		ProvideFor:   []string{},
	}
}

type MockProviderC struct{}

func (p *MockProviderC) Register(app foundation.Application) {}
func (p *MockProviderC) Boot(app foundation.Application)     {}
func (p *MockProviderC) Relationship() binding.Relationship {
	return binding.Relationship{
		Bindings:     []string{"provider_c"},
		Dependencies: []string{"provider_d"},
		ProvideFor:   []string{},
	}
}

type MockProviderD struct{}

func (p *MockProviderD) Register(app foundation.Application) {}
func (p *MockProviderD) Boot(app foundation.Application)     {}
func (p *MockProviderD) Relationship() binding.Relationship {
	return binding.Relationship{
		Bindings:     []string{"provider_d"},
		Dependencies: []string{"provider_c"},
		ProvideFor:   []string{},
	}
}

type MockProviderE struct{}

func (p *MockProviderE) Register(app foundation.Application) {}
func (p *MockProviderE) Boot(app foundation.Application)     {}
func (p *MockProviderE) Relationship() binding.Relationship {
	return binding.Relationship{
		Bindings:     []string{"provider_e"},
		Dependencies: []string{},
		ProvideFor:   []string{},
	}
}

type ComplexProviderA struct{}

func (p *ComplexProviderA) Register(app foundation.Application) {}
func (p *ComplexProviderA) Boot(app foundation.Application)     {}
func (p *ComplexProviderA) Relationship() binding.Relationship {
	return binding.Relationship{
		Bindings:     []string{"complex_a"},
		Dependencies: []string{"complex_b"},
		ProvideFor:   []string{},
	}
}

type ComplexProviderB struct{}

func (p *ComplexProviderB) Register(app foundation.Application) {}
func (p *ComplexProviderB) Boot(app foundation.Application)     {}
func (p *ComplexProviderB) Relationship() binding.Relationship {
	return binding.Relationship{
		Bindings:     []string{"complex_b"},
		Dependencies: []string{"complex_c"},
		ProvideFor:   []string{},
	}
}

type ComplexProviderC struct{}

func (p *ComplexProviderC) Register(app foundation.Application) {}
func (p *ComplexProviderC) Boot(app foundation.Application)     {}
func (p *ComplexProviderC) Relationship() binding.Relationship {
	return binding.Relationship{
		Bindings:     []string{"complex_c"},
		Dependencies: []string{"complex_a"},
		ProvideFor:   []string{},
	}
}

type EmptyDependenciesProvider struct{}

func (p *EmptyDependenciesProvider) Register(app foundation.Application) {}
func (p *EmptyDependenciesProvider) Boot(app foundation.Application)     {}
func (p *EmptyDependenciesProvider) Relationship() binding.Relationship {
	return binding.Relationship{
		Bindings:     []string{"empty_deps"},
		Dependencies: []string{},
		ProvideFor:   []string{"provider_c"},
	}
}

type EmptyProvideForProvider struct{}

func (p *EmptyProvideForProvider) Register(app foundation.Application) {}
func (p *EmptyProvideForProvider) Boot(app foundation.Application)     {}
func (p *EmptyProvideForProvider) Relationship() binding.Relationship {
	return binding.Relationship{
		Bindings:     []string{"empty_provide"},
		Dependencies: []string{"provider_a"},
		ProvideFor:   []string{},
	}
}

type AllEmptyProvider struct{}

func (p *AllEmptyProvider) Register(app foundation.Application) {}
func (p *AllEmptyProvider) Boot(app foundation.Application)     {}
func (p *AllEmptyProvider) Relationship() binding.Relationship {
	return binding.Relationship{
		Bindings:     []string{},
		Dependencies: []string{},
		ProvideFor:   []string{},
	}
}

type EmptyBindingsWithDependenciesProvider struct{}

func (p *EmptyBindingsWithDependenciesProvider) Register(app foundation.Application) {}
func (p *EmptyBindingsWithDependenciesProvider) Boot(app foundation.Application)     {}
func (p *EmptyBindingsWithDependenciesProvider) Relationship() binding.Relationship {
	return binding.Relationship{
		Bindings:     []string{},
		Dependencies: []string{"provider_c"},
		ProvideFor:   []string{},
	}
}

type EmptyBindingsWithProvideForProvider struct{}

func (p *EmptyBindingsWithProvideForProvider) Register(app foundation.Application) {}
func (p *EmptyBindingsWithProvideForProvider) Boot(app foundation.Application)     {}
func (p *EmptyBindingsWithProvideForProvider) Relationship() binding.Relationship {
	return binding.Relationship{
		Bindings:     []string{},
		Dependencies: []string{},
		ProvideFor:   []string{"provider_a"},
	}
}

type EmptyBindingsWithBothProvider struct{}

func (p *EmptyBindingsWithBothProvider) Register(app foundation.Application) {}
func (p *EmptyBindingsWithBothProvider) Boot(app foundation.Application)     {}
func (p *EmptyBindingsWithBothProvider) Relationship() binding.Relationship {
	return binding.Relationship{
		Bindings:     []string{},
		Dependencies: []string{"provider_e"},
		ProvideFor:   []string{"provider_c"},
	}
}

type EmptyBindingsCircularAProvider struct{}

func (p *EmptyBindingsCircularAProvider) Register(app foundation.Application) {}
func (p *EmptyBindingsCircularAProvider) Boot(app foundation.Application)     {}
func (p *EmptyBindingsCircularAProvider) Relationship() binding.Relationship {
	return binding.Relationship{
		Bindings:     []string{},
		Dependencies: []string{"__virtual_1"},
		ProvideFor:   []string{},
	}
}

type EmptyBindingsCircularBProvider struct{}

func (p *EmptyBindingsCircularBProvider) Register(app foundation.Application) {}
func (p *EmptyBindingsCircularBProvider) Boot(app foundation.Application)     {}
func (p *EmptyBindingsCircularBProvider) Relationship() binding.Relationship {
	return binding.Relationship{
		Bindings:     []string{},
		Dependencies: []string{"__virtual_0"},
		ProvideFor:   []string{},
	}
}

type CircularBindingAProvider struct{}

func (p *CircularBindingAProvider) Register(app foundation.Application) {}
func (p *CircularBindingAProvider) Boot(app foundation.Application)     {}
func (p *CircularBindingAProvider) Relationship() binding.Relationship {
	return binding.Relationship{
		Bindings:     []string{"circular_binding_a"},
		Dependencies: []string{"circular_binding_b"},
		ProvideFor:   []string{},
	}
}

type CircularBindingBProvider struct{}

func (p *CircularBindingBProvider) Register(app foundation.Application) {}
func (p *CircularBindingBProvider) Boot(app foundation.Application)     {}
func (p *CircularBindingBProvider) Relationship() binding.Relationship {
	return binding.Relationship{
		Bindings:     []string{"circular_binding_b"},
		Dependencies: []string{"circular_binding_a"},
		ProvideFor:   []string{},
	}
}
