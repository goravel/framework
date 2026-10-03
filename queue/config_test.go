package queue

import (
	"encoding/json"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	mocksconfig "github.com/goravel/framework/mocks/config"
)

type ConfigTestSuite struct {
	suite.Suite
	mockConfig *mocksconfig.Config
	config     *Config
}

func TestConfigTestSuite(t *testing.T) {
	suite.Run(t, new(ConfigTestSuite))
}

func (s *ConfigTestSuite) SetupTest() {
	s.mockConfig = mocksconfig.NewConfig(s.T())
	s.mockConfig.EXPECT().GetString("queue.default").Return("redis").Once()
	s.mockConfig.EXPECT().GetString("queue.connections.redis.queue", "default").Return("default").Once()
	s.mockConfig.EXPECT().GetInt("queue.connections.redis.concurrent", 1).Return(2).Once()
	s.mockConfig.EXPECT().GetString("app.name", "goravel").Return("goravel").Once()
	s.mockConfig.EXPECT().GetBool("app.debug").Return(true).Once()
	s.mockConfig.EXPECT().GetString("queue.failed.database").Return("mysql").Once()
	s.mockConfig.EXPECT().GetString("queue.failed.table").Return("failed_jobs").Once()

	s.config = NewConfig(s.mockConfig)
}

func (s *ConfigTestSuite) TestDebug() {
	s.True(s.config.Debug())
}

func (s *ConfigTestSuite) TestDefaultConnection() {
	s.Equal("redis", s.config.DefaultConnection())
}

func (s *ConfigTestSuite) TestDefaultQueue() {
	s.Equal("default", s.config.DefaultQueue())
}

func (s *ConfigTestSuite) TestNewConfigNormalizesDefaultQueue() {
	tests := []struct {
		name            string
		configuredQueue string
		expect          string
	}{
		{
			name:            "queue list uses the first valid name",
			configuredQueue: " high, default ",
			expect:          "high",
		},
		{
			name:            "empty queue falls back to default",
			configuredQueue: "",
			expect:          "default",
		},
		{
			name:            "whitespace-only queue falls back to default",
			configuredQueue: " , ",
			expect:          "default",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			mockConfig := mocksconfig.NewConfig(s.T())
			mockConfig.EXPECT().GetString("queue.default").Return("redis").Once()
			mockConfig.EXPECT().GetString("queue.connections.redis.queue", "default").Return(tt.configuredQueue).Once()
			mockConfig.EXPECT().GetInt("queue.connections.redis.concurrent", 1).Return(1).Once()
			mockConfig.EXPECT().GetString("app.name", "goravel").Return("goravel").Once()
			mockConfig.EXPECT().GetBool("app.debug").Return(false).Once()
			mockConfig.EXPECT().GetString("queue.failed.database").Return("").Once()
			mockConfig.EXPECT().GetString("queue.failed.table").Return("").Once()

			config := NewConfig(mockConfig)

			s.Equal(tt.expect, config.DefaultQueue())
		})
	}
}

func (s *ConfigTestSuite) TestSplitQueueNames() {
	tests := []struct {
		name   string
		queue  string
		expect []string
	}{
		{
			name:   "single queue",
			queue:  "default",
			expect: []string{"default"},
		},
		{
			name:   "comma separated queues",
			queue:  "high,default",
			expect: []string{"high", "default"},
		},
		{
			name:   "trimmed and empty queue names",
			queue:  " high, , default,",
			expect: []string{"high", "default"},
		},
		{
			name:   "whitespace-only queue falls back to default",
			queue:  " , ",
			expect: []string{"default"},
		},
		{
			name:   "empty queue falls back to default",
			queue:  "",
			expect: []string{"default"},
		},
		{
			name:   "duplicate queues are preserved",
			queue:  "high,high,default",
			expect: []string{"high", "high", "default"},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			s.Equal(tt.expect, splitQueueNames(tt.queue))
		})
	}
}

func (s *ConfigTestSuite) TestDefaultConcurrent() {
	s.Equal(2, s.config.DefaultConcurrent())
}

func (s *ConfigTestSuite) TestDriver() {
	s.mockConfig.EXPECT().GetString("queue.connections.sync.driver").Return("sync").Once()
	s.Equal("sync", s.config.Driver("sync"))
}

func (s *ConfigTestSuite) TestFailedDatabase() {
	s.Equal("mysql", s.config.FailedDatabase())
}

func (s *ConfigTestSuite) TestFailedTable() {
	s.Equal("failed_jobs", s.config.FailedTable())
}

func (s *ConfigTestSuite) TestVia() {
	s.mockConfig.EXPECT().Get("queue.connections.sync.via").Return("sync").Once()
	s.Equal("sync", s.config.Via("sync"))
}

func (s *ConfigTestSuite) TestTimeout() {
	tests := []struct {
		name     string
		value    any
		expected time.Duration
	}{
		{name: "int seconds", value: 10, expected: 10 * time.Second},
		{name: "huge int seconds fall back to default", value: int64(math.MaxInt64), expected: defaultReceiveTimeout},
		{name: "int64 seconds", value: int64(8), expected: 8 * time.Second},
		{name: "uint32 seconds", value: uint32(6), expected: 6 * time.Second},
		{name: "zero uint falls back to default", value: uint64(0), expected: defaultReceiveTimeout},
		{name: "huge uint seconds fall back to default", value: uint64(math.MaxUint64), expected: defaultReceiveTimeout},
		{name: "float64 seconds", value: float64(7), expected: 7 * time.Second},
		{name: "float32 seconds", value: float32(4), expected: 4 * time.Second},
		{name: "fractional float keeps precision", value: 2.9, expected: 2900 * time.Millisecond},
		{name: "NaN float falls back to default", value: math.NaN(), expected: defaultReceiveTimeout},
		{name: "positive infinity falls back to default", value: math.Inf(1), expected: defaultReceiveTimeout},
		{name: "overflowing float falls back to default", value: 1e18, expected: defaultReceiveTimeout},
		{name: "json.Number seconds", value: json.Number("3.5"), expected: 3500 * time.Millisecond},
		{name: "invalid json.Number falls back to default", value: json.Number("not-a-number"), expected: defaultReceiveTimeout},
		{name: "time.Duration value", value: 3 * time.Second, expected: 3 * time.Second},
		{name: "duration string", value: "5s", expected: 5 * time.Second},
		{name: "duration string with ms", value: "1500ms", expected: 1500 * time.Millisecond},
		{name: "numeric string seconds", value: "10", expected: 10 * time.Second},
		{name: "fractional numeric string seconds", value: "2.5", expected: 2500 * time.Millisecond},
		{name: "zero duration string falls back to default", value: "0s", expected: defaultReceiveTimeout},
		{name: "zero falls back to default", value: 0, expected: defaultReceiveTimeout},
		{name: "negative falls back to default", value: -3, expected: defaultReceiveTimeout},
		{name: "negative duration string falls back to default", value: "-5s", expected: defaultReceiveTimeout},
		{name: "garbage string falls back to default", value: "not-a-duration", expected: defaultReceiveTimeout},
		{name: "missing falls back to default", value: nil, expected: 5 * time.Second},
		{name: "unsupported type falls back to default", value: true, expected: defaultReceiveTimeout},
	}

	for _, test := range tests {
		s.Run(test.name, func() {
			s.mockConfig.EXPECT().Get("queue.connections.redis.timeout").Return(test.value).Once()
			s.Equal(test.expected, s.config.Timeout("redis"))
		})
	}
}
