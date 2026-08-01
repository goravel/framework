package broadcasting

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/goravel/framework/contracts/broadcasting"
	mocksfoundation "github.com/goravel/framework/mocks/foundation"
	mockshttp "github.com/goravel/framework/mocks/http/client"
	mockslog "github.com/goravel/framework/mocks/log"
)

func TestCreateDriver_Pusher(t *testing.T) {
	conn := broadcasting.ConnectionConfig{
		Driver:  "pusher",
		Key:     "test-key",
		Secret:  "test-secret",
		AppID:   "test-app",
		Options: map[string]any{"cluster": "mt1", "port": float64(443), "scheme": "https", "host": "api-mt1.pusher.com"},
	}

	mockHttp := mockshttp.NewFactory(t)

	app := mocksfoundation.NewApplication(t)
	app.EXPECT().MakeHttp().Return(mockHttp).Once()

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
