package main

import (
	"os"

	"github.com/goravel/framework/packages"
	"github.com/goravel/framework/packages/modify"
	"github.com/goravel/framework/support/path"
)

func main() {
	scheduleFacade := "Schedule"
	scheduleFacadePath := path.Facades("schedule.go")
	scheduleServiceProvider := "&schedule.ServiceProvider{}"
	modulePath := packages.GetModulePath()

	packages.Setup(os.Args).
		Install(
			// Create the schedule facade file.
			modify.WhenFacade(scheduleFacade, modify.File(scheduleFacadePath).Overwrite(Stubs{}.ScheduleFacade())),

			// Add the schedule service provider to the providers array in bootstrap/providers.go
			modify.AddProviderApply(modulePath, scheduleServiceProvider),
		).
		Uninstall(
			modify.WhenNoFacades([]string{scheduleFacade},
				// Remove the schedule service provider from the providers array in bootstrap/providers.go
				modify.RemoveProviderApply(modulePath, scheduleServiceProvider),
			),

			// Remove the schedule facade file.
			modify.WhenFacade(scheduleFacade, modify.File(scheduleFacadePath).Remove()),
		).
		Execute()
}
