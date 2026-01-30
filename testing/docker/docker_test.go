package docker

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/errors"
	mockscache "github.com/goravel/framework/mocks/cache"
	mocksconfig "github.com/goravel/framework/mocks/config"
	mocksconsole "github.com/goravel/framework/mocks/console"
	mocksorm "github.com/goravel/framework/mocks/database/orm"
	mocksprocess "github.com/goravel/framework/mocks/process"
	mocksdocker "github.com/goravel/framework/mocks/testing/docker"
)

type DockerTestSuite struct {
	suite.Suite
	mockArtisan     *mocksconsole.Artisan
	mockCache       *mockscache.Cache
	mockConfig      *mocksconfig.Config
	mockOrm         *mocksorm.Orm
	mockProcess     *mocksprocess.Process
	mockCacheDriver *mockscache.Driver
	mockDocker      *mocksdocker.CacheDriver
	docker          *Docker
}

func TestDockerTestSuite(t *testing.T) {
	suite.Run(t, new(DockerTestSuite))
}

func (s *DockerTestSuite) SetupTest() {
	s.mockArtisan = mocksconsole.NewArtisan(s.T())
	s.mockCache = mockscache.NewCache(s.T())
	s.mockConfig = mocksconfig.NewConfig(s.T())
	s.mockOrm = mocksorm.NewOrm(s.T())
	s.mockProcess = mocksprocess.NewProcess(s.T())
	s.mockCacheDriver = mockscache.NewDriver(s.T())
	s.mockDocker = mocksdocker.NewCacheDriver(s.T())
	s.docker = NewDocker(s.mockArtisan, s.mockCache, s.mockConfig, s.mockOrm, s.mockProcess)
}

func (s *DockerTestSuite) TestCache() {
	tests := []struct {
		name    string
		store   []string
		setup   func()
		wantErr error
	}{
		{
			name:  "success with default store",
			store: []string{},
			setup: func() {
				s.mockConfig.EXPECT().GetString("cache.default").Return("redis").Once()
				s.mockCache.EXPECT().Store("redis").Return(s.mockCacheDriver).Once()
				s.mockCacheDriver.EXPECT().Docker().Return(s.mockDocker, nil).Once()
			},
			wantErr: nil,
		},
		{
			name:  "success with specified store",
			store: []string{"memcached"},
			setup: func() {
				s.mockCache.EXPECT().Store("memcached").Return(s.mockCacheDriver).Once()
				s.mockCacheDriver.EXPECT().Docker().Return(s.mockDocker, nil).Once()
			},
			wantErr: nil,
		},
		{
			name:  "error when cache is nil",
			store: []string{},
			setup: func() {
				s.docker.cache = nil
			},
			wantErr: errors.CacheFacadeNotSet,
		},
		{
			name:  "error when docker returns error",
			store: []string{"redis"},
			setup: func() {
				s.mockCache.EXPECT().Store("redis").Return(s.mockCacheDriver).Once()
				s.mockCacheDriver.EXPECT().Docker().Return(nil, assert.AnError).Once()
			},
			wantErr: assert.AnError,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.SetupTest()
			tt.setup()
			gotDriver, err := s.docker.Cache(tt.store...)

			if tt.wantErr != nil {
				s.EqualError(err, tt.wantErr.Error())
				s.Nil(gotDriver)
			} else {
				s.Nil(err)
				s.Equal(s.mockDocker, gotDriver)
			}
		})
	}
}
