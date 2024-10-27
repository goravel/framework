package console

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/database/seeder"
	"github.com/goravel/framework/errors"
	mocksconfig "github.com/goravel/framework/mocks/config"
	mocksconsole "github.com/goravel/framework/mocks/console"
	mocksseeder "github.com/goravel/framework/mocks/database/seeder"
)

type SeedCommandTestSuite struct {
	suite.Suite
	mockConfig  *mocksconfig.Config
	mockFacade  *mocksseeder.Facade
	mockContext *mocksconsole.Context
	seedCommand *SeedCommand
}

func TestSeedCommandTestSuite(t *testing.T) {
	suite.Run(t, new(SeedCommandTestSuite))
}

func (s *SeedCommandTestSuite) SetupTest() {
	s.mockConfig = mocksconfig.NewConfig(s.T())
	s.mockFacade = mocksseeder.NewFacade(s.T())
	s.mockContext = mocksconsole.NewContext(s.T())
	s.seedCommand = NewSeedCommand(s.mockConfig, s.mockFacade)
}

func (s *SeedCommandTestSuite) TestHandle() {
	tests := []struct {
		name  string
		setup func()
	}{
		{
			name: "Success",
			setup: func() {
				s.mockContext.EXPECT().OptionBool("force").Return(false).Once()
				s.mockConfig.EXPECT().GetString("app.env").Return("development").Once()
				s.mockContext.EXPECT().OptionSlice("seeder").Return([]string{"mock", "mock2"}).Once()
				s.mockFacade.EXPECT().GetSeeder("mock").Return(&MockSeeder{}).Once()
				s.mockFacade.EXPECT().GetSeeder("mock2").Return(&MockSeeder2{}).Once()
				s.mockFacade.EXPECT().Call([]seeder.Seeder{&MockSeeder{}, &MockSeeder2{}}).Return(nil).Once()
				s.mockContext.EXPECT().Success("Database seeding completed successfully.").Once()
			},
		},
		{
			name: "Run in production without force",
			setup: func() {
				s.mockContext.EXPECT().OptionBool("force").Return(false).Once()
				s.mockConfig.EXPECT().GetString("app.env").Return("production").Once()
				s.mockContext.EXPECT().Error(errors.DBForceIsRequiredInProduction.Error()).Once()
			},
		},
		{
			name: "Seeder not found",
			setup: func() {
				s.mockContext.EXPECT().OptionBool("force").Return(false).Once()
				s.mockConfig.EXPECT().GetString("app.env").Return("development").Once()
				s.mockContext.EXPECT().OptionSlice("seeder").Return([]string{"mock"}).Once()
				s.mockFacade.EXPECT().GetSeeder("mock").Return(nil).Once()
				s.mockContext.EXPECT().Error(errors.DBSeederNotFound.Args("mock").Error()).Once()
			},
		},
		{
			name: "No seeders found",
			setup: func() {
				s.mockContext.EXPECT().OptionBool("force").Return(false).Once()
				s.mockConfig.EXPECT().GetString("app.env").Return("development").Once()
				s.mockContext.EXPECT().OptionSlice("seeder").Return(nil).Once()
				s.mockFacade.EXPECT().GetSeeders().Return(nil).Once()
				s.mockContext.EXPECT().Success("no seeders found").Once()
			},
		},
		{
			name: "Run seeder failed",
			setup: func() {
				s.mockContext.EXPECT().OptionBool("force").Return(false).Once()
				s.mockConfig.EXPECT().GetString("app.env").Return("development").Once()
				s.mockContext.EXPECT().OptionSlice("seeder").Return([]string{"mock", "mock2"}).Once()
				s.mockFacade.EXPECT().GetSeeder("mock").Return(&MockSeeder{}).Once()
				s.mockFacade.EXPECT().GetSeeder("mock2").Return(&MockSeeder2{}).Once()
				s.mockFacade.EXPECT().Call([]seeder.Seeder{&MockSeeder{}, &MockSeeder2{}}).Return(assert.AnError).Once()
				s.mockContext.EXPECT().Error(errors.DBFailToRunSeeder.Args(assert.AnError).Error()).Once()
			},
		},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			s.SetupTest()
			test.setup()

			s.NoError(s.seedCommand.Handle(s.mockContext))
		})
	}
}

func (s *SeedCommandTestSuite) TestConfirmToProceed() {
	err := s.seedCommand.ConfirmToProceed(true)
	s.NoError(err)

	s.mockConfig.EXPECT().GetString("app.env").Return("production").Once()
	err = s.seedCommand.ConfirmToProceed(false)
	s.ErrorIs(err, errors.DBForceIsRequiredInProduction)
}

func (s *SeedCommandTestSuite) TestGetSeeders() {
	seeders := []seeder.Seeder{
		&MockSeeder{},
		&MockSeeder2{},
	}
	names := []string{"mock", "mock2"}

	s.mockFacade.EXPECT().GetSeeder("mock").Return(&MockSeeder{}).Once()
	s.mockFacade.EXPECT().GetSeeder("mock2").Return(&MockSeeder2{}).Once()

	result, err := s.seedCommand.GetSeeders(names)
	s.NoError(err)
	s.ElementsMatch(seeders, result)
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
