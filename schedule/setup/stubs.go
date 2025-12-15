package main

import "strings"

type Stubs struct{}

func (s Stubs) ScheduleFacade(pkg string) string {
	content := `package DummyPackage

import (
	"github.com/goravel/framework/contracts/schedule"
)

func Schedule() schedule.Schedule {
	return App().MakeSchedule()
}
`

	return strings.ReplaceAll(content, "DummyPackage", pkg)
}
