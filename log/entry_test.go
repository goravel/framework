package log

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"

	"github.com/goravel/framework/contracts/log"
)

type EntryTestSuite struct {
	suite.Suite
}

func TestEntryTestSuite(t *testing.T) {
	suite.Run(t, new(EntryTestSuite))
}

func (s *EntryTestSuite) TestAcquireAndReleaseEntry() {
	// Test acquiring an entry from the pool
	entry := acquireEntry()
	s.NotNil(entry)
	s.NotNil(entry.with)

	// Set some values
	entry.code = "test_code"
	entry.domain = "test_domain"
	entry.hint = "test_hint"
	entry.message = "test_message"
	entry.owner = "test_owner"
	entry.user = "test_user"
	entry.ctx = context.Background()
	entry.tags = []string{"tag1", "tag2"}
	entry.with["key"] = "value"

	// Release the entry
	releaseEntry(entry)

	// Verify the entry was reset
	s.Equal("", entry.code)
	s.Equal("", entry.domain)
	s.Equal("", entry.hint)
	s.Equal("", entry.message)
	s.Nil(entry.owner)
	s.Nil(entry.user)
	s.Nil(entry.ctx)
	s.Empty(entry.tags)
	s.Empty(entry.with)
}

func (s *EntryTestSuite) TestEntryGetters() {
	entry := &Entry{
		time:       time.Now(),
		ctx:        context.Background(),
		owner:      "owner",
		user:       "user",
		data:       log.Data{"key": "value"},
		request:    map[string]any{"method": "GET"},
		response:   map[string]any{"status": 200},
		stacktrace: map[string]any{"trace": "stack"},
		with:       map[string]any{"extra": "data"},
		code:       "ERR001",
		domain:     "test",
		hint:       "hint",
		message:    "message",
		tags:       []string{"tag1"},
		level:      log.LevelInfo,
	}

	s.Equal("ERR001", entry.Code())
	s.NotNil(entry.Context())
	s.Equal(log.Data{"key": "value"}, entry.Data())
	s.Equal("test", entry.Domain())
	s.Equal("hint", entry.Hint())
	s.Equal(log.LevelInfo, entry.Level())
	s.Equal("message", entry.Message())
	s.Equal("owner", entry.Owner())
	s.Equal(map[string]any{"method": "GET"}, entry.Request())
	s.Equal(map[string]any{"status": 200}, entry.Response())
	s.Equal([]string{"tag1"}, entry.Tags())
	s.NotZero(entry.Time())
	s.Equal(map[string]any{"trace": "stack"}, entry.Trace())
	s.Equal("user", entry.User())
	s.Equal(map[string]any{"extra": "data"}, entry.With())
}

func (s *EntryTestSuite) TestToSlogRecord() {
	now := time.Now()
	entry := &Entry{
		time:       now,
		ctx:        context.Background(),
		owner:      "owner",
		user:       "user",
		request:    map[string]any{"method": "GET"},
		response:   map[string]any{"status": 200},
		stacktrace: map[string]any{"trace": "stack"},
		with:       map[string]any{"extra": "data"},
		code:       "ERR001",
		domain:     "test",
		hint:       "hint",
		message:    "message",
		tags:       []string{"tag1"},
		level:      log.LevelInfo,
	}

	record := entry.ToSlogRecord()

	s.Equal(now, record.Time)
	s.Equal(slog.LevelInfo, record.Level)
	s.Equal("message", record.Message)

	// Count attributes
	attrCount := 0
	record.Attrs(func(a slog.Attr) bool {
		attrCount++
		return true
	})
	s.Greater(attrCount, 0)
}

func (s *EntryTestSuite) TestToSlogRecordWithEmptyFields() {
	entry := &Entry{
		time:    time.Now(),
		message: "test",
		level:   log.LevelDebug,
		with:    make(map[string]any),
	}

	record := entry.ToSlogRecord()

	s.Equal("test", record.Message)
	s.Equal(slog.LevelDebug, record.Level)
}

func (s *EntryTestSuite) TestFromSlogRecord() {
	now := time.Now()
	record := slog.NewRecord(now, slog.LevelInfo, "test message", 0)
	record.Add("code", "ERR001")
	record.Add("domain", "test")
	record.Add("hint", "hint")
	record.Add("owner", "owner")
	record.Add("user", "user")
	record.Add("tags", []string{"tag1", "tag2"})
	record.Add("request", map[string]any{"method": "GET"})
	record.Add("response", map[string]any{"status": 200})
	record.Add("stacktrace", map[string]any{"trace": "stack"})
	record.Add("with", map[string]any{"extra": "data"})
	record.Add("context", context.Background())

	entry := FromSlogRecord(record)

	s.Equal("ERR001", entry.Code())
	s.Equal("test", entry.Domain())
	s.Equal("hint", entry.Hint())
	s.Equal("test message", entry.Message())
	s.Equal("owner", entry.Owner())
	s.Equal("user", entry.User())
	s.Equal([]string{"tag1", "tag2"}, entry.Tags())
	s.Equal(map[string]any{"method": "GET"}, entry.Request())
	s.Equal(map[string]any{"status": 200}, entry.Response())
	s.Equal(map[string]any{"trace": "stack"}, entry.Trace())
	s.Equal(map[string]any{"extra": "data"}, entry.With())
	s.Equal(log.LevelInfo, entry.Level())
	s.Equal(now, entry.Time())

	// Cleanup
	releaseEntry(entry)
}

func (s *EntryTestSuite) TestFromSlogRecordWithUnknownAttributes() {
	now := time.Now()
	record := slog.NewRecord(now, slog.LevelInfo, "test message", 0)
	record.Add("unknown_key", "unknown_value")
	record.Add("another_key", 123)

	entry := FromSlogRecord(record)

	// Unknown attributes should be added to the 'with' map
	s.Equal("unknown_value", entry.With()["unknown_key"])
	s.Equal(int64(123), entry.With()["another_key"])

	// Cleanup
	releaseEntry(entry)
}

func (s *EntryTestSuite) TestFromSlogRecordWithMessage() {
	now := time.Now()
	record := slog.NewRecord(now, slog.LevelInfo, "original message", 0)
	record.Add("message", "overridden message")

	entry := FromSlogRecord(record)

	// The message attribute should override the record message
	s.Equal("overridden message", entry.Message())

	// Cleanup
	releaseEntry(entry)
}

func (s *EntryTestSuite) TestEntryPoolReuse() {
	// Get an entry, set values, release it
	entry1 := acquireEntry()
	entry1.code = "test"
	releaseEntry(entry1)

	// Get another entry - may be the same one from the pool
	entry2 := acquireEntry()

	// The entry should be reset
	s.Equal("", entry2.code)

	releaseEntry(entry2)
}
