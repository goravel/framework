package database

import (
	"github.com/goravel/framework/contracts/database/seeder"
)

type Seeder struct {
}

type SeederFacade struct {
	Seeders []seeder.Seeder
	Called  []string
}

func NewSeederFacade() seeder.Facade {
	return &SeederFacade{}
}

func (s *SeederFacade) Register(seeders []seeder.Seeder) {
	s.Seeders = append(s.Seeders, seeders...)
}

func (s *SeederFacade) GetSeeder(name string) seeder.Seeder {
	var seeder seeder.Seeder
	for _, item := range s.Seeders {
		if item.Signature() == name {
			seeder = item
			break
		}
	}

	return seeder
}

func (s *SeederFacade) GetSeeders() []seeder.Seeder {
	return s.Seeders
}

// Call executes the specified seeder(s).
func (s *SeederFacade) Call(seeders []seeder.Seeder) error {
	for _, seeder := range seeders {
		signature := seeder.Signature()

		err := seeder.Run()
		if err != nil {
			return err
		}

		if !contains(s.Called, signature) {
			s.Called = append(s.Called, signature)
		}
	}
	return nil
}

// contains checks if a string exists in a slice.
func contains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

// CallOnce executes the specified seeder(s) only if they haven't been executed before.
func (s *SeederFacade) CallOnce(seeders []seeder.Seeder) error {
	for _, item := range seeders {
		signature := item.Signature()

		if contains(s.Called, signature) {
			return nil
		}

		err := s.Call([]seeder.Seeder{item})
		if err != nil {
			return err
		}
	}
	return nil
}
