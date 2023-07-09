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
	config  *Application
	config2 *Application
}

func TestApplicationTestSuite(t *testing.T) {
	assert.Nil(t, file.Create(".env", "APP_KEY=12345678901234567890123456789012"))
	temp, err := os.CreateTemp("", "goravel.env")
	assert.Nil(t, err)
	defer os.Remove(temp.Name())
	_, err = temp.Write([]byte("APP_KEY=12345678901234567890123456789012"))
	assert.Nil(t, err)
	assert.Nil(t, temp.Close())

	suite.Run(t, &ApplicationTestSuite{
		config:  NewApplication(".env"),
		config2: NewApplication(temp.Name()),
	})

	assert.Nil(t, file.Remove(".env"))
}

func (s *ApplicationTestSuite) SetupTest() {

}

func (s *ApplicationTestSuite) TestEnv() {
	s.Equal("goravel", s.config.Env("APP_NAME", "goravel").(string))
	s.Equal("127.0.0.1", s.config.Env("DB_HOST", "127.0.0.1").(string))
	s.Equal("goravel", s.config2.Env("APP_NAME", "goravel").(string))
	s.Equal("127.0.0.1", s.config2.Env("DB_HOST", "127.0.0.1").(string))
}

func (s *ApplicationTestSuite) TestAdd() {
	s.config.Add("app", map[string]any{
		"env": "local",
	})
	s.config2.Add("app", map[string]any{
		"env": "local",
	})

	s.Equal("local", s.config.GetString("app.env"))
	s.Equal("local", s.config2.GetString("app.env"))

	s.config.Add("path.with.dot.case1", "value1")
	s.config2.Add("path.with.dot.case1", "value1")
	s.Equal("value1", s.config.GetString("path.with.dot.case1"))
	s.Equal("value1", s.config2.GetString("path.with.dot.case1"))

	s.config.Add("path.with.dot.case2", "value2")
	s.config2.Add("path.with.dot.case2", "value2")
	s.Equal("value2", s.config.GetString("path.with.dot.case2"))
	s.Equal("value2", s.config2.GetString("path.with.dot.case2"))

	s.config.Add("path.with.dot", map[string]any{"case3": "value3"})
	s.config2.Add("path.with.dot", map[string]any{"case3": "value3"})
	s.Equal("value3", s.config.GetString("path.with.dot.case3"))
	s.Equal("value3", s.config2.GetString("path.with.dot.case3"))
}

func (s *ApplicationTestSuite) TestGet() {
	s.Equal("goravel", s.config.Get("APP_NAME", "goravel").(string))
	s.Equal("goravel", s.config2.Get("APP_NAME", "goravel").(string))
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
	s.config2.Add("database", map[string]any{
		"default": s.config2.Env("DB_CONNECTION", "mysql"),
		"connections": map[string]any{
			"mysql": map[string]any{
				"host": s.config2.Env("DB_HOST", "127.0.0.1"),
			},
		},
	})

	s.Equal("goravel", s.config.GetString("APP_NAME", "goravel"))
	s.Equal("127.0.0.1", s.config.GetString("database.connections.mysql.host"))
	s.Equal("mysql", s.config.GetString("database.default"))
	s.Equal("goravel", s.config2.GetString("APP_NAME", "goravel"))
	s.Equal("127.0.0.1", s.config2.GetString("database.connections.mysql.host"))
	s.Equal("mysql", s.config2.GetString("database.default"))
}

func (s *ApplicationTestSuite) TestGetInt() {
	s.Equal(s.config.GetInt("DB_PORT", 3306), 3306)
	s.Equal(s.config2.GetInt("DB_PORT", 3306), 3306)
}

func (s *ApplicationTestSuite) TestGetBool() {
	s.Equal(true, s.config.GetBool("APP_DEBUG", true))
	s.Equal(true, s.config2.GetBool("APP_DEBUG", true))
}
