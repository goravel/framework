package testing

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	mocksconsole "github.com/goravel/framework/mocks/console"
)

type TestCaseSuite struct {
	suite.Suite
	mockArtisan *mocksconsole.Artisan
	testCase    *TestCase
}

func TestTestCaseSuite(t *testing.T) {
	suite.Run(t, new(TestCaseSuite))
}

// SetupTest will run before each test in the suite.
func (s *TestCaseSuite) SetupTest() {
	s.mockArtisan = mocksconsole.NewArtisan(s.T())
	s.testCase = &TestCase{}
	artisanFacade = s.mockArtisan
}

func (s *TestCaseSuite) TestSeed() {
	s.mockArtisan.On("Call", "db:seed").Return(nil).Once()
	s.testCase.Seed()

	s.mockArtisan.On("Call", "db:seed --seeder mock").Return(nil).Once()
	s.testCase.Seed(&MockSeeder{})

	s.Panics(func() {
		s.mockArtisan.On("Call", "db:seed").Return(assert.AnError).Once()
		s.testCase.Seed()
	})
}

func (s *TestCaseSuite) TestRefreshDatabase() {
	s.mockArtisan.On("Call", "migrate:refresh").Return(nil).Once()
	s.testCase.RefreshDatabase()

	s.Panics(func() {
		s.mockArtisan.On("Call", "migrate:refresh").Return(assert.AnError).Once()
		s.testCase.RefreshDatabase()
	})
}

type MockSeeder struct{}

func (m *MockSeeder) Signature() string {
	return "mock"
}

func (m *MockSeeder) Run() error {
	return nil
}
