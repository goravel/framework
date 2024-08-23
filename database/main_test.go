package database

import (
	"log"
	"os"
	"testing"

	supportdocker "github.com/goravel/framework/support/docker"
	"github.com/goravel/framework/support/env"
)

func TestMain(m *testing.M) {
	exit := m.Run()

	if !env.IsWindows() {
		if err := supportdocker.Stop(); err != nil {
			log.Fatalf("Stop docker error: %s", err)
		}
	}

	os.Exit(exit)
}
