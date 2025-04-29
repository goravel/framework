package log

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/goravel/framework/errors"
	"github.com/goravel/framework/foundation/json"
	mocksconfig "github.com/goravel/framework/mocks/config"
	"github.com/goravel/framework/support/color"
)

func TestNewApplication(t *testing.T) {
	j := json.NewJson()
	app, err := NewApplication(nil, j)
	assert.Nil(t, err)
	assert.NotNil(t, app)

	mockConfig := mocksconfig.NewConfig(t)
	mockConfig.EXPECT().GetString("logging.default").Return("test").Once()
	mockConfig.EXPECT().GetString("logging.channels.test.driver").Return("single").Once()
	mockConfig.EXPECT().GetString("logging.channels.test.path").Return("test").Once()
	mockConfig.EXPECT().GetString("logging.channels.test.level").Return("debug").Once()
	mockConfig.EXPECT().GetBool("logging.channels.test.print").Return(true).Once()
	app, err = NewApplication(mockConfig, j)
	assert.Nil(t, err)
	assert.NotNil(t, app)

	mockConfig = &mocksconfig.Config{}
	mockConfig.EXPECT().GetString("logging.default").Return("test").Once()
	mockConfig.EXPECT().GetString("logging.channels.test.driver").Return("test").Once()

	app, err = NewApplication(mockConfig, j)
	assert.EqualError(t, err, errors.LogDriverNotSupported.Args("test").Error())
	assert.Nil(t, app)
}

func TestApplication_Channel(t *testing.T) {
	mockConfig := mocksconfig.NewConfig(t)
	mockConfig.EXPECT().GetString("logging.default").Return("test").Once()
	mockConfig.EXPECT().GetString("logging.channels.test.driver").Return("single").Once()
	mockConfig.EXPECT().GetString("logging.channels.test.path").Return("test").Once()
	mockConfig.EXPECT().GetString("logging.channels.test.level").Return("debug").Once()
	mockConfig.EXPECT().GetBool("logging.channels.test.print").Return(true).Once()
	app, err := NewApplication(mockConfig, json.NewJson())
	assert.Nil(t, err)
	assert.NotNil(t, app)
	assert.NotNil(t, app.Channel(""))

	mockConfig.EXPECT().GetString("logging.channels.dummy.driver").Return("daily").Once()
	mockConfig.EXPECT().GetString("logging.channels.dummy.path").Return("dummy").Once()
	mockConfig.EXPECT().GetString("logging.channels.dummy.level").Return("debug").Once()
	mockConfig.EXPECT().GetBool("logging.channels.dummy.print").Return(true).Once()
	mockConfig.EXPECT().GetInt("logging.channels.dummy.days").Return(1).Once()
	writer := app.Channel("dummy")
	assert.NotNil(t, writer)

	mockConfig.EXPECT().GetString("logging.channels.test2.driver").Return("test2").Once()
	assert.Contains(t, color.CaptureOutput(func(w io.Writer) {
		assert.Nil(t, app.Channel("test2"))
	}), errors.LogDriverNotSupported.Args("test2").Error())
}

func TestApplication_Stack(t *testing.T) {
	mockConfig := mocksconfig.NewConfig(t)
	mockConfig.EXPECT().GetString("logging.default").Return("test").Once()
	mockConfig.EXPECT().GetString("logging.channels.test.driver").Return("single").Once()
	mockConfig.EXPECT().GetString("logging.channels.test.path").Return("test").Once()
	mockConfig.EXPECT().GetString("logging.channels.test.level").Return("debug").Once()
	mockConfig.EXPECT().GetBool("logging.channels.test.print").Return(true).Once()
	app, err := NewApplication(mockConfig, json.NewJson())
	assert.Nil(t, err)
	assert.NotNil(t, app)
	assert.NotNil(t, app.Stack([]string{}))

	mockConfig.EXPECT().GetString("logging.channels.test2.driver").Return("test2").Once()
	assert.Contains(t, color.CaptureOutput(func(w io.Writer) {
		assert.Nil(t, app.Stack([]string{"", "test2", "daily"}))
	}), errors.LogDriverNotSupported.Args("test2").Error())

	mockConfig.EXPECT().GetString("logging.channels.dummy.driver").Return("daily").Once()
	mockConfig.EXPECT().GetString("logging.channels.dummy.path").Return("dummy").Once()
	mockConfig.EXPECT().GetString("logging.channels.dummy.level").Return("debug").Once()
	mockConfig.EXPECT().GetBool("logging.channels.dummy.print").Return(true).Once()
	mockConfig.EXPECT().GetInt("logging.channels.dummy.days").Return(1).Once()
	assert.NotNil(t, app.Stack([]string{"dummy"}))
}
