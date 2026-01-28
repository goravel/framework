package docker

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	contractsdocker "github.com/goravel/framework/contracts/testing/docker"
	"github.com/goravel/framework/support/env"
)

type ImageDriverTestSuite struct {
	suite.Suite
	image contractsdocker.Image
}

func TestImageDriverTestSuite(t *testing.T) {
	if env.IsWindows() {
		t.Skip("Skipping image driver test on Windows")
	}

	suite.Run(t, new(ImageDriverTestSuite))
}

func (s *ImageDriverTestSuite) SetupTest() {
	s.image = contractsdocker.Image{
		Repository:   "redis",
		Tag:          "latest",
		ExposedPorts: []string{"6379"},
	}
}

func (s *ImageDriverTestSuite) TestNewImageDriver() {
	driver := NewImageDriver(s.image, nil)
	assert.NotNil(s.T(), driver)
	assert.Equal(s.T(), s.image, driver.image)
}

func (s *ImageDriverTestSuite) TestBuildConfigReadyShutdown() {
	driver := NewImageDriver(s.image, nil)
	err := driver.Build()
	s.NoError(err)

	config := driver.Config()
	s.NotEmpty(config.ContainerID)
	s.True(len(config.ExposedPorts) > 0)

	err = driver.Shutdown()
	s.NoError(err)
}

func (s *ImageDriverTestSuite) TestReady() {
	driver := NewImageDriver(s.image, nil)

	err := driver.Ready(func() error {
		return nil
	})
	s.NoError(err)

	err = driver.Ready(func() error {
		return errors.New("error")
	}, 3*time.Second)
	s.Error(err)
}
