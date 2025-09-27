package main

import (
	"fmt"
	"os"

	"github.com/goravel/framework/packages"
	"github.com/goravel/framework/packages/match"
	"github.com/goravel/framework/packages/modify"
	"github.com/goravel/framework/support/file"
	"github.com/goravel/framework/support/path"
	"github.com/goravel/framework/support/stub"
)

func main() {
	stubs := Stubs{}
	appServiceProviderPath := path.App("providers", "app_service_provider.go")
	kernelPath := path.App("console", "kernel.go")
	moduleName := packages.GetModuleNameFromArgs(os.Args)
	scheduleServiceProvider := "&schedule.ServiceProvider{}"
	registerSchedule := "facades.Schedule().Register(console.Kernel{}.Schedule())"
	facadesImport := fmt.Sprintf("%s/app/facades", moduleName)
	consoleImport := fmt.Sprintf("%s/app/console", moduleName)

	packages.Setup(os.Args).
		Install(
			modify.GoFile(path.Config("app.go")).
				Find(match.Imports()).Modify(modify.AddImport(packages.GetModulePath())).
				Find(match.Providers()).Modify(modify.Register(scheduleServiceProvider)),
			modify.File(kernelPath).Overwrite(stub.ConsoleKernel()),
			modify.GoFile(appServiceProviderPath).
				Find(match.Imports()).Modify(modify.AddImport(facadesImport)).
				Find(match.Imports()).Modify(modify.AddImport(consoleImport)).
				Find(match.Register()).Modify(modify.Add(registerSchedule)),
			modify.WhenFacade("Schedule", modify.File(path.Facades("schedule.go")).Overwrite(stubs.ScheduleFacade())),
		).
		Uninstall(
			modify.WhenNoFacades([]string{"Schedule"},
				modify.GoFile(path.Config("app.go")).
					Find(match.Providers()).Modify(modify.Unregister(scheduleServiceProvider)).
					Find(match.Imports()).Modify(modify.RemoveImport(packages.GetModulePath())),
				modify.GoFile(appServiceProviderPath).
					Find(match.Register()).Modify(modify.Remove(registerSchedule)).
					Find(match.Imports()).Modify(modify.RemoveImport(facadesImport)).
					Find(match.Imports()).Modify(modify.RemoveImport(consoleImport)),
			),
			modify.When(isKernelNotModified, modify.File(kernelPath).Remove()),
			modify.WhenFacade("Schedule", modify.File(path.Facades("schedule.go")).Remove()),
		).
		Execute()
}

func isKernelNotModified() bool {
	content, err := file.GetContent(path.App("console", "kernel.go"))
	if err != nil {
		return false
	}

	return content == stub.ConsoleKernel()
}
