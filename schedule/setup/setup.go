package main

import (
	"fmt"
	"os"

	"github.com/goravel/framework/packages"
	"github.com/goravel/framework/packages/match"
	"github.com/goravel/framework/packages/modify"
	"github.com/goravel/framework/support/file"
	"github.com/goravel/framework/support/path"
	"github.com/goravel/framework/support/stubs"
)

func main() {
	scheduleFacade := "Schedule"
	appServiceProviderPath := path.App("providers", "app_service_provider.go")
	kernelPath := path.App("console", "kernel.go")
	scheduleFacadePath := path.Facades("schedule.go")
	moduleName := packages.GetModuleNameFromArgs(os.Args)
	scheduleServiceProvider := "&schedule.ServiceProvider{}"
	registerSchedule := "facades.Schedule().Register(console.Kernel{}.Schedule())"
	facadesImport := fmt.Sprintf("%s/app/facades", moduleName)
	consoleImport := fmt.Sprintf("%s/app/console", moduleName)
	modulePath := packages.GetModulePath()

	packages.Setup(os.Args).
		Install(
			// Create the console kernel file if it does not exist.
			modify.WhenFileNotExists(kernelPath, modify.File(kernelPath).Overwrite(stubs.ConsoleKernel())),

			// Create the schedule facade file.
			modify.WhenFacade(scheduleFacade, modify.File(scheduleFacadePath).Overwrite(Stubs{}.ScheduleFacade())),

			// Add the schedule service provider to the providers array in bootstrap/providers.go
			modify.AddProviderApply(modulePath, scheduleServiceProvider),

			// Add the schedule registration to the AppServiceProvider.
			modify.GoFile(appServiceProviderPath).
				Find(match.Imports()).Modify(modify.AddImport(facadesImport)).
				Find(match.Imports()).Modify(modify.AddImport(consoleImport)).
				Find(match.RegisterFunc()).Modify(modify.Add(registerSchedule)),
		).
		Uninstall(
			modify.WhenNoFacades([]string{scheduleFacade},
				// Remove the schedule registration from the AppServiceProvider.
				modify.GoFile(appServiceProviderPath).
					Find(match.RegisterFunc()).Modify(modify.Remove(registerSchedule)).
					Find(match.Imports()).Modify(modify.RemoveImport(facadesImport)).
					Find(match.Imports()).Modify(modify.RemoveImport(consoleImport)),

				// Remove the schedule service provider from the providers array in bootstrap/providers.go
				modify.RemoveProviderApply(modulePath, scheduleServiceProvider),

				// Remove the console kernel file if it was not modified.
				modify.When(isKernelNotModified, modify.File(kernelPath).Remove()),
			),

			// Remove the schedule facade file.
			modify.WhenFacade(scheduleFacade, modify.File(scheduleFacadePath).Remove()),
		).
		Execute()
}

func isKernelNotModified(_ map[string]any) bool {
	content, err := file.GetContent(path.App("console", "kernel.go"))
	if err != nil {
		return false
	}

	return content == stubs.ConsoleKernel()
}
