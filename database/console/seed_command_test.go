package console

import (
	"errors"
	"testing"

	configmocks "github.com/goravel/framework/contracts/config/mocks"
	"github.com/goravel/framework/contracts/console/command"
	consolemocks "github.com/goravel/framework/contracts/console/mocks"
	"github.com/goravel/framework/contracts/database/seeder"
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
	err := errors.New("application in production use --force to run this command")
	s.mockConfig.On("Env", "APP_ENV").Return("production").Once()
	s.mockContext.On("OptionBool", "force").Return(false).Once()
	assert.EqualError(s.T(), err, s.seedCommand.ConfirmToProceed(false).Error())

	s.mockConfig.On("Env", "APP_ENV").Return("development").Once()
	s.mockContext.On("Arguments").Return([]string{"mock", "mock2"}).Once()
	s.mockFacade.On("GetSeeder", "mock").Return(&MockSeeder{}).Once()
	s.mockFacade.On("GetSeeder", "mock2").Return(&MockSeeder2{}).Once()
	s.mockFacade.On("Call", []seeder.Seeder{&MockSeeder{}, &MockSeeder2{}}).Return(nil).Once()
	assert.NoError(s.T(), s.seedCommand.Handle(s.mockContext))

	s.mockConfig.AssertExpectations(s.T())
	s.mockContext.AssertExpectations(s.T())
	s.mockFacade.AssertExpectations(s.T())
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
	seeders := []seeder.Seeder{
		&MockSeeder{},
		&MockSeeder2{},
	}
	names := []string{"mock", "mock2"}

	for _, name := range names {
		switch name {
		case "mock":
			s.mockFacade.On("GetSeeder", name).Return(&MockSeeder{}).Once()
		case "mock2":
			s.mockFacade.On("GetSeeder", name).Return(&MockSeeder2{}).Once()
		default:
			assert.Fail(s.T(), "Unknown seeder name: "+name)
		}
	}

	result, err := s.seedCommand.GetSeeders(names)
	assert.NoError(s.T(), err)
	assert.ElementsMatch(s.T(), seeders, result)

	s.mockFacade.AssertExpectations(s.T())
}

func (s *SeedCommandTestSuite) TestSignature() {
	assert.Equal(s.T(), "db:seed", s.seedCommand.Signature())
}

func (s *SeedCommandTestSuite) TestDescription() {
	assert.Equal(s.T(), "Seed the database with records", s.seedCommand.Description())
}

func (s *SeedCommandTestSuite) TestExtend() {
	expected := command.Extend{
		Category: "db",
		Flags: []command.Flag{
			&command.BoolFlag{
				Name:    "force",
				Aliases: []string{"f"},
				Usage:   "force the operation to run when in production",
			},
		},
	}

	assert.Equal(s.T(), expected, s.seedCommand.Extend())
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
