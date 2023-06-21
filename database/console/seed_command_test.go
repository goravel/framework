package console

import (
	"testing"

	configmocks "github.com/goravel/framework/contracts/config/mocks"
	consolemocks "github.com/goravel/framework/contracts/console/mocks"
	seedermocks "github.com/goravel/framework/contracts/database/seeder/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type SeedCommandTestSuite struct {
	suite.Suite
	mockConfig  *configmocks.Config
	mockFacade  *seedermocks.Facade
	mockContext *consolemocks.Context
	seedCommand *SeedCommand
}

func TestSeedCommandTestSuite(t *testing.T) {
	suite.Run(t, new(SeedCommandTestSuite))
}

func (s *SeedCommandTestSuite) SetupTest() {
	s.mockConfig = &configmocks.Config{}
	s.mockFacade = &seedermocks.Facade{}
	s.mockContext = &consolemocks.Context{}
	s.seedCommand = NewSeedCommand(s.mockConfig, s.mockFacade)
}

func (s *SeedCommandTestSuite) TestHandle() {
	// write test for handle function here
}

func (s *SeedCommandTestSuite) TestConfirmToProceed() {
	s.mockConfig.On("Env", "APP_ENV").Return("production").Once()
	s.mockContext.On("OptionBool", "force").Return(true).Once()
	err := s.seedCommand.ConfirmToProceed(true)
	assert.NoError(s.T(), err)

	s.mockContext.On("OptionBool", "force").Return(false).Once()

	err = s.seedCommand.ConfirmToProceed(false)
	assert.EqualError(s.T(), err, "application in production use --force to run this command")
}

func (s *SeedCommandTestSuite) TestGetSeeders() {
	// write logic for GetSeeders function here
}
