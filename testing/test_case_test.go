package testing

import (
	"testing"

	"github.com/stretchr/testify/suite"

	consolemocks "github.com/goravel/framework/contracts/console/mocks"
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
	s.mockArtisan = consolemocks.NewArtisan(s.T())
	s.testCase = &TestCase{}
	artisanFacades = s.mockArtisan
}

func (s *TestCaseSuite) TestSeed() {
	s.mockArtisan.On("Call", "db:seed").Once()
	s.testCase.Seed()

	s.mockArtisan.On("Call", "db:seed --seeder mock").Once()
	s.testCase.Seed(&MockSeeder{})
}

func (s *TestCaseSuite) TestRefreshDatabase() {
	s.mockArtisan.On("Call", "migrate:refresh").Once()
	s.testCase.RefreshDatabase()
}

type MockSeeder struct{}

func (m *MockSeeder) Signature() string {
	return "mock"
}

func (m *MockSeeder) Run() error {
	return nil
}
