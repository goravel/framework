package database

import (
	"log"
	"testing"

	supportdocker "github.com/goravel/framework/support/docker"
)

var testDatabaseDocker *supportdocker.Database

func TestMain(m *testing.M) {
	var err error
	testDatabaseDocker, err = supportdocker.InitDatabase()
	if err != nil {
		log.Fatalf("Init docker error: %s", err)
	}

	m.Run()

	defer func() {
		if err := testDatabaseDocker.Stop(); err != nil {
			log.Fatalf("Stop docker error: %s", err)
		}
	}()
}
