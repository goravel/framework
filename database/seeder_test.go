package database

import (
	"log"
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/database/seeder"
)

type SeederTestSuite struct {
	suite.Suite
	seederFacade *SeederFacade
}

func TestSeederTestSuite(t *testing.T) {
	suite.Run(t, new(SeederTestSuite))
}

func (s *SeederTestSuite) SetupTest() {
	s.seederFacade = NewSeederFacade()
}

func (s *SeederTestSuite) TestRegister() {
	s.seederFacade.Register([]seeder.Seeder{
		&MockSeeder{},
		&MockSeeder2{},
	})

	s.Len(s.seederFacade.GetSeeders(), 2)
}

func (s *SeederTestSuite) TestGetSeeder() {
	s.seederFacade.Register([]seeder.Seeder{
		&MockSeeder{},
		&MockSeeder2{},
	})

	seeder := s.seederFacade.GetSeeder("mock")
	s.NotNil(seeder)
	s.Equal("mock", seeder.Signature())
}

func (s *SeederTestSuite) TestGetSeeders() {
	s.seederFacade.Register([]seeder.Seeder{
		&MockSeeder{},
		&MockSeeder2{},
	})

	seeders := s.seederFacade.GetSeeders()
	s.Len(seeders, 2)
}

func (s *SeederTestSuite) TestCall() {
	s.seederFacade.Register([]seeder.Seeder{
		&MockSeeder{},
		&MockSeeder2{},
	})

	s.NoError(s.seederFacade.Call([]seeder.Seeder{
		&MockSeeder{},
		&MockSeeder2{},
	}))

	s.Len(s.seederFacade.Called, 2)
}

func (s *SeederTestSuite) TestCallOnce() {
	s.seederFacade.Register([]seeder.Seeder{
		&MockSeeder{},
		&MockSeeder2{},
	})

	s.NoError(s.seederFacade.CallOnce([]seeder.Seeder{
		&MockSeeder{},
		&MockSeeder2{},
	}))

	s.Len(s.seederFacade.Called, 2)
}

func (s *SeederTestSuite) TestCallOnceWithCalled() {
	s.seederFacade.Register([]seeder.Seeder{
		&MockSeeder{},
		&MockSeeder2{},
	})

	s.seederFacade.Called = []string{"mock"}
	s.NoError(s.seederFacade.CallOnce([]seeder.Seeder{
		&MockSeeder2{},
	}))

	log.Println(s.seederFacade.Called)

	s.Len(s.seederFacade.Called, 2)
}

type MockSeeder struct{}

func (m *MockSeeder) Signature() string {
	return "mock"
}

func (m *MockSeeder) Run() error {
	return nil
}

type MockSeeder2 struct{}

func (m *MockSeeder2) Signature() string {
	return "mock2"
}

func (m *MockSeeder2) Run() error {
	return nil
}
