package docker

import (
	"testing"

	"github.com/stretchr/testify/suite"

	contractstesting "github.com/goravel/framework/contracts/testing"
	"github.com/goravel/framework/support/env"
)

type ContainerManagerTestSuite struct {
	suite.Suite
	container *Container
}

func TestContainerMangerTestSuite(t *testing.T) {
	if env.IsWindows() {
		t.Skip("Skip test that using Docker")
	}

	suite.Run(t, new(ContainerManagerTestSuite))
}

func (s *ContainerManagerTestSuite) SetupTest() {
	s.container = NewContainer(&testDatabaseDriver{})
}

func (s *ContainerManagerTestSuite) TestAddAndAll() {
	driver := &testDatabaseDriver{}

	s.NoError(s.container.add("test", driver))

	containers, err := s.container.all()
	s.NoError(err)
	s.Len(containers, 1)
	s.Equal(driver.Config(), containers["test"])
}

type testDatabaseDriver struct {
}

func (r *testDatabaseDriver) Build() error {
	return nil
}

func (r *testDatabaseDriver) Config() contractstesting.DatabaseConfig {
	return contractstesting.DatabaseConfig{
		ContainerID: "container_id",
		Host:        "host",
		Port:        1234,
		Database:    "database",
		Username:    "username",
		Password:    "password",
	}
}

func (r *testDatabaseDriver) Database(name string) (contractstesting.DatabaseDriver, error) {
	return nil, nil
}

func (r *testDatabaseDriver) Driver() string {
	return "test"
}

func (r *testDatabaseDriver) Fresh() error {
	return nil
}

func (r *testDatabaseDriver) Image(image contractstesting.Image) {

}

func (r *testDatabaseDriver) Ready() error {
	return nil
}

func (r *testDatabaseDriver) Reuse(containerID string, port int) error {
	return nil
}

func (r *testDatabaseDriver) Shutdown() error {
	return nil
}
