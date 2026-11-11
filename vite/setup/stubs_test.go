package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStubs_Config(t *testing.T) {
	content := Stubs{}.Config("config", "goravel/app/facades", "facades")

	assert.Contains(t, content, "package config")
	assert.Contains(t, content, `"goravel/app/facades"`)
	assert.Contains(t, content, "config := facades.Config()")
	assert.Contains(t, content, `config.Add("vite", map[string]any{`)
	assert.Contains(t, content, `"build_directory": "build",`)
	assert.Contains(t, content, `"hot_file": "hot",`)
	assert.Contains(t, content, `"manifest": ".vite/manifest.json",`)
	assert.Contains(t, content, `"inputs": []string{},`)
	assert.Contains(t, content, `"assets": []string{},`)
	assert.NotContains(t, content, `"dev_url":`)
	assert.NotContains(t, content, "DummyPackage")
	assert.NotContains(t, content, "DummyFacadesImport")
	assert.NotContains(t, content, "DummyFacadesPackage")
}

func TestStubs_ViteFacade(t *testing.T) {
	content := Stubs{}.ViteFacade("facades")

	assert.Contains(t, content, "package facades")
	assert.Contains(t, content, `"github.com/goravel/framework/contracts/vite"`)
	assert.Contains(t, content, "func Vite() vite.Vite")
	assert.Contains(t, content, "return App().MakeVite()")
	assert.NotContains(t, content, "DummyPackage")
}
