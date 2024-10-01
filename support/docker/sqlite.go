package docker

import (
	"fmt"

	"github.com/glebarez/sqlite"
	gormio "gorm.io/gorm"

	"github.com/goravel/framework/contracts/database"
	"github.com/goravel/framework/contracts/testing"
	"github.com/goravel/framework/support/file"
)

type SqliteImpl struct {
	database string
	image    *testing.Image
}

func NewSqliteImpl(database string) *SqliteImpl {
	return &SqliteImpl{
		database: database,
	}
}

func (receiver *SqliteImpl) Build() error {
	if _, err := receiver.connect(); err != nil {
		return fmt.Errorf("connect Sqlite error: %v", err)
	}

	return nil
}

func (receiver *SqliteImpl) Config() testing.DatabaseConfig {
	return testing.DatabaseConfig{
		Database: receiver.database,
	}
}

func (receiver *SqliteImpl) Driver() database.Driver {
	return database.DriverSqlite
}

func (receiver *SqliteImpl) Fresh() error {
	if err := receiver.Stop(); err != nil {
		return err
	}

	if _, err := receiver.connect(); err != nil {
		return fmt.Errorf("connect Sqlite error when freshing: %v", err)
	}

	return nil
}

func (receiver *SqliteImpl) Image(image testing.Image) {
	receiver.image = &image
}

func (receiver *SqliteImpl) Stop() error {
	if err := file.Remove(receiver.database); err != nil {
		return fmt.Errorf("stop Sqlite error: %v", err)
	}

	return nil
}

func (receiver *SqliteImpl) connect() (*gormio.DB, error) {
	return gormio.Open(sqlite.Open(fmt.Sprintf("%s?multi_stmts=true", receiver.database)))
}
