package console

import (
	"errors"
	"testing"

	configmocks "github.com/goravel/framework/contracts/config/mocks"
	"github.com/goravel/framework/contracts/console/command"
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
	err := errors.New("application in production use --force to run this command")
	// Test ConfirmToProceed error
	s.mockConfig.On("Env", "APP_ENV").Return("production").Once()
	s.mockContext.On("OptionBool", "force").Return(false).Once()
	assert.EqualError(s.T(), err, s.seedCommand.ConfirmToProceed(false).Error())

	// Test GetSeeders error
	// err = errors.New("no seeder of foo found")
	// s.mockConfig.On("Env", "APP_ENV").Return("development").Once()
	// s.mockContext.On("OptionBool", "force").Return(true).Once()
	// s.mockContext.On("Arguments").Return([]string{"foo"}).Once()
	// s.mockFacade.On("GetSeeder", "foo").Return(nil).Once()
	// assert.EqualError(s.T(), err, s.seedCommand.Handle(s.mockContext).Error())

	// // Test success case
	// s.mockConfig.On("Env", "APP_ENV").Return("development").Once()
	// s.mockContext.On("OptionBool", "force").Return(true).Once()
	// s.mockContext.On("Arguments").Return([]string{}).Once()
	// s.mockFacade.On("GetSeeders").Return([]seeder.Seeder{}).Once()
	// s.mockFacade.On("Call", mock.AnythingOfType("[]seeder.Seeder")).Return(nil).Once()
	// err = s.seedCommand.Handle(s.mockContext)
	// assert.NoError(s.T(), err)

	// s.mockConfig.AssertExpectations(s.T())
	// s.mockContext.AssertExpectations(s.T())
	// s.mockFacade.AssertExpectations(s.T())
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
	// seeders := []seedermocks.Seeder{
	// 	&seedermocks.Seeder{Name: "foo"},
	// 	&seedermocks.Seeder{Name: "bar"},
	// }
	// names := []string{"foo", "bar"}

	// for _, name := range names {
	// 	s.mockFacade.On("GetSeeder", name).Return(&seedermocks.Seeder{Name: name}).Once()
	// }

	// result, err := s.seedCommand.GetSeeders(names)
	// assert.NoError(s.T(), err)
	// assert.Equal(s.T(), seeders, result)

	// s.mockFacade.AssertExpectations(s.T())
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
