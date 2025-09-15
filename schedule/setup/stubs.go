package main

type Stubs struct{}

func (s Stubs) ScheduleFacade() string {
	return `package facades

import (
	"github.com/goravel/framework/contracts/schedule"
)

func Schedule() schedule.Schedule {
	return App().MakeSchedule()
}
`
}
