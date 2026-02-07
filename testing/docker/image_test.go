package docker

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	contractsdocker "github.com/goravel/framework/contracts/testing/docker"
	"github.com/goravel/framework/errors"
	mocksprocess "github.com/goravel/framework/mocks/process"
	"github.com/goravel/framework/support/env"
)

type ImageDriverTestSuite struct {
	suite.Suite
	image       contractsdocker.Image
	mockProcess *mocksprocess.Process
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
	s.mockProcess = mocksprocess.NewProcess(s.T())
}

func (s *ImageDriverTestSuite) TestNewImageDriver() {
	driver := NewImageDriver(s.image, s.mockProcess)
	assert.NotNil(s.T(), driver)
	assert.Equal(s.T(), s.image, driver.image)
}

func (s *ImageDriverTestSuite) TestBuildConfigShutdown() {
	s.Run("happty path", func() {
		containerID := "mocked-container-id"
		mockProcessResult := mocksprocess.NewResult(s.T())
		mockProcessResult.EXPECT().Failed().Return(false).Once()
		mockProcessResult.EXPECT().Output().Return(containerID).Once()
		s.mockProcess.EXPECT().Run(mock.MatchedBy(func(command string) bool {
			return strings.Contains(command, "docker run --rm -d -p ") && strings.Contains(command, ":6379 redis:latest")
		})).Return(mockProcessResult).Once()

		driver := NewImageDriver(s.image, s.mockProcess)
		err := driver.Build()
		s.NoError(err)

		config := driver.Config()

		s.Equal(containerID, config.ContainerID)
		s.True(len(config.ExposedPorts) > 0)

		mockProcessResult.EXPECT().Failed().Return(false).Once()
		s.mockProcess.EXPECT().Run("docker stop " + containerID).Return(mockProcessResult).Once()

		err = driver.Shutdown()
		s.NoError(err)
	})

	s.Run("failed to shutdown", func() {
		containerID := "mocked-container-id"
		mockProcessResult := mocksprocess.NewResult(s.T())
		mockProcessResult.EXPECT().Failed().Return(false).Once()
		mockProcessResult.EXPECT().Output().Return(containerID).Once()
		s.mockProcess.EXPECT().Run(mock.MatchedBy(func(command string) bool {
			return strings.Contains(command, "docker run --rm -d -p ") && strings.Contains(command, ":6379 redis:latest")
		})).Return(mockProcessResult).Once()

		driver := NewImageDriver(s.image, s.mockProcess)
		err := driver.Build()
		s.NoError(err)

		config := driver.Config()

		s.Equal(containerID, config.ContainerID)
		s.True(len(config.ExposedPorts) > 0)

		mockProcessResult.EXPECT().Failed().Return(true).Once()
		mockProcessResult.EXPECT().Error().Return(assert.AnError).Once()
		s.mockProcess.EXPECT().Run("docker stop " + containerID).Return(mockProcessResult).Once()

		err = driver.Shutdown()
		s.Equal(errors.TestingImageStopFailed.Args(s.image.Repository, assert.AnError), err)
	})

	s.Run("containerID is empty", func() {
		mockProcessResult := mocksprocess.NewResult(s.T())
		mockProcessResult.EXPECT().Failed().Return(false).Once()
		mockProcessResult.EXPECT().Output().Return("").Once()
		s.mockProcess.EXPECT().Run(mock.MatchedBy(func(command string) bool {
			return strings.Contains(command, "docker run --rm -d -p ") && strings.Contains(command, ":6379 redis:latest")
		})).Return(mockProcessResult).Once()

		driver := NewImageDriver(s.image, s.mockProcess)
		err := driver.Build()

		s.Equal(errors.TestingImageNoContainerId.Args(s.image.Repository), err)
	})

	s.Run("failed to shutdown", func() {
		mockProcessResult := mocksprocess.NewResult(s.T())
		mockProcessResult.EXPECT().Failed().Return(true).Once()
		mockProcessResult.EXPECT().Error().Return(assert.AnError).Once()
		s.mockProcess.EXPECT().Run(mock.MatchedBy(func(command string) bool {
			return strings.Contains(command, "docker run --rm -d -p ") && strings.Contains(command, ":6379 redis:latest")
		})).Return(mockProcessResult).Once()

		driver := NewImageDriver(s.image, s.mockProcess)
		err := driver.Build()

		s.Equal(errors.TestingImageBuildFailed.Args(s.image.Repository, assert.AnError), err)
	})
}

func (s *ImageDriverTestSuite) TestReady() {
	driver := NewImageDriver(s.image, s.mockProcess)

	err := driver.Ready(func() error {
		return nil
	})
	s.NoError(err)

	err = driver.Ready(func() error {
		return errors.New("error")
	}, 3*time.Second)
	s.Error(err)
}
