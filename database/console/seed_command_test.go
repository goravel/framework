package console

import (
	"testing"

	"github.com/stretchr/testify/suite"

	configmocks "github.com/goravel/framework/contracts/config/mocks"
	consolemocks "github.com/goravel/framework/contracts/console/mocks"
	"github.com/goravel/framework/contracts/database/seeder"
	seedermocks "github.com/goravel/framework/contracts/database/seeder/mocks"
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
	s.mockContext.On("OptionBool", "force").Return(false).Once()
	s.mockConfig.On("Env", "APP_ENV").Return("development").Once()
	s.mockContext.On("OptionSlice", "seeder").Return([]string{"mock", "mock2"}).Once()
	s.mockFacade.On("GetSeeder", "mock").Return(&MockSeeder{}).Once()
	s.mockFacade.On("GetSeeder", "mock2").Return(&MockSeeder2{}).Once()
	s.mockFacade.On("Call", []seeder.Seeder{&MockSeeder{}, &MockSeeder2{}}).Return(nil).Once()
	s.NoError(s.seedCommand.Handle(s.mockContext))

	s.mockConfig.AssertExpectations(s.T())
	s.mockContext.AssertExpectations(s.T())
	s.mockFacade.AssertExpectations(s.T())
}

func (s *SeedCommandTestSuite) TestConfirmToProceed() {
	s.mockConfig.On("Env", "APP_ENV").Return("production").Once()
	err := s.seedCommand.ConfirmToProceed(true)
	s.NoError(err)

	s.mockContext.On("OptionBool", "force").Return(false).Once()
	err = s.seedCommand.ConfirmToProceed(false)
	s.EqualError(err, "application in production use --force to run this command")
}

func (s *SeedCommandTestSuite) TestGetSeeders() {
	// write logic for GetSeeders function here
	seeders := []seeder.Seeder{
		&MockSeeder{},
		&MockSeeder2{},
	}
	names := []string{"mock", "mock2"}

	s.mockFacade.On("GetSeeder", "mock").Return(&MockSeeder{}).Once()
	s.mockFacade.On("GetSeeder", "mock2").Return(&MockSeeder2{}).Once()

	result, err := s.seedCommand.GetSeeders(names)
	s.NoError(err)
	s.ElementsMatch(seeders, result)

	s.mockFacade.AssertExpectations(s.T())
}

type MockSeeder struct {
}

func (m *MockSeeder) Run() error {
	return nil
}

func (m *MockSeeder) Signature() string {
	return "mock"
}

type MockSeeder2 struct {
}

func (m *MockSeeder2) Run() error {
	return nil
}

func (m *MockSeeder2) Signature() string {
	return "mock2"
}
