package testing

import (
	"fmt"
	"testing"

	contractsseeder "github.com/goravel/framework/contracts/database/seeder"
	contractstesting "github.com/goravel/framework/contracts/testing"
	"github.com/goravel/framework/errors"
)

type TestCase struct {
}

func (r *TestCase) Http(t *testing.T) contractstesting.TestRequest {
	return NewTestRequest(t)
}

func (r *TestCase) Seed(seeders ...contractsseeder.Seeder) {
	if artisanFacade == nil {
		panic(errors.ArtisanFacadeNotSet.SetModule(errors.ModuleTesting))
	}

	if err := artisanFacade.Call("db:seed" + getCommandOptionOfSeeders(seeders)); err != nil {
		panic(err)
	}
}

func (r *TestCase) RefreshDatabase(seeders ...contractsseeder.Seeder) {
	if artisanFacade == nil {
		panic(errors.ArtisanFacadeNotSet.SetModule(errors.ModuleTesting))
	}

	if err := artisanFacade.Call("migrate:refresh" + getCommandOptionOfSeeders(seeders)); err != nil {
		panic(err)
	}
}

func getCommandOptionOfSeeders(seeders []contractsseeder.Seeder) string {
	if len(seeders) == 0 {
		return ""
	}

	command := " --seeder"
	for _, seed := range seeders {
		command += fmt.Sprintf(" %s", seed.Signature())
	}

	return command
}
