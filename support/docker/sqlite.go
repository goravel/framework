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
}

func NewSqliteImpl(database string) *SqliteImpl {
	return &SqliteImpl{
		database: database,
	}
}

func (r *SqliteImpl) Build() error {
	if _, err := r.connect(); err != nil {
		return fmt.Errorf("connect Sqlite error: %v", err)
	}

	return nil
}

func (r *SqliteImpl) Config() testing.DatabaseConfig {
	return testing.DatabaseConfig{
		Database: r.database,
	}
}

func (r *SqliteImpl) Database(name string) (testing.DatabaseDriver, error) {
	sqliteImpl := NewSqliteImpl(name)
	if err := sqliteImpl.Build(); err != nil {
		return nil, err
	}

	return sqliteImpl, nil
}

func (r *SqliteImpl) Driver() database.Driver {
	return database.DriverSqlite
}

func (r *SqliteImpl) Fresh() error {
	if err := r.Stop(); err != nil {
		return err
	}

	if _, err := r.connect(); err != nil {
		return fmt.Errorf("connect Sqlite error when freshing: %v", err)
	}

	return nil
}

func (r *SqliteImpl) Image(image testing.Image) {
}

func (r *SqliteImpl) Ready() error {
	_, err := r.connect()

	return err
}

func (r *SqliteImpl) Stop() error {
	if err := file.Remove(r.database); err != nil {
		return fmt.Errorf("stop Sqlite error: %v", err)
	}

	return nil
}

func (r *SqliteImpl) connect() (*gormio.DB, error) {
	return gormio.Open(sqlite.Open(fmt.Sprintf("%s?multi_stmts=true", r.database)))
}
