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
	contractsdatabase "github.com/goravel/framework/contracts/database"
	"github.com/goravel/framework/contracts/foundation"
	"github.com/goravel/framework/crypt"
	"github.com/goravel/framework/database"
	"github.com/goravel/framework/event"
	"github.com/goravel/framework/filesystem"
	"github.com/goravel/framework/foundation/json"
	"github.com/goravel/framework/grpc"
	"github.com/goravel/framework/hash"
	"github.com/goravel/framework/http"
	frameworklog "github.com/goravel/framework/log"
	"github.com/goravel/framework/mail"
	mockscache "github.com/goravel/framework/mocks/cache"
	mocksconfig "github.com/goravel/framework/mocks/config"
	mocksconsole "github.com/goravel/framework/mocks/console"
	mocksorm "github.com/goravel/framework/mocks/database/orm"
	mockslog "github.com/goravel/framework/mocks/log"
	mocksqueue "github.com/goravel/framework/mocks/queue"
	mocksroute "github.com/goravel/framework/mocks/route"
	"github.com/goravel/framework/queue"
	"github.com/goravel/framework/schedule"
	frameworksession "github.com/goravel/framework/session"
	supportdocker "github.com/goravel/framework/support/docker"
	"github.com/goravel/framework/support/env"
	"github.com/goravel/framework/support/file"
	frameworktranslation "github.com/goravel/framework/translation"
	"github.com/goravel/framework/validation"
)

type ApplicationTestSuite struct {
	suite.Suite
	app *Application
}

func TestApplicationTestSuite(t *testing.T) {
	assert.Nil(t, file.Create(".env", "APP_KEY=12345678901234567890123456789012"))

	suite.Run(t, new(ApplicationTestSuite))

	assert.Nil(t, file.Remove(".env"))
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
	s.Equal(filepath.Join("app", "goravel.go"), s.app.Path("goravel.go"))
}

func (s *ApplicationTestSuite) TestBasePath() {
	s.Equal("goravel.go", s.app.BasePath("goravel.go"))
}

func (s *ApplicationTestSuite) TestConfigPath() {
	s.Equal(filepath.Join("config", "goravel.go"), s.app.ConfigPath("goravel.go"))
}

func (s *ApplicationTestSuite) TestDatabasePath() {
	s.Equal(filepath.Join("database", "goravel.go"), s.app.DatabasePath("goravel.go"))
}

func (s *ApplicationTestSuite) TestStoragePath() {
	s.Equal(filepath.Join("storage", "goravel.go"), s.app.StoragePath("goravel.go"))
}

func (s *ApplicationTestSuite) TestLangPath() {
	mockConfig := &mocksconfig.Config{}
	mockConfig.EXPECT().GetString("app.lang_path", "lang").Return("test").Once()

	s.app.Singleton(frameworkconfig.Binding, func(app foundation.Application) (any, error) {
		return mockConfig, nil
	})

	s.Equal(filepath.Join("test", "goravel.go"), s.app.LangPath("goravel.go"))
}

func (s *ApplicationTestSuite) TestPublicPath() {
	s.Equal(filepath.Join("public", "goravel.go"), s.app.PublicPath("goravel.go"))
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
	mockConfig := &mocksconfig.Config{}
	mockConfig.EXPECT().GetString("auth.defaults.guard").Return("user").Once()

	s.app.Singleton(frameworkconfig.Binding, func(app foundation.Application) (any, error) {
		return mockConfig, nil
	})
	s.app.Singleton(cache.Binding, func(app foundation.Application) (any, error) {
		return &mockscache.Cache{}, nil
	})
	s.app.Singleton(database.BindingOrm, func(app foundation.Application) (any, error) {
		return &mocksorm.Orm{}, nil
	})

	serviceProvider := &auth.ServiceProvider{}
	serviceProvider.Register(s.app)

	s.NotNil(s.app.MakeAuth(http.Background()))
	mockConfig.AssertExpectations(s.T())
}

func (s *ApplicationTestSuite) TestMakeCache() {
	mockConfig := &mocksconfig.Config{}
	mockConfig.EXPECT().GetString("cache.default").Return("memory").Once()
	mockConfig.EXPECT().GetString("cache.stores.memory.driver").Return("memory").Once()
	mockConfig.EXPECT().GetString("cache.prefix").Return("goravel").Once()

	s.app.Singleton(frameworkconfig.Binding, func(app foundation.Application) (any, error) {
		return mockConfig, nil
	})
	s.app.Singleton(frameworklog.Binding, func(app foundation.Application) (any, error) {
		return &mockslog.Log{}, nil
	})

	serviceProvider := &cache.ServiceProvider{}
	serviceProvider.Register(s.app)

	s.NotNil(s.app.MakeCache())
	mockConfig.AssertExpectations(s.T())
}

func (s *ApplicationTestSuite) TestMakeConfig() {
	serviceProvider := &frameworkconfig.ServiceProvider{}
	serviceProvider.Register(s.app)

	s.NotNil(s.app.MakeConfig())
}

func (s *ApplicationTestSuite) TestMakeCrypt() {
	mockConfig := &mocksconfig.Config{}
	mockConfig.EXPECT().GetString("app.key").Return("12345678901234567890123456789012").Once()

	s.app.Singleton(frameworkconfig.Binding, func(app foundation.Application) (any, error) {
		return mockConfig, nil
	})
	s.app.SetJson(json.NewJson())

	serviceProvider := &crypt.ServiceProvider{}
	serviceProvider.Register(s.app)

	s.NotNil(s.app.MakeCrypt())
	mockConfig.AssertExpectations(s.T())
}

func (s *ApplicationTestSuite) TestMakeEvent() {
	s.app.Singleton(queue.Binding, func(app foundation.Application) (any, error) {
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
	s.app.Singleton(frameworkconfig.Binding, func(app foundation.Application) (any, error) {
		return &mocksconfig.Config{}, nil
	})

	serviceProvider := &grpc.ServiceProvider{}
	serviceProvider.Register(s.app)

	s.NotNil(s.app.MakeGrpc())
}

func (s *ApplicationTestSuite) TestMakeHash() {
	mockConfig := &mocksconfig.Config{}
	mockConfig.EXPECT().GetString("hashing.driver", "argon2id").Return("argon2id").Once()
	mockConfig.EXPECT().GetInt("hashing.argon2id.time", 4).Return(4).Once()
	mockConfig.EXPECT().GetInt("hashing.argon2id.memory", 65536).Return(65536).Once()
	mockConfig.EXPECT().GetInt("hashing.argon2id.threads", 1).Return(1).Once()

	s.app.Singleton(frameworkconfig.Binding, func(app foundation.Application) (any, error) {
		return mockConfig, nil
	})

	serviceProvider := &hash.ServiceProvider{}
	serviceProvider.Register(s.app)

	s.NotNil(s.app.MakeHash())
	mockConfig.AssertExpectations(s.T())
}

func (s *ApplicationTestSuite) TestMakeLang() {
	mockConfig := &mocksconfig.Config{}
	mockConfig.EXPECT().GetString("app.locale").Return("en").Once()
	mockConfig.EXPECT().GetString("app.fallback_locale").Return("en").Once()
	mockConfig.EXPECT().GetString("app.lang_path", "lang").Return("lang").Once()

	s.app.Singleton(frameworkconfig.Binding, func(app foundation.Application) (any, error) {
		return mockConfig, nil
	})
	s.app.Singleton(frameworklog.Binding, func(app foundation.Application) (any, error) {
		return &mockslog.Log{}, nil
	})

	serviceProvider := &frameworktranslation.ServiceProvider{}
	serviceProvider.Register(s.app)
	ctx := http.Background()

	s.NotNil(s.app.MakeLang(ctx))
	mockConfig.AssertExpectations(s.T())
}

func (s *ApplicationTestSuite) TestMakeLog() {
	serviceProvider := &frameworklog.ServiceProvider{}
	serviceProvider.Register(s.app)

	s.NotNil(s.app.MakeLog())
}

func (s *ApplicationTestSuite) TestMakeMail() {
	s.app.Singleton(frameworkconfig.Binding, func(app foundation.Application) (any, error) {
		return &mocksconfig.Config{}, nil
	})
	s.app.Singleton(queue.Binding, func(app foundation.Application) (any, error) {
		return &mocksqueue.Queue{}, nil
	})

	serviceProvider := &mail.ServiceProvider{}
	serviceProvider.Register(s.app)

	s.NotNil(s.app.MakeMail())
}

func (s *ApplicationTestSuite) TestMakeOrm() {
	if env.IsWindows() {
		s.T().Skip("Skipping tests of using docker")
	}

	postgresDocker := supportdocker.Postgres()
	config := postgresDocker.Config()
	mockConfig := mocksconfig.NewConfig(s.T())
	mockConfig.EXPECT().GetString("database.default").Return("postgres").Once()
	mockConfig.EXPECT().Get("database.connections.postgres.read").Return(nil).Once()
	mockConfig.EXPECT().Get("database.connections.postgres.write").Return(nil).Twice()
	mockConfig.EXPECT().GetString("database.connections.postgres.driver").Return(contractsdatabase.DriverPostgres.String()).Twice()
	mockConfig.EXPECT().GetString("database.connections.postgres.prefix").Return("").Twice()
	mockConfig.EXPECT().GetBool("database.connections.postgres.singular").Return(true).Twice()
	mockConfig.EXPECT().GetString("database.connections.postgres.host").Return("localhost").Twice()
	mockConfig.EXPECT().GetString("database.connections.postgres.username").Return(config.Username).Twice()
	mockConfig.EXPECT().GetString("database.connections.postgres.password").Return(config.Password).Twice()
	mockConfig.EXPECT().GetInt("database.connections.postgres.port").Return(config.Port).Twice()
	mockConfig.EXPECT().GetString("database.connections.postgres.sslmode").Return("disable").Twice()
	mockConfig.EXPECT().GetString("database.connections.postgres.timezone").Return("UTC").Twice()
	mockConfig.EXPECT().GetString("database.connections.postgres.database").Return(config.Database).Twice()
	mockConfig.EXPECT().GetBool("app.debug").Return(true).Once()
	mockConfig.EXPECT().GetInt("database.pool.max_idle_conns", 10).Return(10).Once()
	mockConfig.EXPECT().GetInt("database.pool.max_open_conns", 100).Return(100).Once()
	mockConfig.EXPECT().GetInt("database.pool.conn_max_idletime", 3600).Return(3600).Once()
	mockConfig.EXPECT().GetInt("database.pool.conn_max_lifetime", 3600).Return(3600).Once()

	s.app.Singleton(frameworkconfig.Binding, func(app foundation.Application) (any, error) {
		return mockConfig, nil
	})

	serviceProvider := &database.ServiceProvider{}
	serviceProvider.Register(s.app)

	s.NotNil(s.app.MakeOrm())
}

func (s *ApplicationTestSuite) TestMakeQueue() {
	s.app.Singleton(frameworkconfig.Binding, func(app foundation.Application) (any, error) {
		return &mocksconfig.Config{}, nil
	})

	serviceProvider := &queue.ServiceProvider{}
	serviceProvider.Register(s.app)

	s.NotNil(s.app.MakeQueue())
}

func (s *ApplicationTestSuite) TestMakeRateLimiter() {
	serviceProvider := &http.ServiceProvider{}
	serviceProvider.Register(s.app)

	s.NotNil(s.app.MakeRateLimiter())
}

func (s *ApplicationTestSuite) TestMakeRoute() {
	mockConfig := &mocksconfig.Config{}

	s.app.Singleton(frameworkconfig.Binding, func(app foundation.Application) (any, error) {
		return mockConfig, nil
	})

	mockRoute := &mocksroute.Route{}
	s.app.Singleton("goravel.route", func(app foundation.Application) (any, error) {
		return mockRoute, nil
	})

	s.NotNil(s.app.MakeRoute())
	mockConfig.AssertExpectations(s.T())
}

func (s *ApplicationTestSuite) TestMakeSchedule() {
	mockConfig := &mocksconfig.Config{}
	mockConfig.EXPECT().GetBool("app.debug").Return(false).Once()

	s.app.Singleton(frameworkconfig.Binding, func(app foundation.Application) (any, error) {
		return mockConfig, nil
	})
	s.app.Singleton(console.Binding, func(app foundation.Application) (any, error) {
		return &mocksconsole.Artisan{}, nil
	})
	s.app.Singleton(frameworklog.Binding, func(app foundation.Application) (any, error) {
		return &mockslog.Log{}, nil
	})

	serviceProvider := &schedule.ServiceProvider{}
	serviceProvider.Register(s.app)

	s.NotNil(s.app.MakeSchedule())
	mockConfig.AssertExpectations(s.T())
}

func (s *ApplicationTestSuite) TestMakeSession() {
	mockConfig := &mocksconfig.Config{}
	mockConfig.EXPECT().GetInt("session.lifetime").Return(120).Once()
	mockConfig.EXPECT().GetInt("session.gc_interval", 30).Return(30).Once()
	mockConfig.EXPECT().GetString("session.files").Return("storage/framework/sessions").Once()

	s.app.Singleton(frameworkconfig.Binding, func(app foundation.Application) (any, error) {
		return mockConfig, nil
	})
	s.app.SetJson(json.NewJson())

	serviceProvider := &frameworksession.ServiceProvider{}
	// error
	s.Nil(s.app.MakeSession())

	serviceProvider.Register(s.app)
	s.NotNil(s.app.MakeSession())

	mockConfig.AssertExpectations(s.T())
}

func (s *ApplicationTestSuite) TestMakeStorage() {
	mockConfig := &mocksconfig.Config{}
	mockConfig.EXPECT().GetString("filesystems.default").Return("local").Once()
	mockConfig.EXPECT().GetString("filesystems.disks.local.driver").Return("local").Once()
	mockConfig.EXPECT().GetString("filesystems.disks.local.root").Return("").Once()
	mockConfig.EXPECT().GetString("filesystems.disks.local.url").Return("").Once()

	s.app.Singleton(frameworkconfig.Binding, func(app foundation.Application) (any, error) {
		return mockConfig, nil
	})

	serviceProvider := &filesystem.ServiceProvider{}
	serviceProvider.Register(s.app)

	s.NotNil(s.app.MakeStorage())
	mockConfig.AssertExpectations(s.T())
}

func (s *ApplicationTestSuite) TestMakeValidation() {
	serviceProvider := &validation.ServiceProvider{}
	serviceProvider.Register(s.app)

	s.NotNil(s.app.MakeValidation())
}
