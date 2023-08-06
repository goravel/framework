package testing

import (
	"fmt"

	"github.com/goravel/framework/contracts/database/seeder"
)

type TestCase struct {
}

func (receiver *TestCase) Seed(seeds ...seeder.Seeder) {
	command := "db:seed"
	if len(seeds) > 0 {
		command += " --seeder"
		for _, seed := range seeds {
			command += fmt.Sprintf(" %s", seed.Signature())
		}
	}

	artisanFacades.Call(command)
}

func (receiver *TestCase) RefreshDatabase(seeds ...seeder.Seeder) {
	artisanFacades.Call("migrate:refresh")
}
