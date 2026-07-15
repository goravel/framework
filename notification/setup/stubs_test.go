package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStubs_NotificationFacade(t *testing.T) {
	content := Stubs{}.NotificationFacade("facades")

	assert.Contains(t, content, "package facades")
	assert.Contains(t, content, `"github.com/goravel/framework/contracts/notification"`)
	assert.Contains(t, content, "func Notification() notification.Manager")
	assert.Contains(t, content, "return App().MakeNotification()")
	assert.NotContains(t, content, "DummyPackage")
}
