package seeder

import (
	"github.com/gookit/color"

	"github.com/goravel/framework/contracts/database/seeder"
)

type Seeder struct {
	Seeders []seeder.Seeder
	Called  []string
}

func NewSeeder() *Seeder {
	return &Seeder{}
}

func (s *Seeder) Register(seeders []seeder.Seeder) {
	existingSignatures := make(map[string]bool)

	for _, seeder := range seeders {
		signature := seeder.Signature()

		if existingSignatures[signature] {
			color.Redf("Duplicate seeder signature: %s in %T\n", signature, seeder)
		} else {
			existingSignatures[signature] = true
			s.Seeders = append(s.Seeders, seeder)
		}
	}
}

func (s *Seeder) GetSeeder(name string) seeder.Seeder {
	var seeder seeder.Seeder
	for _, item := range s.Seeders {
		if item.Signature() == name {
			seeder = item
			break
		}
	}

	return seeder
}

func (s *Seeder) GetSeeders() []seeder.Seeder {
	return s.Seeders
}

// Call executes the specified seeder(s).
func (s *Seeder) Call(seeders []seeder.Seeder) error {
	for _, seeder := range seeders {
		signature := seeder.Signature()

		if err := seeder.Run(); err != nil {
			return err
		}

		if !contains(s.Called, signature) {
			s.Called = append(s.Called, signature)
		}
	}
	return nil
}

// CallOnce executes the specified seeder(s) only if they haven't been executed before.
func (s *Seeder) CallOnce(seeders []seeder.Seeder) error {
	for _, item := range seeders {
		signature := item.Signature()

		if contains(s.Called, signature) {
			continue
		}

		if err := s.Call([]seeder.Seeder{item}); err != nil {
			return err
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