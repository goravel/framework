package db

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/database/logger"
	mocksconfig "github.com/goravel/framework/mocks/config"
	mockslog "github.com/goravel/framework/mocks/log"
	"github.com/goravel/framework/support/carbon"
	"github.com/goravel/framework/support/debug"
)

func TestNewLogger(t *testing.T) {
	var (
		mockConfig *mocksconfig.Config
	)
	tests := []struct {
		name      string
		setup     func()
		wantLevel logger.Level
		wantSlow  time.Duration
	}{
		{
			name: "debug mode enabled",
			setup: func() {
				mockConfig.EXPECT().GetBool("app.debug").Return(true).Once()
				mockConfig.EXPECT().GetInt("database.slow_threshold", 200).Return(300).Once()
			},
			wantLevel: logger.Info,
			wantSlow:  300 * time.Millisecond,
		},
		{
			name: "debug mode disabled",
			setup: func() {
				mockConfig.EXPECT().GetBool("app.debug").Return(false).Once()
				mockConfig.EXPECT().GetInt("database.slow_threshold", 200).Return(300).Once()
			},
			wantLevel: logger.Warn,
			wantSlow:  300 * time.Millisecond,
		},
		{
			name: "negative slow threshold",
			setup: func() {
				mockConfig.EXPECT().GetBool("app.debug").Return(false).Once()
				mockConfig.EXPECT().GetInt("database.slow_threshold", 200).Return(0).Once()
			},
			wantLevel: logger.Warn,
			wantSlow:  200 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockConfig = mocksconfig.NewConfig(t)
			tt.setup()
			logger := NewLogger(mockConfig, nil)

			assert.Equal(t, tt.wantLevel, logger.(*Logger).level)
			assert.Equal(t, tt.wantSlow, logger.(*Logger).slowThreshold)
		})
	}
}

type LoggerTestSuite struct {
	suite.Suite
	mockLog *mockslog.Log
	logger  *Logger
}

func TestLoggerSuite(t *testing.T) {
	suite.Run(t, new(LoggerTestSuite))
}

func (s *LoggerTestSuite) SetupTest() {
	s.mockLog = mockslog.NewLog(s.T())
	s.logger = &Logger{
		log:           s.mockLog,
		level:         logger.Info,
		slowThreshold: 200 * time.Millisecond,
	}
}

func (s *LoggerTestSuite) TestLevel() {
	result := s.logger.Level(logger.Error)
	s.Equal(logger.Error, s.logger.level)
	s.Equal(s.logger, result)
}

func (s *LoggerTestSuite) TestInfo() {
	ctx := context.Background()

	s.mockLog.EXPECT().WithContext(ctx).Return(s.mockLog).Once()
	s.mockLog.EXPECT().Infof("test message", mock.Anything).Return().Once()

	s.logger.Infof(ctx, "test message")
}

func (s *LoggerTestSuite) TestWarn() {
	ctx := context.Background()

	s.mockLog.EXPECT().WithContext(ctx).Return(s.mockLog).Once()
	s.mockLog.EXPECT().Warningf("test warning", mock.Anything).Return().Once()

	s.logger.Warningf(ctx, "test warning")
}

func (s *LoggerTestSuite) TestError() {
	tests := []struct {
		name      string
		data      []any
		shouldLog bool
	}{
		{
			name:      "normal error",
			data:      []any{assert.AnError},
			shouldLog: true,
		},
		{
			name:      "connection refused error",
			data:      []any{&net.OpError{Err: errors.New("connection refused")}},
			shouldLog: false,
		},
		{
			name:      "access denied error",
			data:      []any{errors.New("Access denied for user")},
			shouldLog: false,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			ctx := context.Background()

			if tt.shouldLog {
				s.mockLog.EXPECT().WithContext(ctx).Return(s.mockLog).Once()
				s.mockLog.EXPECT().Errorf("test message", mock.Anything).Return().Once()
			}

			s.logger.Errorf(ctx, "test message", tt.data...)
		})
	}
}

func (s *LoggerTestSuite) TestTrace() {
	var (
		ctx  = context.Background()
		sql  = "SELECT * FROM users"
		rows = int64(1)
	)

	tests := []struct {
		name    string
		rows    int64
		elapsed time.Duration
		err     error
		level   logger.Level
		setup   func()
	}{
		{
			name:    "error case",
			rows:    1,
			elapsed: 100 * time.Millisecond,
			err:     assert.AnError,
			level:   logger.Error,
			setup: func() {
				s.mockLog.EXPECT().WithContext(ctx).Return(s.mockLog).Once()
				s.mockLog.EXPECT().Errorf("[%.3fms] [rows:%v] %s\t%s", mock.Anything, rows, sql, assert.AnError).Return().Once()
			},
		},
		{
			name:    "error case - rows -1",
			rows:    -1,
			elapsed: 100 * time.Millisecond,
			err:     assert.AnError,
			level:   logger.Error,
			setup: func() {
				s.mockLog.EXPECT().WithContext(ctx).Return(s.mockLog).Once()
				s.mockLog.EXPECT().Errorf("[%.3fms] [rows:%v] %s\t%s", mock.Anything, "-", sql, assert.AnError).Return().Once()
			},
		},
		{
			name:    "slow query",
			rows:    1,
			elapsed: 300 * time.Millisecond,
			level:   logger.Warn,
			setup: func() {
				s.mockLog.EXPECT().WithContext(ctx).Return(s.mockLog).Once()
				s.mockLog.EXPECT().Warningf("[%.3fms] [rows:%v] [SLOW] %s", mock.Anything, rows, sql).Return().Once()
			},
		},
		{
			name:    "slow query - rows -1",
			rows:    -1,
			elapsed: 300 * time.Millisecond,
			level:   logger.Warn,
			setup: func() {
				s.mockLog.EXPECT().WithContext(ctx).Return(s.mockLog).Once()
				s.mockLog.EXPECT().Warningf("[%.3fms] [rows:%v] [SLOW] %s", mock.Anything, "-", sql).Return().Once()
			},
		},
		{
			name:    "normal query",
			rows:    1,
			elapsed: 50 * time.Millisecond,
			level:   logger.Info,
			setup: func() {
				s.mockLog.EXPECT().WithContext(ctx).Return(s.mockLog).Once()
				s.mockLog.EXPECT().Infof("[%.3fms] [rows:%v] %s", mock.Anything, rows, sql).Return().Once()
			},
		},
		{
			name:    "normal query - rows -1",
			rows:    -1,
			elapsed: 50 * time.Millisecond,
			level:   logger.Info,
			setup: func() {
				s.mockLog.EXPECT().WithContext(ctx).Return(s.mockLog).Once()
				s.mockLog.EXPECT().Infof("[%.3fms] [rows:%v] %s", mock.Anything, "-", sql).Return().Once()
			},
		},
		{
			name:    "silent mode",
			elapsed: 50 * time.Millisecond,
			level:   logger.Silent,
			setup:   func() {},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			tt.setup()

			s.logger.level = tt.level
			begin := carbon.Now().SubDuration(tt.elapsed.String())
			s.logger.Trace(ctx, begin, sql, tt.rows, tt.err)
		})
	}
}

func TestAddToContext(t *testing.T) {
	ctx := context.Background()
	ctx = EnableQueryLog(ctx)
	addQueryLogToContext(ctx, "SELECT * FROM users", 100)
	addQueryLogToContext(ctx, "SELECT * FROM users", 200)

	value := ctx.Value(queryLogKey{})
	debug.Dump(value)
	assert.True(t, false)
}
