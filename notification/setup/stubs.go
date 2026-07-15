package main

import "strings"

type Stubs struct{}

func (s Stubs) NotificationFacade(pkg string) string {
	content := `package DummyPackage

import (
	"github.com/goravel/framework/contracts/notification"
)

func Notification() notification.Manager {
	return App().MakeNotification()
}
`

	return strings.ReplaceAll(content, "DummyPackage", pkg)
}
