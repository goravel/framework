package docker

import (
	"fmt"

	"github.com/glebarez/sqlite"
	gormio "gorm.io/gorm"

	"github.com/goravel/framework/contracts/database/orm"
	"github.com/goravel/framework/contracts/testing"
	"github.com/goravel/framework/support/file"
)

type Sqlite struct {
	database string
	image    *testing.Image
}

func NewSqlite(database string) *Sqlite {
	return &Sqlite{
		database: database,
	}
}

func (receiver *Sqlite) Build() error {
	if _, err := receiver.connect(); err != nil {
		return fmt.Errorf("connect Sqlite error: %v", err)
	}

	return nil
}

func (receiver *Sqlite) Config() testing.DatabaseConfig {
	return testing.DatabaseConfig{
		Database: receiver.database,
	}
}

func (receiver *Sqlite) Fresh() error {
	if err := receiver.Stop(); err != nil {
		return err
	}

	if _, err := receiver.connect(); err != nil {
		return fmt.Errorf("connect Sqlite error when freshing: %v", err)
	}

	return nil
}

func (receiver *Sqlite) Image(image testing.Image) {
	receiver.image = &image
}

func (receiver *Sqlite) Name() orm.Driver {
	return orm.DriverSqlite
}

func (receiver *Sqlite) Stop() error {
	if err := file.Remove(receiver.database); err != nil {
		return fmt.Errorf("stop Sqlite error: %v", err)
	}

	return nil
}

func (receiver *Sqlite) connect() (*gormio.DB, error) {
	return gormio.Open(sqlite.Open(fmt.Sprintf("%s?multi_stmts=true", receiver.database)))
}
