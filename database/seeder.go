package database

import (
	"fmt"
	"log"
	"reflect"

	"github.com/goravel/framework/contracts/database/seeder"
)

type Seeder struct {
	Called []string
}

type Facade struct {
	Seeders []seeder.Seeder
}

func NewSeederFacade() seeder.Facade {
	return &Facade{}
}

func (f *Facade) Register(seeders []seeder.Seeder) {
	f.Seeders = append(f.Seeders, seeders...)
}

func (f *Facade) GetSeeder(name string) seeder.Seeder {
	var seeder seeder.Seeder
	for _, item := range f.Seeders {
		itemType := reflect.TypeOf(item).Elem()
		if itemType.String() == name {
			seeder = item
			break
		}
	}
	return seeder
}

// Call executes the specified seeder(s).
// Example usage:
//
//	seeder := &Seeder{}
//	seeder.Call([]seeder.Seeder{&UserSeeder{}})
//	seeder.Call([]seeder.Seeder{&UserSeeder{}, &PostSeeder{}})
//
// Parameters:
//   - seeders ([]seeder.Seeder): The seeder class or a slice of seeder classes to execute.
//
// Returns:
//   - error: An error if the execution fails.
func (s *Seeder) Call(seeders []seeder.Seeder) error {
	for _, seeder := range seeders {
		name := fmt.Sprintf("%T", seeder)

		if contains(s.Called, name) {
			continue
		}

		err := seeder.Run()
		if err != nil {
			log.Println("Error executing seeder:", err)
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
//
// Example usage:
//
//	seeder := &Seeder{}
//	seeder.CallOnce([]seeder.Seeder{&UserSeeder{}})
//	seeder.CallOnce([]seeder.Seeder{&UserSeeder{}, &PostSeeder{}})
//
// Parameters:
//   - seeders ([]seeder.Seeder): The seeder class or a slice of seeder classes to execute.
//
// Returns:
//   - error: An error if the execution fails.
func (s *Seeder) CallOnce(seeders []seeder.Seeder) error {
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
