package testing

import (
	"testing"

	"github.com/stretchr/testify/suite"

	consolemocks "github.com/goravel/framework/mocks/console"
)

type TestCaseSuite struct {
	suite.Suite
	mockArtisan *consolemocks.Artisan
	testCase    *TestCase
}

func TestTestCaseSuite(t *testing.T) {
	suite.Run(t, new(TestCaseSuite))
}

// SetupTest will run before each test in the suite.
func (s *TestCaseSuite) SetupTest() {
	s.mockArtisan = &consolemocks.Artisan{}
	s.testCase = &TestCase{}
	artisanFacades = s.mockArtisan
}

func (s *TestCaseSuite) TestSeed() {
	s.mockArtisan.On("Call", "db:seed").Once()
	s.testCase.Seed()

	s.mockArtisan.On("Call", "db:seed --seeder mock").Once()
	s.testCase.Seed(&MockSeeder{})

	s.mockArtisan.AssertExpectations(s.T())
}

func (s *TestCaseSuite) TestRefreshDatabase() {
	s.mockArtisan.On("Call", "migrate:refresh").Once()
	s.testCase.RefreshDatabase()

	s.mockArtisan.AssertExpectations(s.T())
}

type MockSeeder struct{}

func (m *MockSeeder) Signature() string {
	return "mock"
}

func (m *MockSeeder) Run() error {
	return nil
}
