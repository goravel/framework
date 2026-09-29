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
	s.T().Setenv("APP_KEY", "12345678901234567890123456789013")
	s.T().Setenv("OS_APP_NAME", "goravel")
	s.T().Setenv("OS_APP_PORT", "3306")
	s.T().Setenv("OS_APP_DEBUG", "true")
	s.T().Setenv("OS_TIMEOUT", "5s")

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

func (s *ApplicationTestSuite) TestEnvStringFunction() {
	// existing value
	s.T().Setenv("ENVSTRING_VAR", "hello")
	s.Equal("hello", s.config.EnvString("ENVSTRING_VAR"))
	s.Equal("hello", s.customConfig.EnvString("ENVSTRING_VAR"))

	// default used when not set
	s.Equal("fallback", s.config.EnvString("ENVSTRING_NOT_SET", "fallback"))

	// empty string -> use provided default, otherwise empty string
	s.T().Setenv("ENVSTRING_EMPTY", "")
	s.Equal("fallback", s.config.EnvString("ENVSTRING_EMPTY", "fallback"))
	s.Equal("", s.config.EnvString("ENVSTRING_EMPTY"))
}

func (s *ApplicationTestSuite) TestEnvBoolFunction() {
	// true/false values
	s.T().Setenv("ENVBOOL_TRUE", "true")
	s.True(s.config.EnvBool("ENVBOOL_TRUE"))
	s.T().Setenv("ENVBOOL_FALSE", "false")
	s.False(s.config.EnvBool("ENVBOOL_FALSE"))

	// not set -> default respected
	s.True(s.config.EnvBool("ENVBOOL_NOT_SET", true))
	s.False(s.config.EnvBool("ENVBOOL_NOT_SET2", false))

	// empty string -> use default if provided; otherwise cast to false
	s.T().Setenv("ENVBOOL_EMPTY", "")
	s.True(s.config.EnvBool("ENVBOOL_EMPTY", true))
	s.False(s.config.EnvBool("ENVBOOL_EMPTY"))

	// invalid -> false
	s.T().Setenv("ENVBOOL_INVALID", "invalid")
	s.False(s.config.EnvBool("ENVBOOL_INVALID"))
}

func (s *ApplicationTestSuite) TestUnmarshalKey() {
	s.config.Add("database", map[string]any{
		"default": "mysql",
		"connections": map[string]any{
			"mysql": map[string]any{
				"host":     "127.0.0.1",
				"port":     3306,
				"database": "goravel",
				"username": "root",
				"password": "secret",
			},
		},
	})

	type DatabaseConfig struct {
		Default     string         `json:"default"`
		Connections map[string]any `json:"connections"`
	}

	var dbConfig DatabaseConfig
	err := s.config.UnmarshalKey("database", &dbConfig)
	s.NoError(err)
	s.Equal("mysql", dbConfig.Default)
	s.NotNil(dbConfig.Connections["mysql"])

	type MySQLConnection struct {
		Host     string
		Port     int
		Database string
		Username string
		Password string
	}

	var mysqlConfig MySQLConnection
	err = s.config.UnmarshalKey("database.connections.mysql", &mysqlConfig)
	s.NoError(err)
	s.Equal("127.0.0.1", mysqlConfig.Host)
	s.Equal(3306, mysqlConfig.Port)
	s.Equal("goravel", mysqlConfig.Database)
	s.Equal("root", mysqlConfig.Username)
	s.Equal("secret", mysqlConfig.Password)

	var emptyConfig MySQLConnection
	err = s.config.UnmarshalKey("non_existent_key", &emptyConfig)
	s.NoError(err)
	s.Equal("", emptyConfig.Host)
	s.Equal(0, emptyConfig.Port)

	// Test with custom config for customConfig instance
	s.customConfig.Add("app", map[string]any{
		"name":  "Goravel",
		"debug": true,
		"port":  8080,
	})

	type AppConfig struct {
		Name  string `json:"name"`
		Debug bool   `json:"debug"`
		Port  int    `json:"port"`
	}

	var appConfig AppConfig
	err = s.customConfig.UnmarshalKey("app", &appConfig)
	s.NoError(err)
	s.Equal("Goravel", appConfig.Name)
	s.True(appConfig.Debug)
	s.Equal(8080, appConfig.Port)
}

func TestOsVariables(t *testing.T) {
	t.Setenv("APP_KEY", "12345678901234567890123456789013")
	t.Setenv("APP_NAME", "goravel")
	t.Setenv("APP_PORT", "3306")
	t.Setenv("APP_DEBUG", "true")

	config := NewApplication(support.EnvFilePath)

	assert.Equal(t, "12345678901234567890123456789013", config.GetString("APP_KEY"))
	assert.Equal(t, "goravel", config.GetString("APP_NAME"))
	assert.Equal(t, 3306, config.GetInt("APP_PORT"))
	assert.True(t, config.GetBool("APP_DEBUG"))
}

func (s *ApplicationTestSuite) TestGetInt64() {
	s.config.Add("BIG_PORT", "9000000000")
	s.Equal(int64(9000000000), s.config.GetInt64("BIG_PORT"))

	s.Equal(int64(0), s.config.GetInt64("NOT_EXIST_INT64"))
	s.Equal(int64(42), s.config.GetInt64("NOT_EXIST_INT64", 42))
}

func (s *ApplicationTestSuite) TestGetFloat64() {
	s.Equal(3.14, s.config.GetFloat64("FLOAT_VALUE"))
	s.Equal(0.0, s.config.GetFloat64("NOT_EXIST_FLOAT"))
	s.Equal(1.25, s.config.GetFloat64("NOT_EXIST_FLOAT", 1.25))
}

func (s *ApplicationTestSuite) TestGetTime() {
	s.config.Add("BIRTHDAY", "2023-08-15T10:30:00Z")
	expected, err := time.Parse(time.RFC3339, "2023-08-15T10:30:00Z")
	s.NoError(err)
	s.True(expected.Equal(s.config.GetTime("BIRTHDAY")))

	s.config.Add("SHORT_DATE", "2024-01-02")
	expected2, err := time.Parse("2006-01-02", "2024-01-02")
	s.NoError(err)
	s.True(expected2.Equal(s.config.GetTime("SHORT_DATE")))

	fallback := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	s.Equal(time.Time{}, s.config.GetTime("NOT_EXIST_TIME"))
	s.True(fallback.Equal(s.config.GetTime("NOT_EXIST_TIME", fallback)))

	s.config.Add("INVALID_TIME", "not-a-time")
	s.Equal(time.Time{}, s.config.GetTime("INVALID_TIME"))
	s.True(fallback.Equal(s.config.GetTime("INVALID_TIME", fallback)))
}

func (s *ApplicationTestSuite) TestGetIntSlice() {
	s.config.Add("PORTS", []int{80, 443, 8080})
	s.Equal([]int{80, 443, 8080}, s.config.GetIntSlice("PORTS"))

	s.config.Add("EMPTY_PORTS", []int{})
	s.Equal([]int{}, s.config.GetIntSlice("EMPTY_PORTS"))

	s.Nil(s.config.GetIntSlice("NOT_EXIST_INT_SLICE"))
	s.Equal([]int{1, 2, 3}, s.config.GetIntSlice("NOT_EXIST_INT_SLICE", []int{1, 2, 3}))
}

func (s *ApplicationTestSuite) TestGetStringMap() {
	s.config.Add("FEATURES", map[string]any{
		"login":  true,
		"signup": true,
	})
	got := s.config.GetStringMap("FEATURES")
	s.Equal(true, got["login"])
	s.Equal(true, got["signup"])

	s.Nil(s.config.GetStringMap("NOT_EXIST_MAP"))
	s.Equal(map[string]any{"a": "b"}, s.config.GetStringMap("NOT_EXIST_MAP", map[string]any{"a": "b"}))
}

func (s *ApplicationTestSuite) TestGetStringMapString() {
	s.config.Add("HEADERS", map[string]string{
		"X-Foo": "bar",
		"X-Baz": "qux",
	})
	s.Equal(map[string]string{"X-Foo": "bar", "X-Baz": "qux"}, s.config.GetStringMapString("HEADERS"))

	s.Nil(s.config.GetStringMapString("NOT_EXIST_MAP_STRING"))
	s.Equal(map[string]string{"k": "v"}, s.config.GetStringMapString("NOT_EXIST_MAP_STRING", map[string]string{"k": "v"}))
}

func (s *ApplicationTestSuite) TestHas() {
	s.True(s.config.Has("APP_KEY"))
	s.True(s.config.Has("APP_DEBUG"))
	s.False(s.config.Has("DEFINITELY_NOT_SET_12345"))
}

func (s *ApplicationTestSuite) TestForget() {
	s.config.Add("TEMP_VALUE", "hello")
	s.True(s.config.Has("TEMP_VALUE"))
	s.Equal("hello", s.config.GetString("TEMP_VALUE"))

	s.config.Forget("TEMP_VALUE")
	s.False(s.config.Has("TEMP_VALUE"))
	s.Equal("", s.config.GetString("TEMP_VALUE"))
}

func (s *ApplicationTestSuite) TestAll() {
	s.config.Add("ALL_TEST_KEY", "value")
	s.config.Add("ALL_TEST_BOOL", true)
	s.config.Add("ALL_TEST_INT", 42)

	all := s.config.All()
	s.NotNil(all)
	s.Equal("value", all["all_test_key"])
	s.Equal(true, all["all_test_bool"])
	s.Equal(42, all["all_test_int"])
}
