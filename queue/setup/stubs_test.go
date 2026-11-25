package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStubs_Config(t *testing.T) {
	content := Stubs{}.Config("config", "goravel/app/facades", "facades")

	assert.Contains(t, content, `"default": config.Env("QUEUE_CONNECTION", "sync"),`)
	assert.NotContains(t, content, `"default": "sync",`)
	assert.NotContains(t, content, "DummyPackage")
	assert.NotContains(t, content, "DummyFacadesImport")
	assert.NotContains(t, content, "DummyFacadesPackage")
}
