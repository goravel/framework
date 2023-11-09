package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/support/file"
)

type ApplicationTestSuite struct {
	suite.Suite
	config       *Application
	customConfig *Application
}

func TestApplicationTestSuite(t *testing.T) {
	assert.Nil(t, file.Create(".env", `
APP_KEY=12345678901234567890123456789012
APP_DEBUG=true
DB_PORT=3306
`))
	temp, err := os.CreateTemp("", "goravel.env")
	assert.Nil(t, err)
	defer os.Remove(temp.Name())
	defer temp.Close()

	_, err = temp.Write([]byte(`
APP_KEY=12345678901234567890123456789012
APP_DEBUG=true
DB_PORT=3306
`))
	assert.Nil(t, err)
	assert.Nil(t, temp.Close())

	suite.Run(t, &ApplicationTestSuite{
		config:       NewApplication(".env"),
		customConfig: NewApplication(temp.Name()),
	})

	assert.Nil(t, file.Remove(".env"))
}

func (s *ApplicationTestSuite) SetupTest() {

}

func (s *ApplicationTestSuite) TestOsVariables() {
	s.Nil(os.Setenv("APP_KEY", "12345678901234567890123456789013"))
	s.Nil(os.Setenv("OS_APP_NAME", "goravel"))
	s.Nil(os.Setenv("OS_APP_PORT", "3306"))
	s.Nil(os.Setenv("OS_APP_DEBUG", "true"))

	s.Equal("12345678901234567890123456789013", s.config.GetString("APP_KEY"))
	s.Equal("12345678901234567890123456789013", s.customConfig.GetString("APP_KEY"))
	s.Equal("goravel", s.config.GetString("OS_APP_NAME"))
	s.Equal("goravel", s.customConfig.GetString("OS_APP_NAME"))
	s.Equal(3306, s.config.GetInt("OS_APP_PORT"))
	s.Equal(3306, s.customConfig.GetInt("OS_APP_PORT"))
	s.True(s.config.GetBool("OS_APP_DEBUG"))
	s.True(s.customConfig.GetBool("OS_APP_DEBUG"))
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
		"connections": map[string]any{
			"mysql": map[string]any{
				"host": s.config.Env("DB_HOST", "127.0.0.1"),
			},
		},
	})
	s.customConfig.Add("database", map[string]any{
		"default": s.customConfig.Env("DB_CONNECTION", "mysql"),
		"connections": map[string]any{
			"mysql": map[string]any{
				"host": s.customConfig.Env("DB_HOST", "127.0.0.1"),
			},
		},
	})

	s.Equal("goravel", s.config.GetString("APP_NAME", "goravel"))
	s.Equal("127.0.0.1", s.config.GetString("database.connections.mysql.host"))
	s.Equal("mysql", s.config.GetString("database.default"))
	s.Equal("goravel", s.customConfig.GetString("APP_NAME", "goravel"))
	s.Equal("127.0.0.1", s.customConfig.GetString("database.connections.mysql.host"))
	s.Equal("mysql", s.customConfig.GetString("database.default"))
}

func (s *ApplicationTestSuite) TestGetInt() {
	s.Equal(3306, s.config.GetInt("DB_PORT"))
	s.Equal(3306, s.customConfig.GetInt("DB_PORT"))
}

func (s *ApplicationTestSuite) TestGetBool() {
	s.Equal(true, s.config.GetBool("APP_DEBUG"))
	s.Equal(true, s.customConfig.GetBool("APP_DEBUG"))
}

func TestOsVariables(t *testing.T) {
	assert.Nil(t, os.Setenv("APP_KEY", "12345678901234567890123456789013"))
	assert.Nil(t, os.Setenv("APP_NAME", "goravel"))
	assert.Nil(t, os.Setenv("APP_PORT", "3306"))
	assert.Nil(t, os.Setenv("APP_DEBUG", "true"))

	config := NewApplication(".env")

	assert.Equal(t, "12345678901234567890123456789013", config.GetString("APP_KEY"))
	assert.Equal(t, "goravel", config.GetString("APP_NAME"))
	assert.Equal(t, 3306, config.GetInt("APP_PORT"))
	assert.True(t, config.GetBool("APP_DEBUG"))
}
