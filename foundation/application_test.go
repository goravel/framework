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
	"github.com/goravel/framework/contracts"
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

func (s *ApplicationTestSuite) TestStoragePath() {
	s.Equal(filepath.Join(support.RootPath, "storage", "goravel.go"), s.app.StoragePath("goravel.go"))
}

func (s *ApplicationTestSuite) TestResourcePath() {
	s.Equal(filepath.Join(support.RootPath, "resources", "goravel.go"), s.app.ResourcePath("goravel.go"))
}

func (s *ApplicationTestSuite) TestLangPath() {
	mockConfig := mocksconfig.NewConfig(s.T())
	mockConfig.EXPECT().GetString("app.lang_path", "lang").Return("test").Once()

	s.app.Singleton(contracts.BindingConfig, func(app foundation.Application) (any, error) {
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
	mockConfig.EXPECT().GetString("jwt.secret").Return("secret").Once()
	mockConfig.EXPECT().Get("auth.guards.user.ttl").Return(100).Once()
	mockConfig.EXPECT().GetInt("jwt.refresh_ttl").Return(100).Once()

	s.app.Singleton(contracts.BindingConfig, func(app foundation.Application) (any, error) {
		return mockConfig, nil
	})
	s.app.Singleton(contracts.BindingCache, func(app foundation.Application) (any, error) {
		return &mockscache.Cache{}, nil
	})
	s.app.Singleton(contracts.BindingOrm, func(app foundation.Application) (any, error) {
		return &mocksorm.Orm{}, nil
	})
	s.app.Singleton(contracts.BindingLog, func(app foundation.Application) (any, error) {
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

	s.app.Singleton(contracts.BindingConfig, func(app foundation.Application) (any, error) {
		return mockConfig, nil
	})
	s.app.Singleton(contracts.BindingLog, func(app foundation.Application) (any, error) {
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

	s.app.Singleton(contracts.BindingConfig, func(app foundation.Application) (any, error) {
		return mockConfig, nil
	})
	s.app.SetJson(mocksfoundation.NewJson(s.T()))

	serviceProvider := &crypt.ServiceProvider{}
	serviceProvider.Register(s.app)

	s.NotNil(s.app.MakeCrypt())
}

func (s *ApplicationTestSuite) TestMakeEvent() {
	s.app.Singleton(contracts.BindingQueue, func(app foundation.Application) (any, error) {
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
	s.app.Singleton(contracts.BindingConfig, func(app foundation.Application) (any, error) {
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

	s.app.Singleton(contracts.BindingConfig, func(app foundation.Application) (any, error) {
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

	s.app.Singleton(contracts.BindingConfig, func(app foundation.Application) (any, error) {
		return mockConfig, nil
	})
	s.app.Singleton(contracts.BindingLog, func(app foundation.Application) (any, error) {
		return &mockslog.Log{}, nil
	})

	serviceProvider := &frameworktranslation.ServiceProvider{}
	serviceProvider.Register(s.app)
	ctx := http.Background()

	s.NotNil(s.app.MakeLang(ctx))
}

func (s *ApplicationTestSuite) TestMakeLog() {
	mockConfig := mocksconfig.NewConfig(s.T())
	s.app.Singleton(contracts.BindingConfig, func(app foundation.Application) (any, error) {
		return mockConfig, nil
	})

	mockConfig.EXPECT().GetString("logging.default").Return("").Once()

	s.app.SetJson(mocksfoundation.NewJson(s.T()))

	serviceProvider := &frameworklog.ServiceProvider{}
	serviceProvider.Register(s.app)

	s.NotNil(s.app.MakeLog())
}

func (s *ApplicationTestSuite) TestMakeMail() {
	s.app.Singleton(contracts.BindingConfig, func(app foundation.Application) (any, error) {
		return &mocksconfig.Config{}, nil
	})
	s.app.Singleton(contracts.BindingQueue, func(app foundation.Application) (any, error) {
		return &mocksqueue.Queue{}, nil
	})

	serviceProvider := &mail.ServiceProvider{}
	serviceProvider.Register(s.app)

	s.NotNil(s.app.MakeMail())
}

func (s *ApplicationTestSuite) TestMakeQueue() {
	mockConfig := mocksconfig.NewConfig(s.T())
	s.app.Singleton(contracts.BindingConfig, func(app foundation.Application) (any, error) {
		return mockConfig, nil
	})
	s.app.Singleton(contracts.BindingDB, func(app foundation.Application) (any, error) {
		return &mocksdb.DB{}, nil
	})
	s.app.Singleton(contracts.BindingLog, func(app foundation.Application) (any, error) {
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

	s.app.Singleton(contracts.BindingConfig, func(app foundation.Application) (any, error) {
		return mockConfig, nil
	})

	s.app.Singleton(contracts.BindingRoute, func(app foundation.Application) (any, error) {
		return &mocksroute.Route{}, nil
	})

	s.NotNil(s.app.MakeRoute())
}

func (s *ApplicationTestSuite) TestMakeSchedule() {
	mockConfig := mocksconfig.NewConfig(s.T())
	mockConfig.EXPECT().GetBool("app.debug").Return(false).Once()

	s.app.Singleton(contracts.BindingConfig, func(app foundation.Application) (any, error) {
		return mockConfig, nil
	})
	s.app.Singleton(contracts.BindingConsole, func(app foundation.Application) (any, error) {
		return &mocksconsole.Artisan{}, nil
	})
	s.app.Singleton(contracts.BindingLog, func(app foundation.Application) (any, error) {
		return &mockslog.Log{}, nil
	})
	s.app.Singleton(contracts.BindingCache, func(app foundation.Application) (any, error) {
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

	s.app.Singleton(contracts.BindingConfig, func(app foundation.Application) (any, error) {
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

	s.app.Singleton(contracts.BindingConfig, func(app foundation.Application) (any, error) {
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
