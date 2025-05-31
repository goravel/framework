package queue

import (
	"testing"

	"github.com/stretchr/testify/suite"

	contractsqueue "github.com/goravel/framework/contracts/queue"
	"github.com/goravel/framework/errors"
	mocksdb "github.com/goravel/framework/mocks/database/db"
	mocksfoundation "github.com/goravel/framework/mocks/foundation"
	mocksqueue "github.com/goravel/framework/mocks/queue"
)

type DriverCreatorTestSuite struct {
	suite.Suite
	mockConfig    *mocksqueue.Config
	mockDB        *mocksdb.DB
	mockJobStorer *mocksqueue.JobStorer
	mockJson      *mocksfoundation.Json
	driverCreator *DriverCreator
}

func TestDriverCreatorTestSuite(t *testing.T) {
	suite.Run(t, new(DriverCreatorTestSuite))
}

func (s *DriverCreatorTestSuite) SetupTest() {
	s.mockConfig = mocksqueue.NewConfig(s.T())
	s.mockDB = mocksdb.NewDB(s.T())
	s.mockJobStorer = mocksqueue.NewJobStorer(s.T())
	s.mockJson = mocksfoundation.NewJson(s.T())
	s.driverCreator = &DriverCreator{
		config:    s.mockConfig,
		db:        s.mockDB,
		jobStorer: s.mockJobStorer,
		json:      s.mockJson,
	}
}

func (s *DriverCreatorTestSuite) TestNewDriverCreator() {
	creator := NewDriverCreator(s.mockConfig, s.mockDB, s.mockJobStorer, s.mockJson, nil)
	s.NotNil(creator)
	s.Equal(s.mockConfig, creator.config)
	s.Equal(s.mockDB, creator.db)
	s.Equal(s.mockJobStorer, creator.jobStorer)
	s.Equal(s.mockJson, creator.json)
}

func (s *DriverCreatorTestSuite) TestCreate() {
	tests := []struct {
		name        string
		connection  string
		driver      string
		setup       func()
		expectedErr error
	}{
		{
			name:       "sync driver",
			connection: "sync",
			driver:     contractsqueue.DriverSync,
			setup: func() {
				s.mockConfig.EXPECT().Driver("sync").Return(contractsqueue.DriverSync).Once()
			},
			expectedErr: nil,
		},
		{
			name:       "database driver - success",
			connection: "database",
			driver:     contractsqueue.DriverDatabase,
			setup: func() {
				s.mockConfig.EXPECT().Driver("database").Return(contractsqueue.DriverDatabase).Once()
				s.mockConfig.EXPECT().GetString("queue.connections.database.connection").Return("mysql").Once()
				s.mockConfig.EXPECT().GetString("queue.connections.database.table", "jobs").Return("jobs").Once()
				s.mockConfig.EXPECT().GetInt("queue.connections.database.retry_after", 60).Return(60).Once()
				s.mockDB.EXPECT().Connection("mysql").Return(s.mockDB).Once()
			},
			expectedErr: nil,
		},
		{
			name:       "database driver - db is nil",
			connection: "database",
			driver:     contractsqueue.DriverDatabase,
			setup: func() {
				s.driverCreator.db = nil
				s.mockConfig.EXPECT().Driver("database").Return(contractsqueue.DriverDatabase).Once()
			},
			expectedErr: errors.QueueInvalidDatabaseConnection.Args("database"),
		},
		{
			name:       "machinery driver",
			connection: "machinery",
			driver:     contractsqueue.DriverMachinery,
			setup: func() {
				s.mockConfig.EXPECT().Driver("machinery").Return(contractsqueue.DriverMachinery).Once()
				s.mockConfig.EXPECT().GetString("queue.connections.machinery.connection").Return("redis").Once()
				s.mockConfig.EXPECT().GetString("database.redis.redis.host").Return("localhost").Once()
				s.mockConfig.EXPECT().GetString("database.redis.redis.password").Return("").Once()
				s.mockConfig.EXPECT().GetInt("database.redis.redis.port").Return(6379).Once()
				s.mockConfig.EXPECT().GetInt("database.redis.redis.database").Return(0).Once()
				s.mockConfig.EXPECT().GetString("app.name").Return("goravel").Once()
				s.mockConfig.EXPECT().GetBool("app.debug").Return(false).Once()
			},
			expectedErr: nil,
		},
		{
			name:       "custom driver - success with driver instance",
			connection: "custom",
			driver:     contractsqueue.DriverCustom,
			setup: func() {
				mockDriver := mocksqueue.NewDriver(s.T())
				s.mockConfig.EXPECT().Driver("custom").Return(contractsqueue.DriverCustom).Once()
				s.mockConfig.EXPECT().Via("custom").Return(mockDriver).Once()
			},
			expectedErr: nil,
		},
		{
			name:       "custom driver - success with driver function",
			connection: "custom",
			driver:     contractsqueue.DriverCustom,
			setup: func() {
				mockDriver := mocksqueue.NewDriver(s.T())
				s.mockConfig.EXPECT().Driver("custom").Return(contractsqueue.DriverCustom).Once()
				s.mockConfig.EXPECT().Via("custom").Return(func() (contractsqueue.Driver, error) {
					return mockDriver, nil
				}).Once()
			},
			expectedErr: nil,
		},
		{
			name:       "custom driver - invalid implementation",
			connection: "custom",
			driver:     contractsqueue.DriverCustom,
			setup: func() {
				s.mockConfig.EXPECT().Driver("custom").Return(contractsqueue.DriverCustom).Once()
				s.mockConfig.EXPECT().Via("custom").Return("invalid").Once()
			},
			expectedErr: errors.QueueDriverInvalid.Args("custom"),
		},
		{
			name:       "unsupported driver",
			connection: "unsupported",
			driver:     "unsupported",
			setup: func() {
				s.mockConfig.EXPECT().Driver("unsupported").Return("unsupported").Once()
			},
			expectedErr: errors.QueueDriverNotSupported.Args("unsupported"),
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			s.SetupTest()
			test.setup()

			driver, err := s.driverCreator.Create(test.connection)

			if test.expectedErr != nil {
				s.Error(err)
				s.Equal(test.expectedErr, err)
				s.Nil(driver)
			} else {
				s.NoError(err)
				s.NotNil(driver)
			}
		})
	}
}
