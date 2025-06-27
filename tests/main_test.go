package tests

import (
	"os"
	"testing"

	"github.com/goravel/framework/foundation/json"
	mocksfoundation "github.com/goravel/framework/mocks/foundation"
	"github.com/goravel/mysql"
	"github.com/goravel/postgres"
	"github.com/goravel/sqlite"
	"github.com/goravel/sqlserver"
)

func TestMain(m *testing.M) {
	mockApp := &mocksfoundation.Application{}
	mockApp.EXPECT().GetJson().Return(json.New())
	postgres.App = mockApp
	mysql.App = mockApp
	sqlite.App = mockApp
	sqlserver.App = mockApp

	os.Exit(m.Run())
}
