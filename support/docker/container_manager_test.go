package docker

import (
	"testing"

	"github.com/goravel/framework/support/env"
	"github.com/stretchr/testify/suite"
)

type ContainerManagerTestSuite struct {
	suite.Suite
	containerManager *ContainerManager
}

func TestContainerMangerTestSuite(t *testing.T) {
	if env.IsWindows() {
		t.Skip("Skip test that using Docker")
	}

	suite.Run(t, new(ContainerManagerTestSuite))
}

func (s *ContainerManagerTestSuite) SetupTest() {
	s.containerManager = NewContainerManager()
}

func (s *ContainerManagerTestSuite) TestGet() {
	driver, err := s.containerManager.Get(ContainerTypeMysql)
	s.NoError(err)
	s.NotNil(driver)

	driver, err = s.containerManager.Get(ContainerTypePostgres)
	s.NoError(err)
	s.NotNil(driver)

	driver, err = s.containerManager.Get(ContainerTypeSqlite)
	s.NoError(err)
	s.NotNil(driver)
	s.NoError(driver.Stop())

	driver, err = s.containerManager.Get(ContainerTypeSqlserver)
	s.NoError(err)
	s.NotNil(driver)
}

func (s *ContainerManagerTestSuite) TestAddAndAll() {
	port := 5432
	containerID := "123456"

	postgresDriver := NewPostgresImpl(testDatabase, testUsername, testPassword)
	postgresDriver.port = port
	postgresDriver.containerID = containerID

	sqliteDriver := NewSqliteImpl(testDatabase)

	mysqlDriver := NewMysqlImpl(testDatabase, testUsername, testPassword)
	mysqlDriver.port = port
	mysqlDriver.containerID = containerID

	sqlserverDriver := NewSqlserverImpl(testDatabase, testUsername, testPassword)
	sqlserverDriver.port = port
	sqlserverDriver.containerID = containerID

	s.NoError(s.containerManager.add(ContainerTypePostgres, postgresDriver))
	s.NoError(s.containerManager.add(ContainerTypeSqlite, sqliteDriver))
	s.NoError(s.containerManager.add(ContainerTypeMysql, mysqlDriver))
	s.NoError(s.containerManager.add(ContainerTypeSqlserver, sqlserverDriver))

	containers, err := s.containerManager.all()
	s.NoError(err)
	s.Len(containers, 4)
	s.Equal(postgresDriver.Config(), containers[ContainerTypePostgres])
	s.Equal(sqliteDriver.Config(), containers[ContainerTypeSqlite])
	s.Equal(mysqlDriver.Config(), containers[ContainerTypeMysql])
	s.Equal(sqlserverDriver.Config(), containers[ContainerTypeSqlserver])

	defer func() {
		s.NoError(s.containerManager.Remove())
	}()
}
