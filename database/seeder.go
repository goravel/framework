package database

import (
	"fmt"
	"reflect"

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
		itemType := reflect.TypeOf(item).Elem()
		if itemType.String() == name {
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
		name := fmt.Sprintf("%T", seeder)

		if contains(s.Called, name) {
			continue
		}

		err := seeder.Run()
		if err != nil {
			return err
		}

		s.Called = append(s.Called, name)
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
	seederType := reflect.TypeOf(seeders)
	seederTypeName := seederType.String()
	seederPointerTypeName := "*" + seederTypeName

	for _, called := range s.Called {
		if called == seederTypeName || called == seederPointerTypeName {
			return nil
		}
	}

	return s.Call(seeders)
}
