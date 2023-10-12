package foundation

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/auth"
	"github.com/goravel/framework/cache"
	"github.com/goravel/framework/config"
	"github.com/goravel/framework/console"
	cachemocks "github.com/goravel/framework/contracts/cache/mocks"
	configmocks "github.com/goravel/framework/contracts/config/mocks"
	consolemocks "github.com/goravel/framework/contracts/console/mocks"
	"github.com/goravel/framework/contracts/database/orm"
	ormmocks "github.com/goravel/framework/contracts/database/orm/mocks"
	"github.com/goravel/framework/contracts/foundation"
	logmocks "github.com/goravel/framework/contracts/log/mocks"
	queuemocks "github.com/goravel/framework/contracts/queue/mocks"
	routemocks "github.com/goravel/framework/contracts/route/mocks"
	"github.com/goravel/framework/crypt"
	"github.com/goravel/framework/database"
	"github.com/goravel/framework/database/gorm"
	"github.com/goravel/framework/event"
	"github.com/goravel/framework/filesystem"
	"github.com/goravel/framework/grpc"
	"github.com/goravel/framework/hash"
	"github.com/goravel/framework/http"
	"github.com/goravel/framework/log"
	"github.com/goravel/framework/mail"
	"github.com/goravel/framework/queue"
	"github.com/goravel/framework/schedule"
	"github.com/goravel/framework/support/file"
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

func (s *ApplicationTestSuite) TestPublicPath() {
	s.Equal(filepath.Join("public", "goravel.go"), s.app.PublicPath("goravel.go"))
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
	mockConfig := &configmocks.Config{}
	mockConfig.On("GetString", "auth.defaults.guard").Return("user").Once()

	s.app.Singleton(config.Binding, func(app foundation.Application) (any, error) {
		return mockConfig, nil
	})
	s.app.Singleton(cache.Binding, func(app foundation.Application) (any, error) {
		return &cachemocks.Cache{}, nil
	})
	s.app.Singleton(database.BindingOrm, func(app foundation.Application) (any, error) {
		return &ormmocks.Orm{}, nil
	})

	serviceProvider := &auth.ServiceProvider{}
	serviceProvider.Register(s.app)

	s.NotNil(s.app.MakeAuth())
	mockConfig.AssertExpectations(s.T())
}

func (s *ApplicationTestSuite) TestMakeCache() {
	mockConfig := &configmocks.Config{}
	mockConfig.On("GetString", "cache.default").Return("memory").Once()
	mockConfig.On("GetString", "cache.stores.memory.driver").Return("memory").Once()
	mockConfig.On("GetString", "cache.prefix").Return("goravel").Once()

	s.app.Singleton(config.Binding, func(app foundation.Application) (any, error) {
		return mockConfig, nil
	})
	s.app.Singleton(log.Binding, func(app foundation.Application) (any, error) {
		return &logmocks.Log{}, nil
	})

	serviceProvider := &cache.ServiceProvider{}
	serviceProvider.Register(s.app)

	s.NotNil(s.app.MakeCache())
	mockConfig.AssertExpectations(s.T())
}

func (s *ApplicationTestSuite) TestMakeConfig() {
	serviceProvider := &config.ServiceProvider{}
	serviceProvider.Register(s.app)

	s.NotNil(s.app.MakeConfig())
}

func (s *ApplicationTestSuite) TestMakeCrypt() {
	mockConfig := &configmocks.Config{}
	mockConfig.On("GetString", "app.key").Return("12345678901234567890123456789012").Once()

	s.app.Singleton(config.Binding, func(app foundation.Application) (any, error) {
		return mockConfig, nil
	})

	serviceProvider := &crypt.ServiceProvider{}
	serviceProvider.Register(s.app)

	s.NotNil(s.app.MakeCrypt())
	mockConfig.AssertExpectations(s.T())
}

func (s *ApplicationTestSuite) TestMakeEvent() {
	s.app.Singleton(queue.Binding, func(app foundation.Application) (any, error) {
		return &queuemocks.Queue{}, nil
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
	s.app.Singleton(config.Binding, func(app foundation.Application) (any, error) {
		return &configmocks.Config{}, nil
	})

	serviceProvider := &grpc.ServiceProvider{}
	serviceProvider.Register(s.app)

	s.NotNil(s.app.MakeGrpc())
}

func (s *ApplicationTestSuite) TestMakeHash() {
	mockConfig := &configmocks.Config{}
	mockConfig.On("GetString", "hashing.driver", "argon2id").Return("argon2id").Once()
	mockConfig.On("GetInt", "hashing.argon2id.time", 4).Return(4).Once()
	mockConfig.On("GetInt", "hashing.argon2id.memory", 65536).Return(65536).Once()
	mockConfig.On("GetInt", "hashing.argon2id.threads", 1).Return(1).Once()

	s.app.Singleton(config.Binding, func(app foundation.Application) (any, error) {
		return mockConfig, nil
	})

	serviceProvider := &hash.ServiceProvider{}
	serviceProvider.Register(s.app)

	s.NotNil(s.app.MakeHash())
	mockConfig.AssertExpectations(s.T())
}

func (s *ApplicationTestSuite) TestMakeLog() {
	serviceProvider := &log.ServiceProvider{}
	serviceProvider.Register(s.app)

	s.NotNil(s.app.MakeLog())
}

func (s *ApplicationTestSuite) TestMakeMail() {
	s.app.Singleton(config.Binding, func(app foundation.Application) (any, error) {
		return &configmocks.Config{}, nil
	})
	s.app.Singleton(queue.Binding, func(app foundation.Application) (any, error) {
		return &queuemocks.Queue{}, nil
	})

	serviceProvider := &mail.ServiceProvider{}
	serviceProvider.Register(s.app)

	s.NotNil(s.app.MakeMail())
}

func (s *ApplicationTestSuite) TestMakeOrm() {
	if testing.Short() {
		s.T().Skip("Skipping tests of using docker")
	}

	mysqlDocker := gorm.NewMysqlDocker()
	pool, resource, _, err := mysqlDocker.New()
	s.Nil(err)

	mockConfig := &configmocks.Config{}
	mockConfig.On("GetString", "database.default").Return("mysql").Once()
	mockConfig.On("Get", "database.connections.mysql.read").Return(nil).Once()
	mockConfig.On("Get", "database.connections.mysql.write").Return(nil).Once()
	mockConfig.On("GetString", "database.connections.mysql.driver").Return(orm.DriverMysql.String()).Twice()
	mockConfig.On("GetString", "database.connections.mysql.charset").Return("utf8mb4").Once()
	mockConfig.On("GetString", "database.connections.mysql.loc").Return("Local").Once()
	mockConfig.On("GetString", "database.connections.mysql.database").Return("mysql").Once()
	mockConfig.On("GetString", "database.connections.mysql.host").Return("localhost").Once()
	mockConfig.On("GetString", "database.connections.mysql.username").Return(gorm.DbUser).Once()
	mockConfig.On("GetString", "database.connections.mysql.password").Return(gorm.DbPassword).Once()
	mockConfig.On("GetString", "database.connections.mysql.prefix").Return("").Once()
	mockConfig.On("GetInt", "database.connections.mysql.port").Return(mysqlDocker.Port).Once()
	mockConfig.On("GetBool", "database.connections.mysql.singular").Return(true).Once()
	mockConfig.On("GetBool", "app.debug").Return(true).Once()
	mockConfig.On("GetInt", "database.pool.max_idle_conns", 10).Return(10)
	mockConfig.On("GetInt", "database.pool.max_open_conns", 100).Return(100)
	mockConfig.On("GetInt", "database.pool.conn_max_idletime", 3600).Return(3600)
	mockConfig.On("GetInt", "database.pool.conn_max_lifetime", 3600).Return(3600)

	s.app.Singleton(config.Binding, func(app foundation.Application) (any, error) {
		return mockConfig, nil
	})

	serviceProvider := &database.ServiceProvider{}
	serviceProvider.Register(s.app)

	s.NotNil(s.app.MakeOrm())
	s.Nil(pool.Purge(resource))
	mockConfig.AssertExpectations(s.T())
}

func (s *ApplicationTestSuite) TestMakeQueue() {
	s.app.Singleton(config.Binding, func(app foundation.Application) (any, error) {
		return &configmocks.Config{}, nil
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
	mockConfig := &configmocks.Config{}

	s.app.Singleton(config.Binding, func(app foundation.Application) (any, error) {
		return mockConfig, nil
	})

	mockRoute := &routemocks.Route{}
	s.app.Singleton("goravel.route", func(app foundation.Application) (any, error) {
		return mockRoute, nil
	})

	s.NotNil(s.app.MakeRoute())
	mockConfig.AssertExpectations(s.T())
}

func (s *ApplicationTestSuite) TestMakeSchedule() {
	s.app.Singleton(console.Binding, func(app foundation.Application) (any, error) {
		return &consolemocks.Artisan{}, nil
	})
	s.app.Singleton(log.Binding, func(app foundation.Application) (any, error) {
		return &logmocks.Log{}, nil
	})

	serviceProvider := &schedule.ServiceProvider{}
	serviceProvider.Register(s.app)

	s.NotNil(s.app.MakeSchedule())
}

func (s *ApplicationTestSuite) TestMakeStorage() {
	mockConfig := &configmocks.Config{}
	mockConfig.On("GetString", "filesystems.default").Return("local").Once()
	mockConfig.On("GetString", "filesystems.disks.local.driver").Return("local").Once()
	mockConfig.On("GetString", "filesystems.disks.local.root").Return("").Once()
	mockConfig.On("GetString", "filesystems.disks.local.url").Return("").Once()

	s.app.Singleton(config.Binding, func(app foundation.Application) (any, error) {
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
