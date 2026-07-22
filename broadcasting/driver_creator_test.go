package broadcasting

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/goravel/framework/contracts/broadcasting"
	mocksfoundation "github.com/goravel/framework/mocks/foundation"
	mockslog "github.com/goravel/framework/mocks/log"
)

func TestCreateDriver_Pusher(t *testing.T) {
	conn := broadcasting.ConnectionConfig{
		Driver: "pusher",
		Key:    "test-key",
		Secret: "test-secret",
		AppID:  "test-app",
		Options: broadcasting.PusherOptions{
			Cluster: "mt1",
		},
	}
	app := mocksfoundation.NewApplication(t)

	driver, err := CreateDriver(conn, app)
	assert.NoError(t, err)
	assert.NotNil(t, driver)
}

func TestCreateDriver_Log(t *testing.T) {
	conn := broadcasting.ConnectionConfig{
		Driver: "log",
	}
	mockLog := mockslog.NewLog(t)

	app := mocksfoundation.NewApplication(t)
	app.EXPECT().MakeLog().Return(mockLog).Once()

	driver, err := CreateDriver(conn, app)
	assert.NoError(t, err)
	assert.NotNil(t, driver)
}

func TestCreateDriver_Null(t *testing.T) {
	conn := broadcasting.ConnectionConfig{
		Driver: "null",
	}
	app := mocksfoundation.NewApplication(t)

	driver, err := CreateDriver(conn, app)
	assert.NoError(t, err)
	assert.NotNil(t, driver)
}

func TestCreateDriver_Unknown(t *testing.T) {
	conn := broadcasting.ConnectionConfig{
		Driver: "unknown",
	}
	app := mocksfoundation.NewApplication(t)

	driver, err := CreateDriver(conn, app)
	assert.Error(t, err)
	assert.Nil(t, driver)
}
