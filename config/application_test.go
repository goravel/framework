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
FORGET_ME=forget-me-value
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
FORGET_ME=forget-me-value
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

	// A forgotten key decodes to zero values instead of the stored data.
	s.config.Add("forgettable", map[string]any{
		"host": "127.0.0.1",
		"port": 3306,
	})
	s.config.Forget("forgettable")

	var forgottenConfig MySQLConnection
	s.NoError(s.config.UnmarshalKey("forgettable", &forgottenConfig))
	s.Equal(MySQLConnection{}, forgottenConfig)

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

	s.config.Add("BAD_INT64", "abc")
	s.Equal(int64(0), s.config.GetInt64("BAD_INT64", 42))
}

func (s *ApplicationTestSuite) TestGetFloat64() {
	s.Equal(3.14, s.config.GetFloat64("FLOAT_VALUE"))
	s.Equal(0.0, s.config.GetFloat64("NOT_EXIST_FLOAT"))
	s.Equal(1.25, s.config.GetFloat64("NOT_EXIST_FLOAT", 1.25))

	s.config.Add("BAD_FLOAT", "abc")
	s.Equal(0.0, s.config.GetFloat64("BAD_FLOAT", 1.25))
}

func (s *ApplicationTestSuite) TestGetTime() {
	tests := []struct {
		name  string
		key   string
		value any
		want  time.Time
	}{
		{name: "time.Time value", key: "TIME_VALUE", value: time.Date(2023, 8, 15, 10, 30, 0, 0, time.UTC), want: time.Date(2023, 8, 15, 10, 30, 0, 0, time.UTC)},
		{name: "RFC3339", key: "TIME_RFC3339", value: "2023-08-15T10:30:00Z", want: time.Date(2023, 8, 15, 10, 30, 0, 0, time.UTC)},
		{name: "RFC3339 with fractional seconds", key: "TIME_FRACTIONAL", value: "2023-08-15T10:30:00.123456789Z", want: time.Date(2023, 8, 15, 10, 30, 0, 123456789, time.UTC)},
		{name: "date and time", key: "TIME_DATE_TIME", value: "2023-08-15 10:30:00", want: time.Date(2023, 8, 15, 10, 30, 0, 0, time.UTC)},
		{name: "date", key: "TIME_DATE", value: "2024-01-02", want: time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)},
		{name: "unparseable string", key: "TIME_INVALID", value: "not-a-time", want: time.Time{}},
		{name: "non string and non time", key: "TIME_INT", value: 1700000000, want: time.Time{}},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.config.Add(tt.key, tt.value)
			s.True(tt.want.Equal(s.config.GetTime(tt.key)))
		})
	}

	fallback := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	s.True(s.config.GetTime("NOT_EXIST_TIME").IsZero())
	s.True(fallback.Equal(s.config.GetTime("NOT_EXIST_TIME", fallback)))
	// Set but unparseable ignores the default, matching the other typed getters.
	s.True(s.config.GetTime("TIME_INVALID", fallback).IsZero())
}

func (s *ApplicationTestSuite) TestGetIntSlice() {
	s.config.Add("PORTS", []int{80, 443, 8080})
	s.Equal([]int{80, 443, 8080}, s.config.GetIntSlice("PORTS"))

	s.config.Add("EMPTY_PORTS", []int{})
	s.Equal([]int{}, s.config.GetIntSlice("EMPTY_PORTS"))

	s.config.Add("NIL_PORTS", []int(nil))
	s.Equal([]int{1}, s.config.GetIntSlice("NIL_PORTS", []int{1}))

	s.Nil(s.config.GetIntSlice("NOT_EXIST_INT_SLICE"))
	s.Equal([]int{1, 2, 3}, s.config.GetIntSlice("NOT_EXIST_INT_SLICE", []int{1, 2, 3}))
}

func (s *ApplicationTestSuite) TestGetStringSlice() {
	s.config.Add("NAMES", []string{"a", "b"})
	s.Equal([]string{"a", "b"}, s.config.GetStringSlice("NAMES"))

	// A stored empty slice is not replaced by the default.
	s.config.Add("EMPTY_NAMES", []string{})
	s.Equal([]string{}, s.config.GetStringSlice("EMPTY_NAMES", []string{"default"}))

	// A set but nil value falls back to the default, matching GetIntSlice.
	s.config.Add("NIL_NAMES", []string(nil))
	s.Equal([]string{"default"}, s.config.GetStringSlice("NIL_NAMES", []string{"default"}))

	s.Nil(s.config.GetStringSlice("NOT_EXIST_NAMES"))
	s.Equal([]string{"a"}, s.config.GetStringSlice("NOT_EXIST_NAMES", []string{"a"}))
}

func (s *ApplicationTestSuite) TestGetStringMap() {
	s.config.Add("FEATURES", map[string]any{
		"login":  true,
		"signup": true,
	})
	got := s.config.GetStringMap("FEATURES")
	s.Equal(true, got["login"])
	s.Equal(true, got["signup"])

	// A stored empty map is not replaced by the default.
	s.config.Add("EMPTY_FEATURES", map[string]any{})
	s.Empty(s.config.GetStringMap("EMPTY_FEATURES", map[string]any{"a": true}))

	s.Nil(s.config.GetStringMap("NOT_EXIST_MAP"))
	s.Equal(map[string]any{"a": "b"}, s.config.GetStringMap("NOT_EXIST_MAP", map[string]any{"a": "b"}))
}

func (s *ApplicationTestSuite) TestGetStringMapString() {
	s.config.Add("HEADERS", map[string]string{
		"X-Foo": "bar",
		"X-Baz": "qux",
	})
	s.Equal(map[string]string{"X-Foo": "bar", "X-Baz": "qux"}, s.config.GetStringMapString("HEADERS"))

	// A stored empty map is not replaced by the default.
	s.config.Add("EMPTY_HEADERS", map[string]string{})
	s.Empty(s.config.GetStringMapString("EMPTY_HEADERS", map[string]string{"k": "v"}))

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
	s.Equal("fallback", s.config.GetString("TEMP_VALUE", "fallback"))

	// A value loaded from the .env file is forgotten as well.
	s.True(s.config.Has("FORGET_ME"))
	s.config.Forget("FORGET_ME")
	s.False(s.config.Has("FORGET_ME"))
	s.Equal("", s.config.GetString("FORGET_ME"))
	s.Nil(s.config.Get("FORGET_ME"))
	s.NotContains(s.config.All(), "forget_me")

	// A value coming from the process environment is forgotten as well.
	s.T().Setenv("FORGET_ENV_VALUE", "from-env")
	s.Equal("from-env", s.config.GetString("FORGET_ENV_VALUE"))
	s.config.Forget("FORGET_ENV_VALUE")
	s.False(s.config.Has("FORGET_ENV_VALUE"))
	s.Equal("", s.config.GetString("FORGET_ENV_VALUE"))

	// Add makes the path visible again.
	s.config.Add("TEMP_VALUE", "hello again")
	s.True(s.config.Has("TEMP_VALUE"))
	s.Equal("hello again", s.config.GetString("TEMP_VALUE"))
}

func (s *ApplicationTestSuite) TestForgetNested() {
	s.config.Add("nested", map[string]any{
		"keep": "keep",
		"drop": map[string]any{
			"value": "drop",
		},
	})

	s.config.Forget("nested.drop")
	s.False(s.config.Has("nested.drop"))
	s.False(s.config.Has("nested.drop.value"))
	s.Equal("", s.config.GetString("nested.drop.value"))
	s.True(s.config.Has("nested.keep"))
	s.Equal("keep", s.config.GetString("nested.keep"))

	all := s.config.All()
	nested, ok := all["nested"].(map[string]any)
	s.True(ok)
	s.Equal(map[string]any{"keep": "keep"}, nested)

	// Forgetting a parent hides the whole subtree.
	s.config.Forget("nested")
	s.False(s.config.Has("nested.keep"))
	s.NotContains(s.config.All(), "nested")

	// Add on a parent re-enables the subtree.
	s.config.Add("nested", map[string]any{"keep": "keep again"})
	s.True(s.config.Has("nested.keep"))
	s.Equal("keep again", s.config.GetString("nested.keep"))
}

func (s *ApplicationTestSuite) TestForgetTypedMap() {
	s.config.Add("mail", map[string]any{
		"headers": map[string]string{
			"X-Foo": "bar",
			"X-Baz": "qux",
		},
	})

	s.config.Forget("mail.headers.X-Foo")
	s.False(s.config.Has("mail.headers.X-Foo"))
	s.Equal("", s.config.GetString("mail.headers.X-Foo"))

	mail, ok := s.config.All()["mail"].(map[string]any)
	s.True(ok)
	s.Equal(map[string]string{"X-Baz": "qux"}, mail["headers"])
}

func (s *ApplicationTestSuite) TestForgetKeepsEmptyParent() {
	s.config.Add("only_child", map[string]any{"value": "x"})

	s.config.Forget("only_child.value")
	s.False(s.config.Has("only_child.value"))
	// The parent stays in All, matching Has, instead of disappearing with its last child.
	s.True(s.config.Has("only_child"))
	s.Equal(map[string]any{}, s.config.All()["only_child"])
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

	// All does not share mutable state with the configuration.
	s.config.Add("all_paths", []string{"one"})
	s.config.Add("all_nested", map[string]any{"inner": map[string]any{"value": "origin"}})

	all = s.config.All()
	all["all_paths"].([]string)[0] = "mutated"
	all["all_nested"].(map[string]any)["inner"].(map[string]any)["value"] = "mutated"

	s.Equal([]string{"one"}, s.config.GetStringSlice("all_paths"))
	s.Equal("origin", s.config.GetString("all_nested.inner.value"))
}
