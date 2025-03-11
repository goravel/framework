package postgres

import (
	"testing"

	mocksconfig "github.com/goravel/framework/mocks/config"
	"github.com/goravel/framework/testing/utils"
	"github.com/goravel/postgres/contracts"
	mocks "github.com/goravel/postgres/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type GormTestSuite struct {
	suite.Suite
	mockConfig *mocks.ConfigBuilder
	gorm       *Gorm
}

func TestGormSuite(t *testing.T) {
	suite.Run(t, new(GormTestSuite))
}

func (s *GormTestSuite) SetupTest() {
	s.mockConfig = mocks.NewConfigBuilder(s.T())
	s.gorm = NewGorm(s.mockConfig, utils.NewTestLog())
}

func (s *GormTestSuite) TestBuild() {
	writes := []contracts.FullConfig{
		{
			Config: contracts.Config{
				Host:     "localhost",
				Database: "goravel",
				Username: "goravel",
				Password: "Framework!123",
				Schema:   "public",
			},
			Timezone: "UTC",
		},
	}
	reads := []contracts.FullConfig{
		{
			Config: contracts.Config{
				Host:     "localhost",
				Database: "goravel",
				Username: "goravel",
				Password: "Framework!123",
				Schema:   "public",
			},
			Timezone: "UTC",
		},
	}

	s.Run("single config", func() {
		docker := NewDocker(nil, writes[0].Database, writes[0].Username, writes[0].Password)
		s.NoError(docker.Build())

		writes[0].Config.Port = docker.port
		_, err := docker.connect()
		s.NoError(err)

		mockConfigFacade := mocksconfig.NewConfig(s.T())

		// instance
		s.mockConfig.EXPECT().Writes().Return(writes).Once()
		s.mockConfig.EXPECT().Reads().Return([]contracts.FullConfig{}).Once()

		// gormConfig
		s.mockConfig.EXPECT().Config().Return(mockConfigFacade).Once()
		mockConfigFacade.EXPECT().GetBool("app.debug").Return(true).Once()
		mockConfigFacade.EXPECT().GetInt("database.slow_threshold", 200).Return(200).Once()
		s.mockConfig.EXPECT().Writes().Return(writes).Once()

		// configurePool
		mockConfigFacade.EXPECT().GetInt("database.pool.max_idle_conns", 10).Return(10).Once()
		mockConfigFacade.EXPECT().GetInt("database.pool.max_open_conns", 100).Return(100).Once()
		mockConfigFacade.EXPECT().GetInt("database.pool.conn_max_idletime", 3600).Return(3600).Once()
		mockConfigFacade.EXPECT().GetInt("database.pool.conn_max_lifetime", 3600).Return(3600).Once()
		s.mockConfig.EXPECT().Config().Return(mockConfigFacade).Once()

		// configureReadWriteSeparate
		s.mockConfig.EXPECT().Writes().Return(writes).Once()
		s.mockConfig.EXPECT().Reads().Return([]contracts.FullConfig{}).Once()

		db, err := s.gorm.Build()
		s.NoError(err)
		s.NotNil(db)
		s.NoError(docker.Shutdown())
	})

	s.Run("config with writes and reads", func() {
		docker := NewDocker(nil, writes[0].Database, writes[0].Username, writes[0].Password)
		s.NoError(docker.Build())

		writes[0].Config.Port = docker.port
		_, err := docker.connect()
		s.NoError(err)

		reads[0].Config.Port = docker.port
		_, err = docker.connect()
		s.NoError(err)

		mockConfigFacade := mocksconfig.NewConfig(s.T())

		// instance
		s.mockConfig.EXPECT().Writes().Return(writes).Once()
		s.mockConfig.EXPECT().Reads().Return(reads).Once()

		// gormConfig
		s.mockConfig.EXPECT().Config().Return(mockConfigFacade).Once()
		mockConfigFacade.EXPECT().GetBool("app.debug").Return(true).Once()
		mockConfigFacade.EXPECT().GetInt("database.slow_threshold", 200).Return(200).Once()
		s.mockConfig.EXPECT().Writes().Return(writes).Once()

		// configurePool
		mockConfigFacade.EXPECT().GetInt("database.pool.max_idle_conns", 10).Return(10).Once()
		mockConfigFacade.EXPECT().GetInt("database.pool.max_open_conns", 100).Return(100).Once()
		mockConfigFacade.EXPECT().GetInt("database.pool.conn_max_idletime", 3600).Return(3600).Once()
		mockConfigFacade.EXPECT().GetInt("database.pool.conn_max_lifetime", 3600).Return(3600).Once()
		s.mockConfig.EXPECT().Config().Return(mockConfigFacade).Once()

		// configureReadWriteSeparate
		s.mockConfig.EXPECT().Writes().Return(writes).Once()
		s.mockConfig.EXPECT().Reads().Return(reads).Once()

		db, err := s.gorm.Build()
		s.NoError(err)
		s.NotNil(db)
		s.NoError(docker.Shutdown())
	})

	s.Run("failed to generate DSN", func() {
		s.mockConfig.EXPECT().Writes().Return([]contracts.FullConfig{
			{},
		}).Once()
		s.mockConfig.EXPECT().Reads().Return([]contracts.FullConfig{}).Once()

		db, err := s.gorm.Build()
		s.Equal(FailedToGenerateDSN, err)
		s.Nil(db)
	})

	s.Run("not found database configuration", func() {
		s.mockConfig.EXPECT().Writes().Return([]contracts.FullConfig{}).Once()
		s.mockConfig.EXPECT().Reads().Return([]contracts.FullConfig{}).Once()

		db, err := s.gorm.Build()
		s.Equal(ConfigNotFound, err)
		s.Nil(db)
	})
}

func (s *GormTestSuite) TestDNS() {
	tests := []struct {
		name     string
		config   contracts.FullConfig
		expected string
	}{
		{
			name: "with dsn",
			config: contracts.FullConfig{
				Config: contracts.Config{
					Dsn: "postgres://user:pass@localhost:5432/dbname",
				},
			},
			expected: "postgres://user:pass@localhost:5432/dbname",
		},
		{
			name: "without dsn",
			config: contracts.FullConfig{
				Config: contracts.Config{
					Host:     "localhost",
					Port:     5432,
					Database: "testdb",
					Username: "user",
					Password: "pass",
					Schema:   "public",
				},
				Sslmode:  "disable",
				Timezone: "UTC",
			},
			expected: "postgres://user:pass@localhost:5432/testdb?sslmode=disable&timezone=UTC&search_path=public",
		},
		{
			name: "invalid config",
			config: contracts.FullConfig{
				Config: contracts.Config{
					Database: "testdb",
				},
			},
			expected: "",
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			result := s.gorm.dns(test.config)
			assert.Equal(s.T(), test.expected, result)
		})
	}
}
