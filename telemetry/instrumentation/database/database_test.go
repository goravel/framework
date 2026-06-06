package database

import (
	"database/sql"
	"database/sql/driver"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
	"gorm.io/gorm"
)

func TestDBSystem(t *testing.T) {
	tests := []struct {
		driver   string
		expected string
	}{
		{"postgres", "postgresql"},
		{"mysql", "mysql"},
		{"sqlite", "sqlite"},
		{"sqlserver", "microsoft.sql_server"},
		{"alien", "alien"},
	}

	for _, tt := range tests {
		t.Run(tt.driver, func(t *testing.T) {
			assert.Equal(t, tt.expected, dbSystem(tt.driver).Value.AsString())
			assert.Equal(t, semconv.DBSystemNameKey, dbSystem(tt.driver).Key)
		})
	}
}

func TestOperationName(t *testing.T) {
	tests := []struct {
		query    string
		expected string
	}{
		{"SELECT * FROM users", "SELECT"},
		{"  insert into users values (?)", "INSERT"},
		{"UPDATE users SET name = ?", "UPDATE"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, operationName(tt.query))
		})
	}
}

func TestIsRecordableError(t *testing.T) {
	assert.False(t, isRecordableError(nil))
	assert.False(t, isRecordableError(gorm.ErrRecordNotFound))
	assert.False(t, isRecordableError(sql.ErrNoRows))
	assert.False(t, isRecordableError(driver.ErrSkip))
	assert.False(t, isRecordableError(io.EOF))
	assert.True(t, isRecordableError(assert.AnError))
}
