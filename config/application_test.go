package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/support"
	"github.com/goravel/framework/support/file"
)

type ApplicationTestSuite struct {
	suite.Suite
	config       *Application
	customConfig *Application
}

func TestApplicationTestSuite(t *testing.T) {
	assert.Nil(t, file.PutContent(support.EnvFilePath, `
APP_KEY=12345678901234567890123456789012
APP_DEBUG=true
DB_PORT=3306
TIMEOUT=10s
FLOAT_VALUE=3.14
`))
	temp, err := os.CreateTemp("", "goravel.env")
	assert.NoError(t, err)
	defer func() {
		_ = temp.Close()
		_ = os.Remove(temp.Name())
	}()

	_, err = temp.Write([]byte(`
APP_KEY=12345678901234567890123456789012
APP_DEBUG=true
DB_PORT=3306
TIMEOUT=20s
FLOAT_VALUE=6.28
`))
	assert.NoError(t, err)

	suite.Run(t, &ApplicationTestSuite{
		config:       NewApplication(support.EnvFilePath),
		customConfig: NewApplication(temp.Name()),
	})

	assert.NoError(t, file.Remove(support.EnvFilePath))
}

func (s *ApplicationTestSuite) SetupTest() {

}

func (s *ApplicationTestSuite) TestOsVariables() {
	s.Nil(os.Setenv("APP_KEY", "12345678901234567890123456789013"))
	s.Nil(os.Setenv("OS_APP_NAME", "goravel"))
	s.Nil(os.Setenv("OS_APP_PORT", "3306"))
	s.Nil(os.Setenv("OS_APP_DEBUG", "true"))
	s.Nil(os.Setenv("OS_TIMEOUT", "5s"))

	s.Equal("12345678901234567890123456789013", s.config.GetString("APP_KEY"))
	s.Equal("12345678901234567890123456789013", s.customConfig.GetString("APP_KEY"))
	s.Equal("goravel", s.config.GetString("OS_APP_NAME"))
	s.Equal("goravel", s.customConfig.GetString("OS_APP_NAME"))
	s.Equal(3306, s.config.GetInt("OS_APP_PORT"))
	s.Equal(3306, s.customConfig.GetInt("OS_APP_PORT"))
	s.True(s.config.GetBool("OS_APP_DEBUG"))
	s.True(s.customConfig.GetBool("OS_APP_DEBUG"))
	s.Equal(5*time.Second, s.config.GetDuration("OS_TIMEOUT"))
	s.Equal(5*time.Second, s.customConfig.GetDuration("OS_TIMEOUT"))
}

func (s *ApplicationTestSuite) TestEnv() {
	s.Equal("12345678901234567890123456789012", s.config.Env("APP_KEY").(string))
	s.Equal("goravel", s.config.Env("APP_NAME", "goravel").(string))
	s.Equal("12345678901234567890123456789012", s.customConfig.Env("APP_KEY").(string))
	s.Equal("goravel", s.customConfig.Env("APP_NAME", "goravel").(string))
}

func (s *ApplicationTestSuite) TestAdd() {
	s.config.Add("app", map[string]any{
		"env": "local",
	})
	s.customConfig.Add("app", map[string]any{
		"env": "local",
	})

	s.Equal("local", s.config.GetString("app.env"))
	s.Equal("local", s.customConfig.GetString("app.env"))

	s.config.Add("path.with.dot.case1", "value1")
	s.customConfig.Add("path.with.dot.case1", "value1")
	s.Equal("value1", s.config.GetString("path.with.dot.case1"))
	s.Equal("value1", s.customConfig.GetString("path.with.dot.case1"))

	s.config.Add("path.with.dot.case2", "value2")
	s.customConfig.Add("path.with.dot.case2", "value2")
	s.Equal("value2", s.config.GetString("path.with.dot.case2"))
	s.Equal("value2", s.customConfig.GetString("path.with.dot.case2"))

	s.config.Add("path.with.dot", map[string]any{"case3": "value3"})
	s.customConfig.Add("path.with.dot", map[string]any{"case3": "value3"})
	s.Equal("value3", s.config.GetString("path.with.dot.case3"))
	s.Equal("value3", s.customConfig.GetString("path.with.dot.case3"))

	s.config.Add("key.with.timestamp", 5*time.Second)
	s.customConfig.Add("key.with.timestamp", "20s")
	s.Equal(5*time.Second, s.config.GetDuration("key.with.timestamp"))
	s.Equal(20*time.Second, s.customConfig.GetDuration("key.with.timestamp"))
}

func (s *ApplicationTestSuite) TestGet() {
	s.Equal("12345678901234567890123456789012", s.config.Get("APP_KEY").(string))
	s.Equal("goravel", s.config.Get("APP_NAME", "goravel").(string))
	s.Equal("12345678901234567890123456789012", s.customConfig.Get("APP_KEY").(string))
	s.Equal("goravel", s.customConfig.Get("APP_NAME", "goravel").(string))
}

func (s *ApplicationTestSuite) TestGetString() {
	s.config.Add("database", map[string]any{
		"default": s.config.Env("DB_CONNECTION", "mysql"),
		"migrations": map[string]any{
			"table": "migrations",
		},
	})
	s.customConfig.Add("database", map[string]any{
		"default": s.customConfig.Env("DB_CONNECTION", "mysql"),
		"migrations": map[string]any{
			"table": "migrations",
		},
	})

	s.Equal("goravel", s.config.GetString("APP_NAME", "goravel"))
	s.Equal("migrations", s.config.GetString("database.migrations.table"))
	s.Equal("mysql", s.config.GetString("database.default"))
	s.Equal("goravel", s.customConfig.GetString("APP_NAME", "goravel"))
	s.Equal("migrations", s.customConfig.GetString("database.migrations.table"))
	s.Equal("mysql", s.customConfig.GetString("database.default"))
}

func (s *ApplicationTestSuite) TestGetInt() {
	s.Equal(3306, s.config.GetInt("DB_PORT"))
	s.Equal(3306, s.customConfig.GetInt("DB_PORT"))
	s.Equal(0, s.config.GetInt("NOT_EXIST"))
	s.Equal(123, s.config.GetInt("NOT_EXIST", 123))
	s.Equal(3, s.config.GetInt("FLOAT_VALUE"))
}

func (s *ApplicationTestSuite) TestGetBool() {
	s.True(s.config.GetBool("APP_DEBUG"))
	s.True(s.customConfig.GetBool("APP_DEBUG"))
	s.False(s.config.GetBool("NON_EXISTENT_BOOL"))
	s.True(s.config.GetBool("NON_EXISTENT_BOOL", true))
	s.False(s.config.GetBool("DB_PORT"))

	s.config.Add("MY_BOOL_TRUE", "true")
	s.config.Add("MY_BOOL_FALSE", "false")
	s.True(s.config.GetBool("MY_BOOL_TRUE"))
	s.False(s.config.GetBool("MY_BOOL_FALSE"))

	s.config.Add("MY_BOOL_INVALID", "invalid")
	s.False(s.config.GetBool("MY_BOOL_INVALID"))
}

func (s *ApplicationTestSuite) TestGetDuration() {
	s.Equal(10*time.Second, s.config.GetDuration("TIMEOUT"))
	s.Equal(20*time.Second, s.customConfig.GetDuration("TIMEOUT"))

	s.Equal(time.Duration(0), s.config.GetDuration("NON_EXISTENT_DURATION"))
	s.Equal(time.Second, s.config.GetDuration("NON_EXISTENT_DURATION", time.Second))

	s.config.Add("INVALID_DURATION", "invalid")
	s.customConfig.Add("INVALID_DURATION", "invalid")
	s.Equal(time.Duration(0), s.config.GetDuration("INVALID_DURATION"))
	s.Equal(time.Duration(0), s.config.GetDuration("INVALID_DURATION", time.Second))
}

func TestOsVariables(t *testing.T) {
	assert.Nil(t, os.Setenv("APP_KEY", "12345678901234567890123456789013"))
	assert.Nil(t, os.Setenv("APP_NAME", "goravel"))
	assert.Nil(t, os.Setenv("APP_PORT", "3306"))
	assert.Nil(t, os.Setenv("APP_DEBUG", "true"))

	config := NewApplication(support.EnvFilePath)

	assert.Equal(t, "12345678901234567890123456789013", config.GetString("APP_KEY"))
	assert.Equal(t, "goravel", config.GetString("APP_NAME"))
	assert.Equal(t, 3306, config.GetInt("APP_PORT"))
	assert.True(t, config.GetBool("APP_DEBUG"))
}
