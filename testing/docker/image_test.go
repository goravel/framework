package docker

import (
	"testing"

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
	driver := NewImageDriver(s.image)
	assert.NotNil(s.T(), driver)
	assert.Equal(s.T(), s.image, driver.image)
}

func (s *ImageDriverTestSuite) TestBuildConfigShutdown() {
	driver := NewImageDriver(s.image)
	err := driver.Build()
	s.NoError(err)

	config := driver.Config()
	s.NotEmpty(config.ContainerID)
	s.True(config.ExposedPorts[6379] > 0)

	err = driver.Shutdown()
	s.NoError(err)
}
