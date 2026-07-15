package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStubs_Config(t *testing.T) {
	content := Stubs{}.Config("config", "github.com/goravel/framework/facades", "facades")

	assert.Contains(t, content, "package config")
	assert.Contains(t, content, `"github.com/goravel/framework/facades"`)
	assert.Contains(t, content, "facades.Config()")
	assert.Contains(t, content, `config.Add("notification"`)
	assert.Contains(t, content, `config.Env("NOTIFICATION_CHANNEL", "mail")`)
	assert.Contains(t, content, `config.Env("NOTIFICATION_DB_CONNECTION", "")`)
	assert.NotContains(t, content, "DummyPackage")
	assert.NotContains(t, content, "DummyFacadesImport")
	assert.NotContains(t, content, "DummyFacadesPackage")
}

func TestStubs_NotificationFacade(t *testing.T) {
	content := Stubs{}.NotificationFacade("facades")

	assert.Contains(t, content, "package facades")
	assert.Contains(t, content, `"github.com/goravel/framework/contracts/notification"`)
	assert.Contains(t, content, "func Notification() notification.Manager")
	assert.Contains(t, content, "return App().MakeNotification()")
	assert.NotContains(t, content, "DummyPackage")
}
