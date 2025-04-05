package testing

import (
	"fmt"
	"testing"

	contractsseeder "github.com/goravel/framework/contracts/database/seeder"
	contractshttp "github.com/goravel/framework/contracts/testing/http"
	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/testing/http"
)

type TestCase struct {
}

func (r *TestCase) Http(t *testing.T) contractshttp.Request {
	return http.NewTestRequest(t, json, routeFacade, sessionFacade)
}

func (r *TestCase) Seed(seeders ...contractsseeder.Seeder) {
	if artisanFacade == nil {
		panic(errors.ArtisanFacadeNotSet.SetModule(errors.ModuleTesting))
	}

	if err := artisanFacade.Call("--no-ansi db:seed" + getCommandOptionOfSeeders(seeders)); err != nil {
		panic(err)
	}
}

func (r *TestCase) RefreshDatabase(seeders ...contractsseeder.Seeder) {
	if artisanFacade == nil {
		panic(errors.ArtisanFacadeNotSet.SetModule(errors.ModuleTesting))
	}

	if err := artisanFacade.Call("--no-ansi migrate:refresh" + getCommandOptionOfSeeders(seeders)); err != nil {
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
